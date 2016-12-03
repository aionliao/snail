package cache

import "bzcom/biubiu/media/av"


type Cache struct {
	gop                *GopCache
	videoSeq           *SpecialCache
	audioSeq           *SpecialCache
	metadata           *SpecialCache
}

func NewCache(num int) Cache {
	return Cache{
		gop:      NewGopCache(num),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (self *Cache) Write(p av.Packet) {
	if p.IsMetadata {
		self.metadata.Write(p)
		return
	} else {
		if !p.IsVideo {
			ah := p.Header.(av.AudioPacketHeader)
			if ah.SoundFormat() == av.SOUND_AAC &&
			   ah.AACPacketType() == av.AAC_SEQHDR {
				self.audioSeq.Write(p)
				return
			}
		} else {
			vh := p.Header.(av.VideoPacketHeader)
			if vh.IsSeq() {
				self.videoSeq.Write(p)
				return
			}
		}
	}
	self.gop.Write(p)
}

func (self *Cache) Send(w av.WriteCloser) error {
	if err := self.metadata.Send(w); err != nil {
		return err
	}

	if err := self.videoSeq.Send(w); err != nil {
		return err
	}

	if err := self.audioSeq.Send(w); err != nil {
		return err
	}

	if err := self.gop.Send(w); err != nil {
		return err
	}

	return nil
}
