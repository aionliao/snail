package hls

import (
	"bytes"
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/container/flv"
	"bzcom/biubiu/media/container/ts"
	"bzcom/biubiu/media/parser"
	"log"
	"net"
	"net/http"
)

type Server struct {
	conns map[string]*Source
}

func NewServer() *Server {
	return &Server{
		conns: make(map[string]*Source),
	}
}

func (self *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})
	http.Serve(l, mux)
	return nil
}

func (self *Server) GetWriter(info av.Info) av.WriteCloser {
	s := NewSource(info)
	_, ok := self.conns[info.Key]
	if !ok {
		self.conns[info.Key] = s
	}
	return s
}

type Source struct {
	av.RWBaser
	startTs    uint32
	stopd      bool
	info       av.Info
	bwriter    *bytes.Buffer
	btswriter  *bytes.Buffer
	demuxer    *flv.Demuxer
	muxer      *ts.Muxer
	audioCache *AudioCache
	tsparser   *parser.CodecParser
}

func NewSource(info av.Info) *Source {
	info.Inter = true
	return &Source{
		info:     info,
		demuxer:  flv.NewDemuxer(),
		muxer:    ts.NewMuxer(),
		tsparser: parser.NewCodecParser(),
		bwriter:  bytes.NewBuffer(make([]byte, 100*1024)),
	}
}

func (self *Source) Write(p av.Packet) error {
	if p.IsMetadata {
		return nil
	}
	self.SetPreTime()
	if err := self.demuxer.Demux(&p); err != nil {
		return err
	}

	// first ,parse aac, h264
	self.bwriter.Reset()
	var ah av.AudioPacketHeader
	var vh av.VideoPacketHeader
	if p.IsVideo {
		vh = p.Header.(av.VideoPacketHeader)
		if vh.CodecID() != av.VIDEO_H264 {
			return nil
		}
		if vh.IsKeyFrame() && vh.IsSeq() {
			return self.tsparser.Parse(&p, self.bwriter)
		}
	} else {
		ah = p.Header.(av.AudioPacketHeader)
		if ah.SoundFormat() != av.SOUND_AAC {
			return nil
		}
		if ah.AACPacketType() == av.AAC_SEQHDR {
			return self.tsparser.Parse(&p, self.bwriter)
		}
	}
	if err := self.tsparser.Parse(&p, self.bwriter); err != nil {
		return err
	}
	p.Data = self.bwriter.Bytes()

	// mux ts
	if p.IsVideo && vh.IsKeyFrame() {
		if self.btswriter == nil {
			log.Println("new ts file")
			self.startTs = p.TimeStamp
			self.btswriter = bytes.NewBuffer(nil)
			self.btswriter.Write(self.muxer.PAT())
			self.btswriter.Write(self.muxer.PMT(av.SOUND_AAC, true))
		} else {
			if p.TimeStamp-self.startTs >= 5000 {
				self.startTs = p.TimeStamp
				log.Println("close old and new file")
				self.btswriter.Reset()
				self.btswriter.Write(self.muxer.PAT())
				self.btswriter.Write(self.muxer.PMT(av.SOUND_AAC, true))
			}
		}
	}
	if self.btswriter != nil {
		if err := self.muxer.Mux(&p, self.btswriter); err != nil {
			return err
		}
	}

	return nil
}

func (self *Source) Info() (ret av.Info) {
	return self.info
}

func (self *Source) Stopd() bool {
	return self.stopd
}

func (self *Source) Close(err error) {
	self.stopd = true
}
