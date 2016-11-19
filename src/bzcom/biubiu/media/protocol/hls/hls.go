package hls

/**
  hls ,semi-finished products
*/
import (
	"bytes"
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/container/flv"
	"bzcom/biubiu/media/container/ts"
	"bzcom/biubiu/media/parser"
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

const (
	videoHZ                = 90000
	aacSampleLen           = 1024
	h264_default_hz uint64 = 90
)

type Source struct {
	av.RWBaser
	stopd     bool
	info      av.Info
	bwriter   *bytes.Buffer
	btswriter *bytes.Buffer
	demuxer   *flv.Demuxer
	muxer     *ts.Muxer
	pts, dts  uint64
	stat      *status
	align     *align
	cache     *audioCache
	tsparser  *parser.CodecParser
}

func NewSource(info av.Info) *Source {
	info.Inter = true
	return &Source{
		info:     info,
		align:    &align{},
		stat:     newStatus(),
		cache:    newAudioCache(),
		demuxer:  flv.NewDemuxer(),
		muxer:    ts.NewMuxer(),
		tsparser: parser.NewCodecParser(),
		bwriter:  bytes.NewBuffer(make([]byte, 100*1024)),
	}
}

func (self *Source) segdo() {
	newf := true
	if self.btswriter == nil {
		self.btswriter = bytes.NewBuffer(nil)
	} else if self.btswriter != nil && self.stat.durationMs() >= 5000 {
		self.flushAudio()
		self.btswriter.Reset()
		self.stat.resetAndNew()
	} else {
		newf = false
	}
	if newf {
		self.btswriter.Write(self.muxer.PAT())
		self.btswriter.Write(self.muxer.PMT(av.SOUND_AAC, true))
	}
}

func (self *Source) calcPtsDts(isVideo bool, ts, compositionTs uint32) {
	self.dts = uint64(ts) * h264_default_hz
	if isVideo {
		self.pts = self.dts + uint64(compositionTs)*h264_default_hz
	} else {
		sampleRate, _ := self.tsparser.SampleRate()
		self.align.align(&self.dts, uint32(videoHZ*aacSampleLen/sampleRate))
		self.pts = self.dts
	}
}
func (self *Source) flushAudio() error {
	return self.muxAudio(1)
}

func (self *Source) muxAudio(limit byte) error {
	if self.cache.CacheNum() < limit {
		return nil
	}
	var p av.Packet
	_, pts, buf := self.cache.GetFrame()
	p.Data = buf
	p.TimeStamp = uint32(pts / h264_default_hz)
	return self.muxer.Mux(&p, self.btswriter)
}

func (self *Source) tsMux(p *av.Packet) error {
	if p.IsVideo {
		return self.muxer.Mux(p, self.btswriter)
	} else {
		self.cache.Cache(p.Data, self.pts)
		return self.muxAudio(cache_max_frames)
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
	var compositionTime int32
	var ah av.AudioPacketHeader
	var vh av.VideoPacketHeader
	if p.IsVideo {
		vh = p.Header.(av.VideoPacketHeader)
		if vh.CodecID() != av.VIDEO_H264 {
			return nil
		}
		compositionTime = vh.CompositionTime()
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
	self.bwriter.Reset()
	err := self.tsparser.Parse(&p, self.bwriter)
	if err != nil {
		return err
	}
	p.Data = self.bwriter.Bytes()

	// mux ts
	if p.IsVideo && vh.IsKeyFrame() {
		self.segdo()
	}

	if self.btswriter != nil {

		self.stat.update(p.IsVideo, p.TimeStamp)

		self.calcPtsDts(p.IsVideo, p.TimeStamp, uint32(compositionTime))

		self.tsMux(&p)

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
