package avc

import (
	"bytes"
	"bzcom/media/container/flv"
	"errors"
	"io"
)

const (
	i_frame byte = 0
	p_frame byte = 1
	b_frame byte = 2
)

const (
	nalu_type_not_define byte = 0
	nalu_type_slice      byte = 1  //slice_layer_without_partioning_rbsp() sliceheader
	nalu_type_dpa        byte = 2  // slice_data_partition_a_layer_rbsp( ), slice_header
	nalu_type_dpb        byte = 3  // slice_data_partition_b_layer_rbsp( )
	nalu_type_dpc        byte = 4  // slice_data_partition_c_layer_rbsp( )
	nalu_type_idr        byte = 5  // slice_layer_without_partitioning_rbsp( ),sliceheader
	nalu_type_sei        byte = 6  //sei_rbsp( )
	nalu_type_sps        byte = 7  //seq_parameter_set_rbsp( )
	nalu_type_pps        byte = 8  //pic_parameter_set_rbsp( )
	nalu_type_aud        byte = 9  // access_unit_delimiter_rbsp( )
	nalu_type_eoesq      byte = 10 //end_of_seq_rbsp( )
	nalu_type_eostream   byte = 11 //end_of_stream_rbsp( )
	nalu_type_filler     byte = 12 //filler_data_rbsp( )
)

const (
	naluBytesLen int = 4
	maxSpsPpsLen int = 2 * 1024
)

var (
	decDataNil        = errors.New("dec buf is nil")
	spsDataError      = errors.New("sps data error")
	ppsHeaderError    = errors.New("pps header error")
	ppsDataError      = errors.New("pps data error")
	naluHeaderInvalid = errors.New("nalu header invalid")
	videoDataInvalid  = errors.New("video data not match")
	dataSizeNotMatch  = errors.New("data size not match")
	naluBodyLenError  = errors.New("nalu body len error")
)

var naluIndication = []byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xf0}

type Demuxer struct {
	frameType    byte
	specificInfo []byte
	pps          *bytes.Buffer
}

type sequenceHeader struct {
	configVersion        byte //8bits
	avcProfileIndication byte //8bits
	profileCompatility   byte //8bits
	avcLevelIndication   byte //8bits
	reserved1            byte //6bits
	naluLen              byte //2bits
	reserved2            byte //3bits
	spsNum               byte //5bits
	ppsNum               byte //8bits
	spsLen               int
	ppsLen               int
}

func NewDemuxer() *Demuxer {
	return &Demuxer{
		pps: bytes.NewBuffer(make([]byte, maxSpsPpsLen)),
	}
}

//return value 1:sps, value2 :pps
func (self *Demuxer) parseSpecificInfo(src []byte) error {
	if len(src) < 9 {
		return decDataNil
	}
	sps := []byte{}
	pps := []byte{}
	header := []byte{0x00, 0x00, 0x00, 0x01}
	var seq sequenceHeader
	seq.configVersion = src[0]
	seq.avcProfileIndication = src[1]
	seq.profileCompatility = src[2]
	seq.avcLevelIndication = src[3]
	seq.reserved1 = src[4] & 0xfc
	seq.naluLen = src[4]&0x03 + 1
	seq.reserved2 = src[5] >> 5

	//get sps
	seq.spsNum = src[5] & 0x1f
	seq.spsLen = int(src[6])<<8 | int(src[7])

	if len(src[8:]) < seq.spsLen || seq.spsLen <= 0 {
		return spsDataError
	}
	sps = append(sps, header...)
	sps = append(sps, src[8:(8+seq.spsLen)]...)

	//get pps
	tmpBuf := src[(8 + seq.spsLen):]
	if len(tmpBuf) < 4 {
		return ppsHeaderError
	}
	seq.ppsNum = tmpBuf[0]
	seq.ppsLen = int(0)<<16 | int(tmpBuf[1])<<8 | int(tmpBuf[2])
	if len(tmpBuf[3:]) < seq.ppsLen || seq.ppsLen <= 0 {
		return ppsDataError
	}

	pps = append(pps, header...)
	pps = append(pps, tmpBuf[3:]...)

	self.specificInfo = append(self.specificInfo, sps...)
	self.specificInfo = append(self.specificInfo, pps...)

	return nil
}

func (self *Demuxer) isNaluHeader(src []byte) bool {
	if len(src) < naluBytesLen {
		return false
	}
	return src[0] == 0x00 &&
		src[1] == 0x00 &&
		src[2] == 0x00 &&
		src[3] == 0x01
}

func (self *Demuxer) setNaluHeader(src []byte) error {
	if len(src) != naluBytesLen {
		return naluHeaderInvalid
	}
	src[0] = 0x00
	src[1] = 0x00
	src[2] = 0x00
	src[3] = 0x01
	return nil
}

func (self *Demuxer) naluSize(src []byte) (int, error) {
	if len(src) < naluBytesLen {
		return 0, errors.New("nalusizedata invalid")
	}
	size := int(0)
	for i := 0; i < len(src); i++ {
		size = size<<8 + int(src[i])
	}
	return size, nil
}

func (self *Demuxer) getAnnexbH264(src []byte, w io.Writer) error {
	dataSize := len(src)
	if dataSize < naluBytesLen {
		return videoDataInvalid
	}
	self.pps.Reset()
	_, err := w.Write(naluIndication)
	if err != nil {
		return err
	}

	index := 0
	nalLen := 0
	hasSpsPps := false
	hasWriteSpsPps := false

	for dataSize > 0 {
		nalLen, err = self.naluSize(src[index:])
		if err != nil {
			return dataSizeNotMatch
		}
		index += naluBytesLen
		dataSize -= naluBytesLen

		if dataSize >= nalLen && len(src[index:]) >= nalLen && nalLen > 0 {
			self.setNaluHeader(src[index-naluBytesLen : index])
			nalType := src[index] & 0x1f
			switch nalType {
			case nalu_type_aud:
			case nalu_type_idr:
				if !hasWriteSpsPps {
					hasWriteSpsPps = true
					if !hasSpsPps {
						if _, err := w.Write(self.specificInfo); err != nil {
							return err
						}
					} else {
						if _, err := w.Write(self.pps.Bytes()); err != nil {
							return err
						}
					}
				}
				fallthrough
			case nalu_type_slice:
				fallthrough
			case nalu_type_sei:
				_, err := w.Write(src[index-naluBytesLen : index+nalLen])
				if err != nil {
					return err
				}
			case nalu_type_sps:
				fallthrough
			case nalu_type_pps:
				hasSpsPps = true
				_, err := self.pps.Write(src[index-naluBytesLen : index+nalLen])
				if err != nil {
					return err
				}
			}
			index += nalLen
			dataSize -= nalLen
		} else {
			return naluBodyLenError
		}
	}
	return nil
}

func (self *Demuxer) SampleRate() int {
	return 0
}

func (self *Demuxer) Demux(tag *flv.Tag, w io.Writer) (err error) {
	switch tag.MT.AVCPacketType {
	case flv.AVC_SEQHDR:
		err = self.parseSpecificInfo(tag.Data)
	case flv.AVC_NALU:
		if _, err = self.naluSize(tag.Data); err != nil {
			return err
		}
		// is annexb
		if self.isNaluHeader(tag.Data[naluBytesLen:]) {
			_, err = w.Write(tag.Data[naluBytesLen:])
		} else {
			err = self.getAnnexbH264(tag.Data, w)
		}
	}
	return
}
