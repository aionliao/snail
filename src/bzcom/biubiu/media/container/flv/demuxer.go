package flv

import "bzcom/biubiu/media/av"

type Demuxer struct {
}

func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

func (self *Demuxer) DemuxH(p *av.Packet) error {
	var tag Tag
	_, err := tag.ParseMeidaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag

	return nil
}

func (self *Demuxer) Demux(p *av.Packet) error {
	var tag Tag
	n, err := tag.ParseMeidaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag
	p.Data = p.Data[n:]

	return nil
}
