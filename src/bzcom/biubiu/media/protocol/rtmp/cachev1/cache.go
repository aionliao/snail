package cachev1

/***
implement a easy's cache which contains gop, seq, metadata
*/

import (
	"bzcom/biubiu/media/av"
	"sync"
)

type Cache struct {
	gop                *Gop
	videoSeq           *SpecialData
	audioSeq           *SpecialData
	metadata           *SpecialData
	lock               sync.RWMutex
	hasSefaultMinTs    bool
	lastVideoTimestamp uint32
	lastAudioTimestamp uint32
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGop(2),
		videoSeq: NewSpecialData(),
		audioSeq: NewSpecialData(),
		metadata: NewSpecialData(),
	}
}

func (self *Cache) Write(p *av.Packet) {
	if p.IsMetadata {
		self.metadata.Write(p)
		return
	} else {
		if !p.IsVideo {
			ah := p.Header.(av.AudioPacketHeader)
			if ah.SoundFormat() == av.SOUND_AAC && ah.AACPacketType() == av.AAC_SEQHDR {
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
