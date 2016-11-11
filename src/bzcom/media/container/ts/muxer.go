package ts

import (
	"bzcom/media/container/flv"
	"io"
)

const (
	tsDefaultDataLen = 184
	tsPacketLen      = 188
	H264DefaultHZ    = 90
)

type Muxer struct {
	videoCc  byte
	audioCc  byte
	patCc    byte
	pmtCc    byte
	pes      *pesHeader
	tsPacket [tsPacketLen]byte
	pat      [tsPacketLen]byte
	pmt      [tsPacketLen]byte
}

type Frame struct {
	dts         uint64
	pts         uint64
	pid         uint32
	sid         uint32
	isVideo     bool
	isKey       bool
	soundFormat byte
	data        []byte
}

// Init info  based on params
func (self *Frame) Make(b []byte, isKey, isVideo bool, pts, dts uint64, soundFormat byte) {
	if isVideo == true {
		self.pid = 0x100
		self.sid = 0xe0
	} else {
		self.pid = 0x101
		self.sid = 0xc0
	}
	self.data = b
	self.pts = pts
	self.dts = dts
	self.isKey = isKey
	self.isVideo = isVideo
	self.soundFormat = soundFormat
}

func NewMuxer() *Muxer {
	return &Muxer{
		pes: &pesHeader{},
	}
}

func (self *Muxer) Mux(frame *Frame, w io.Writer) error {
	i := byte(0)
	first := byte(1)
	dataLen := byte(0)
	writedBytes := 0
	pesHeaderIndex := 0
	tmpLen := byte(0)
	//pes header packet
	var pes pesHeader
	err := pes.packet(frame)
	if err != nil {
		return err
	}
	pesHeaderLen := pes.len
	pesPacketBytes := len(frame.data) + int(pesHeaderLen)
	for {
		if pesPacketBytes <= 0 {
			break
		}
		if frame.isVideo {
			self.videoCc++
			if self.videoCc > 0xf {
				self.videoCc = 0
			}
		} else {
			self.audioCc++
			if self.audioCc > 0xf {
				self.audioCc = 0
			}
		}
		i = 0
		//sync byte
		self.tsPacket[i] = 0x47
		i++
		//error indicator, unit start indicator,ts priority,pid
		self.tsPacket[i] = byte(frame.pid >> 8) //pid high 5 bits
		if first == 1 {
			self.tsPacket[i] = self.tsPacket[i] | 0x40 //unit start indicator
		}
		i++
		//pid low 8 bits
		self.tsPacket[i] = byte(frame.pid)
		i++

		//scram control, adaptation control, counter
		if frame.isVideo {
			self.tsPacket[i] = 0x10 | byte(self.videoCc&0x0f)
		} else {
			self.tsPacket[i] = 0x10 | byte(self.audioCc&0x0f)
		}
		i++

		if first == 1 {
			//关键帧需要加pcr
			if frame.isKey {
				self.tsPacket[3] |= 0x20
				self.tsPacket[i] = 7
				i++
				self.tsPacket[i] = 0x50
				i++
				self.writePcr(self.tsPacket[0:], i, frame.dts)
				i += 6
			}
		}
		//frame data
		if pesPacketBytes >= tsDefaultDataLen {
			dataLen = tsDefaultDataLen
			if first == 1 {
				dataLen -= (i - 4)
			}
		} else {
			self.tsPacket[3] |= 0x20 //have adaptation
			remainBytes := byte(0)
			dataLen = byte(pesPacketBytes)
			if first == 1 {
				remainBytes = tsDefaultDataLen - dataLen - (i - 4)
			} else {
				remainBytes = tsDefaultDataLen - dataLen
			}
			self.adaptationBufInit(self.tsPacket[i:], byte(remainBytes))
			i += remainBytes
		}
		if first == 1 && i < tsPacketLen && pesHeaderLen > 0 {
			tmpLen = tsPacketLen - i
			if pesHeaderLen <= tmpLen {
				tmpLen = pesHeaderLen
			}
			copy(self.tsPacket[i:], pes.data[pesHeaderIndex:pesHeaderIndex+int(tmpLen)])
			i += tmpLen
			pesPacketBytes -= int(tmpLen)
			dataLen -= tmpLen
			pesHeaderLen -= tmpLen
			pesHeaderIndex += int(tmpLen)
		}

		if i < tsPacketLen {
			tmpLen = tsPacketLen - i
			if tmpLen <= dataLen {
				dataLen = tmpLen
			}
			copy(self.tsPacket[i:], frame.data[writedBytes:writedBytes+int(dataLen)])
			writedBytes += int(dataLen)
			pesPacketBytes -= int(dataLen)
		}
		if w != nil {
			if _, err := w.Write(self.tsPacket[0:]); err != nil {
				return err
			}
		}
		first = 0
	}

	return nil
}

//PAT return pat data
func (self *Muxer) PAT() []byte {

	i := int(0)
	remainByte := int(0)
	tsHeader := []byte{0x47, 0x40, 0x00, 0x10, 0x00}
	patHeader := []byte{0x00, 0xb0, 0x0d, 0x00, 0x01, 0xc1, 0x00, 0x00, 0x00, 0x01, 0xf0, 0x01}

	if self.patCc > 0xf {
		self.patCc = 0
	}
	tsHeader[3] |= self.patCc & 0x0f
	self.patCc++

	copy(self.pat[i:], tsHeader)
	i += len(tsHeader)

	copy(self.pat[i:], patHeader)
	i += len(patHeader)

	crc32Value := GenCrc32(patHeader)
	self.pat[i] = byte(crc32Value >> 24)
	i++
	self.pat[i] = byte(crc32Value >> 16)
	i++
	self.pat[i] = byte(crc32Value >> 8)
	i++
	self.pat[i] = byte(crc32Value)
	i++

	remainByte = int(tsPacketLen - i)
	for j := 0; j < remainByte; j++ {
		self.pat[i+j] = 0xff
	}

	return self.pat[0:]
}

