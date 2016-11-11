package flv

import (
	"bytes"
	"io"
)

type demuxer struct {
	data bytes.Buffer
}

func newDemuxer() *demuxer {
	return &demuxer{}
}

func parseMediaHeader(b []byte, tagType uint8) (tag Tag, n int, err error) {
	tag.FT.Type = tagType
	n, err = tag.ParseMeidaTagHeader(b)
	if err != nil {
		return
	}
	return
}

// Demux,parse flv tag data and return tag which  has data
func (self *demuxer) Demux(b []byte, tagType uint8) (tag Tag, err error) {
	var n int
	tag, n, err = parseMediaHeader(b, tagType)
	if err != nil {
		return
	}
	tag.Data = b[n:]
	return
}

// DemuxWithWriter, parse tag and write data to w, reduce copy
func (self *demuxer) DemuxWithWriter(b []byte, tagType uint8, w io.Writer) (tag Tag, err error) {
	var n int
	tag, n, err = parseMediaHeader(b, tagType)
	if err != nil {
		return
	}
	_, err = w.Write(b[n:])
	return
}
