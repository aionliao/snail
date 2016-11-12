package codec

import (
	"bzcom/biubiu/media/libs/container/flv"
	"io"
)

type SampleRater interface {
	SampleRate() (int, error)
}

type CodecDemuxer interface {
	SampleRater
	Demux(*flv.Tag, io.Writer) error
}

func NewCodecDemuxer() CodecDemuxer {
	return newCodecDemuxer()
}
