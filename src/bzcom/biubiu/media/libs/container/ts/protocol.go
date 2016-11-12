package ts

import "io"

type TSMuxer interface {
	Mux(*Frame, io.Writer) error
}

func NewTSMuxer() TSMuxer {
	return NewMuxer()
}
