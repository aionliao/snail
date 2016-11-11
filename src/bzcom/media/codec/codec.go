package codec

import (
	"bzcom/media/codec/aac"
	"bzcom/media/codec/avc"
	"bzcom/media/codec/mp3"
	"bzcom/media/container/flv"
	"errors"
	"io"
)

var (
	errNoAudio = errors.New("demuxer no audio")
)

type codecDemuxer struct {
	aacDemuxer  *aac.Demuxer
	mp3Demuxer  *mp3.Demuxer
	h264Demuxer *avc.Demuxer
}

func newCodecDemuxer() *codecDemuxer {
	return &codecDemuxer{}
}

func (self *codecDemuxer) SampleRate() (int, error) {
	if self.aacDemuxer == nil && self.mp3Demuxer == nil {
		return 0, errNoAudio
	}
	if self.aacDemuxer != nil {
		return self.aacDemuxer.SampleRate(), nil
	}
	return self.mp3Demuxer.SampleRate(), nil
}

func (self *codecDemuxer) Demux(tag *flv.Tag, w io.Writer) (err error) {
	switch tag.FT.Type {
	case flv.TAG_VIDEO:
		if tag.MT.CodecID == flv.VIDEO_H264 {
			if self.h264Demuxer == nil {
				self.h264Demuxer = avc.NewDemuxer()
			}
			err = self.h264Demuxer.Demux(tag, w)
		}
	case flv.TAG_AUDIO:
		if tag.MT.SoundFormat == flv.SOUND_AAC {
			if self.aacDemuxer == nil {
				self.aacDemuxer = aac.NewDemuxer()
			}
			err = self.aacDemuxer.Demux(tag, w)
		} else if tag.MT.SoundFormat == flv.SOUND_MP3 {
			if self.mp3Demuxer == nil {
				self.mp3Demuxer = mp3.NewDemuxer()
			}
			err = self.mp3Demuxer.Demux(tag.Data)
		}
	}
	return
}
