package flv

import "bzcom/biubiu/media/av"

type Demuxer struct {
}

func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

// Demux,parse flv tag data and return tag which  has data
func (self *Demuxer) Demux(p *av.Packet) (*av.Packet, error) {
	var tag Tag
	n, err := tag.ParseMeidaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return nil, err
	}
	p.Header = &tag
	p.Data = p.Data[n:]
	return p, nil
}
