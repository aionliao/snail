package flv

type demuxer struct {
}

func newDemuxer() *demuxer {
	return &demuxer{}
}

// Demux,parse flv tag data and return tag which  has data
func (self *demuxer) Demux(b []byte, tagType uint8) (tag Tag, err error) {
	var n int
	tag.FT.Type = tagType
	n, err = tag.ParseMeidaTagHeader(b)
	if err != nil {
		return
	}
	tag.Data = b[n:]
	return
}
