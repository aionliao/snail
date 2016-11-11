package flv

import "fmt"

type FLVTag struct {
	Type      uint8
	DataSize  uint32
	TimeStamp uint32
	StreamID  uint32 // always 0
}

type MediaTag struct {
	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	SoundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	SoundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	SoundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	SoundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	AACPacketType uint8

	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non- seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	FrameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	CodecID uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	AVCPacketType uint8

	CompositionTime int32
}

type Tag struct {
	FT   FLVTag
	MT   MediaTag
	Data []byte
}

func (self *Tag) parseAudioHeader(b []byte) (n int, err error) {
	if len(b) < n+1 {
		err = fmt.Errorf("invalid audiodata len=%d", len(b))
		return
	}
	flags := b[0]
	self.MT.SoundFormat = flags >> 4
	self.MT.SoundRate = (flags >> 2) & 0x3
	self.MT.SoundSize = (flags >> 1) & 0x1
	self.MT.SoundType = flags & 0x1
	n++
	switch self.MT.SoundFormat {
	case SOUND_AAC:
		self.MT.AACPacketType = b[1]
		n++
	}
	return
}

func (self *Tag) parseVideoHeader(b []byte) (n int, err error) {
	if len(b) < n+5 {
		err = fmt.Errorf("invalid videodata len=%d", len(b))
	}
	flags := b[0]
	self.MT.FrameType = flags >> 4
	self.MT.CodecID = flags & 0xf
	n++
	if self.MT.FrameType == FRAME_INTER || self.MT.FrameType == FRAME_KEY {
		self.MT.AVCPacketType = b[1]
		for i := 2; i < 5; i++ {
			self.MT.CompositionTime = self.MT.CompositionTime<<8 + int32(b[i])
		}
		n += 4
	}
	return
}

// ParseMeidaTagHeader, parse video, audio, tag header
func (self *Tag) ParseMeidaTagHeader(b []byte) (n int, err error) {
	switch self.FT.Type {
	case TAG_AUDIO:
		n, err = self.parseAudioHeader(b)
	case TAG_VIDEO:
		n, err = self.parseVideoHeader(b)
	}
	return
}

// ParseFlvTagHeader ,parse FLV tag Header
func ParseFlvTagHeader(b []byte) (tag Tag, err error) {
	tagType := b[0]
	switch tagType {
	case TAG_AUDIO, TAG_VIDEO, TAG_SCRIPTDATA:
		tag.FT.Type = tagType
	default:
		err = fmt.Errorf("invalid tagType:%d", tagType)
		return
	}
	tag.FT.DataSize = uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	ts := uint32(b[4])<<16 | uint32(b[5])<<8 | uint32(b[6])
	tso := b[7]
	tag.FT.TimeStamp = ts | uint32(tso)<<24
	return
}
