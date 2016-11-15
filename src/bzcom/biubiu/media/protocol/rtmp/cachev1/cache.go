package cachev1

import "bzcom/biubiu/media/av"

type Cache struct {
	gop                *Gop
	videoSeq           *Seq
	audioSeq           *Seq
	metadata           *Metadata
	hasSefaultMinTs    bool
	lastVideoTimestamp uint32
	lastAudioTimestamp uint32
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGop(),
		videoSeq: NewSeq(),
		audioSeq: NewSeq(),
		metadata: NewMetadata(),
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
