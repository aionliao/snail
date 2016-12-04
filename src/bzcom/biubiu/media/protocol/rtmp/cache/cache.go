package cache

import (
	//"log"
	"sync"
	"bzcom/biubiu/media/av"
)


const(
	defaultGopNum = 3
)

type Cache struct {
	mutex              sync.RWMutex
	gop                *GopCache
	videoSeq           *SpecialCache
	audioSeq           *SpecialCache
	metadata           *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(defaultGopNum),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}


func (self *Cache) Write(p av.Packet) {
	self.mutex.Lock()
	if p.IsMetadata {
		self.metadata.Write(p)
		goto end
	} else {
		if !p.IsVideo {
			ah := p.Header.(av.AudioPacketHeader)
			if ah.SoundFormat() == av.SOUND_AAC &&
			   ah.AACPacketType() == av.AAC_SEQHDR {
				self.audioSeq.Write(p)
			  goto end
			}
		} else {
			vh := p.Header.(av.VideoPacketHeader)
			if vh.IsSeq() {
				self.videoSeq.Write(p)
				goto end
			}
		}
	}
	self.gop.Write(p)

end:
	self.mutex.Unlock()
}

func (self *Cache) Read(pos, flag int,curid int64)(packet av.Packet,nextpos int,id int64,err error){
	self.mutex.RLock()
	if flag != 0 {
		switch flag{
		case 3:
			packet,err = self.metadata.Read()
		case 2:
			packet,err = self.videoSeq.Read()
		case 1:
			packet,err = self.audioSeq.Read()
		}
		nextpos = pos
	}else{
	  packet, nextpos,id,err = self.gop.Read(pos,curid)
  }
	self.mutex.RUnlock()

	return
}
