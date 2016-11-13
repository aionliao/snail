package codec

import (
	"bzcom/biubiu/media/libs/common"
	"bzcom/biubiu/media/libs/parser/aac"
	"bzcom/biubiu/media/libs/parser/h264"
	"bzcom/biubiu/media/libs/parser/mp3"
	"errors"
	"io"
)

var (
	errNoAudio = errors.New("demuxer no audio")
)

type CodecParser struct {
	aac  *aac.Parser
	mp3  *mp3.Parser
	h264 *h264.Parser
}

func NewCodecParser() *CodecParser {
	return &CodecParser{}
}

func (self *CodecParser) SampleRate() (int, error) {
	if self.aac == nil && self.mp3 == nil {
		return 0, errNoAudio
	}
	if self.aac != nil {
		return self.aac.SampleRate(), nil
	}
	return self.mp3.SampleRate(), nil
}

func (self *CodecParser) Parse(p *common.Packet, w io.Writer) (err error) {

	switch p.IsVideo {
	case true:
		f, ok := p.Header.(common.VideoPacketHeader)
		if ok {
			if f.CodecID() == common.VIDEO_H264 {
				if self.h264 == nil {
					self.h264 = h264.NewParser()
				}
				err = self.h264.Parse(p.Data, f.IsSeq(), w)
			}
		}
	case false:
		f, ok := p.Header.(common.AudioPacketHeader)
		if ok {
			switch f.SoundFormat() {
			case common.SOUND_AAC:
				if self.aac == nil {
					self.aac = aac.NewParser()
				}
				err = self.aac.Parse(p.Data, f.AACPacketType(), w)
			case common.SOUND_MP3:
				if self.mp3 == nil {
					self.mp3 = mp3.NewParser()
				}
				err = self.mp3.Parse(p.Data)
			}
		}

	}
	return
}