// PMT return pmt data
func (self *Muxer) PMT(soundFormat byte, hasVideo bool) []byte {
	i := int(0)
	j := int(0)
	var progInfo []byte
	remainBytes := int(0)
	tsHeader := []byte{0x47, 0x50, 0x01, 0x10, 0x00}
	pmtHeader := []byte{0x02, 0xb0, 0xff, 0x00, 0x01, 0xc1, 0x00, 0x00, 0xe1, 0x00, 0xf0, 0x00}
	if !hasVideo {
		pmtHeader[9] = 0x01
		progInfo = []byte{0x0f, 0xe1, 0x01, 0xf0, 0x00}
	} else {
		progInfo = []byte{0x1b, 0xe1, 0x00, 0xf0, 0x00, //h264 or h265*
			0x0f, 0xe1, 0x01, 0xf0, 0x00, //mp3 or aac
		}
	}
	pmtHeader[2] = byte(len(progInfo) + 9 + 4)

	if self.pmtCc > 0xf {
		self.pmtCc = 0
	}
	tsHeader[3] |= self.pmtCc & 0x0f
	self.pmtCc++

	if soundFormat == flv.SOUND_AAC ||
		soundFormat == flv.SOUND_MP3 {
		if hasVideo {
			progInfo[5] = 0x4
		} else {
			progInfo[0] = 0x4
		}
	}

	copy(self.pmt[i:], tsHeader)
	i += len(tsHeader)

	copy(self.pmt[i:], pmtHeader)
	i += len(pmtHeader)

	copy(self.pmt[i:], progInfo[0:])
	i += len(progInfo)

	crc32Value := GenCrc32(self.pmt[5 : 5+len(pmtHeader)+len(progInfo)])
	self.pmt[i] = byte(crc32Value >> 24)
	i++
	self.pmt[i] = byte(crc32Value >> 16)
	i++
	self.pmt[i] = byte(crc32Value >> 8)
	i++
	self.pmt[i] = byte(crc32Value)
	i++

	remainBytes = int(tsPacketLen - i)
	for j = 0; j < remainBytes; j++ {
		self.pmt[i+j] = 0xff
	}

	return self.pmt[0:]
}

func (self *Muxer) adaptationBufInit(src []byte, remainBytes byte) {
	src[0] = byte(remainBytes - 1)
	if remainBytes == 1 {
	} else {
		src[1] = 0x00
		for i := 2; i < len(src); i++ {
			src[i] = 0xff
		}
	}
	return
}

func (self *Muxer) writePcr(b []byte, i byte, pcr uint64) error {
	b[i] = byte(pcr >> 25)
	i++
	b[i] = byte((pcr >> 17) & 0xff)
	i++
	b[i] = byte((pcr >> 9) & 0xff)
	i++
	b[i] = byte((pcr >> 1) & 0xff)
	i++
	b[i] = byte(((pcr & 0x1) << 7) | 0x7e)
	i++
	b[i] = 0x00

	return nil
}

type pesHeader struct {
	len  byte
	data [tsPacketLen]byte
}

//pesPacket return pes packet
func (self *pesHeader) packet(frame *Frame) error {
	i := 0
	flag := 0x80
	pesSize := 0
	headerSize := 5

	if frame.isVideo {
		if frame.pts != frame.dts {
			headerSize += 5 //add dts
			flag |= 0x40
		}
	}

	//PES header
	self.data[i] = 0x00
	i++
	self.data[i] = 0x00
	i++
	self.data[i] = 0x01
	i++
	self.data[i] = byte(frame.sid)
	i++

	pesSize = len(frame.data) + headerSize + 3
	if pesSize > 0xffff {
		pesSize = 0
	}
	self.data[i] = byte(pesSize >> 8)
	i++
	self.data[i] = byte(pesSize)
	i++

	self.data[i] = 0x80
	i++
	self.data[i] = byte(flag)
	i++
	self.data[i] = byte(headerSize)
	i++

	self.writePts(self.data[0:], i, flag>>6, frame.pts)
	i += 5
	if frame.isVideo && frame.pts != frame.dts {
		self.writePts(self.data[0:], i, 1, frame.dts)
		i += 5
	}

	self.len = byte(i)

	return nil
}

func (self *pesHeader) writePts(src []byte, i int, fb int, pts uint64) {
	val := uint32(0)

	if pts > 0x1ffffffff {
		pts -= 0x1ffffffff
	}

	val = uint32(fb<<4) | ((uint32(pts>>30) & 0x07) << 1) | 1
	src[i] = byte(val)
	i++

	val = ((uint32(pts>>15) & 0x7fff) << 1) | 1
	src[i] = byte(val >> 8)
	i++
	src[i] = byte(val)
	i++

	val = (uint32(pts&0x7fff) << 1) | 1
	src[i] = byte(val >> 8)
	i++
	src[i] = byte(val)
}
