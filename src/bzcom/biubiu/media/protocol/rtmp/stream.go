package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/rtmp/cache"
	"errors"
	"sync"
	"time"
	"log"
)


var (
	EmptyID = ""
)

type RtmpStream struct {
	lock    sync.RWMutex
	streams map[string]Stream
}

func NewRtmpStream() *RtmpStream {
	ret := &RtmpStream{
		streams: make(map[string]Stream),
	}
	go ret.CheckAlive()
	return ret
}

func (self *RtmpStream) HandleReader(r av.ReadCloser) {
	self.lock.Lock()
	info := r.Info()
	s, ok := self.streams[info.Key]
	if !ok {
		log.Println("new stream by reader in")
		s = NewStream()
		s.AddReader(r)
		self.streams[info.Key] = s
	} else {
		id := s.ID()
		if id != EmptyID && id != info.UID {
			log.Println("reader change")
			s.ReaderStop()
			ns := NewStream()
			ns.AddReader(r)
			s.Copy(&ns)
			self.streams[info.Key] = ns
		}else{
			log.Println("reader reach")
			s.AddReader(r)
			self.streams[info.Key] = s
		}
	}
	self.lock.Unlock()
}

func (self *RtmpStream) HandleWriter(w av.WriteCloser) {
	self.lock.Lock()
	info := w.Info()
	s, ok := self.streams[info.Key]
	if !ok {
		log.Println("new stream by writer first in")
		s = NewStream()
	}
	s.AddWriter(w)
	self.streams[info.Key] = s
	self.lock.Unlock()
}

func (self *RtmpStream) CheckAlive() {
	for {
		time.Sleep(time.Second * 5)
		self.lock.Lock()
		for k, v := range self.streams {
			if v.CheckAlive() == 0 {
				log.Println("stream not alive,so del it")
				delete(self.streams, k)
			}
		}
		self.lock.Unlock()
	}
}


var (
	ErrWriteTimeout  = errors.New("write timeout")
	ErrReadTimeout   = errors.New("read timeout")
	ErrForceClose    = errors.New("force close")
	ErrStopOldReader = errors.New("stop old reader")
)

type Stream struct {
	cache   *cache.Cache
	r       av.ReadCloser
	ws      map[string]destConsole
}


func NewStream() Stream {
	return Stream{
		ws:    make(map[string]destConsole),
	}
}

func (self *Stream) ID() string {
	if self.r != nil {
		return self.r.Info().UID
	}
	return EmptyID
}

func (self *Stream) Copy(dst *Stream) {
	for _, v := range self.ws {
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
	}
}

func (self *Stream) AddReader(r av.ReadCloser) {
	self.r = r
	if self.cache == nil{
		self.cache = cache.NewCache()
	}
	sc := newSourceConsole(r, *self.cache)
	go sc.do()
}

func (self *Stream) AddWriter(w av.WriteCloser) {
	info := w.Info()
	if self.cache == nil{
		self.cache = cache.NewCache()
	}
	wc := newDestConsole(w, *self.cache)
	self.ws[info.UID] = wc
	go wc.do()
}

func (self *Stream) ReaderStop() {
	if  self.r != nil {
		log.Println("force stop reader,and close")
		self.r.Close(ErrStopOldReader)
		for _, v := range self.ws {
			log.Println("force stop writer,not close")
			v.close()
			if v.w.Info().IsInterval() {
				v.w.Close(ErrForceClose)
			}
		}
	}
}

func (self *Stream) CheckAlive() (n int) {
	if self.r != nil  {
		if self.r.Alive() {
			n++
		} else {
			log.Println("reader not alive")
			self.r.Close(ErrReadTimeout)
		}
	}
	for k, v := range self.ws {
		if !v.w.Alive() {
			log.Println("writer not alive")
			delete(self.ws, k)
			v.close()
			v.w.Close(ErrWriteTimeout)
			continue
		}
		n++
	}
	return
}

type sourceConsole struct{
	cache  cache.Cache
	r       av.ReadCloser
}

func newSourceConsole(r av.ReadCloser,c cache.Cache) sourceConsole{
	return sourceConsole{
		r:r,
		cache:c,
	}
}

func(self*sourceConsole)do(){
	var p av.Packet
	for {
		err := self.r.Read(&p)
		if err != nil {
			log.Println("reader close:",err)
			return
		}
		self.cache.Write(p)
	}
}

const(
	metadataFlag = 3
	videoSeqFlag = 2
	audioSeqFlag = 1
	normalFlag = 0
)

type destConsole struct{
	c cache.Cache
	w av.WriteCloser
	closed chan struct{}
}


func newDestConsole( w av.WriteCloser,c cache.Cache) destConsole{
	return destConsole{
		w:w,
		c:c,
		closed:make(chan struct{}),
	}
}

func (self *destConsole)close(){
	close(self.closed)
}

func (self *destConsole)do(){
	pos := -1
	flag := metadataFlag
	var indexId int64
	var err error
	var p av.Packet
	for{
		select{
		case <-self.closed:
			log.Println("write force stop")
			return
		default:
			self.c.Wait()
			for{
			p, pos,indexId, err = self.c.Read(pos,flag,indexId)
			if err == cache.Empty ||len(p.Data) == 0 {
				break
			}
			if p.IsMetadata{
				flag=videoSeqFlag
			}else{
				if p.Header == nil{
					break
				}
				if !p.IsVideo {
					ah := p.Header.(av.AudioPacketHeader)
					if ah.SoundFormat() == av.SOUND_AAC &&
						 ah.AACPacketType() == av.AAC_SEQHDR {
						 flag=normalFlag
					}
				} else {
					vh := p.Header.(av.VideoPacketHeader)
					if vh.IsSeq() {
						flag=audioSeqFlag
					}
				}
			}
			if err = self.w.Write(p);err !=nil{
				log.Println("write close:",err)
				return
			}
		}
	  }
	}
}
