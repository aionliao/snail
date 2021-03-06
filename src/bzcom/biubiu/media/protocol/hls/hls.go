package hls

import (
	"bytes"
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/container/flv"
	"bzcom/biubiu/media/container/ts"
	"bzcom/biubiu/media/parser"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoPublisher         = errors.New("No publisher")
	ErrInvalidReq          = errors.New("invalid req url path")
	ErrNoSupportVideoCodec = errors.New("no support video codec")
	ErrNoSupportAudioCodec = errors.New("no support audio codec")
)

var crossdomainxml = []byte(`<?xml version="1.0" ?>
<cross-domain-policy>
	<allow-access-from domain="*" />
	<allow-http-request-headers-from domain="*" headers="*"/>
</cross-domain-policy>`)

type Server struct {
	l     net.Listener
	lock  sync.RWMutex
	conns map[string]*Source
}

func NewServer() *Server {
	ret := &Server{
		conns: make(map[string]*Source),
	}
	go ret.checkStop()
	return ret
}

func (self *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		self.handle(w, r)
	})
	self.l = l
	http.Serve(l, mux)
	return nil
}

func (self *Server) GetWriter(info av.Info) av.WriteCloser {
	self.lock.RLock()
	s := NewSource(info)
	_, ok := self.conns[info.Key]
	if !ok {
		self.conns[info.Key] = s
	}
	self.lock.RUnlock()
	return s
}

func (self *Server) getConn(key string) *Source {
	self.lock.RLock()
	v, ok := self.conns[key]
	if !ok {
		self.lock.RUnlock()
		return nil
	}
	self.lock.RUnlock()
	return v
}

func (self *Server) checkStop() {
	for {
		time.Sleep(time.Second * 5)
		self.lock.RLock()
		for k, v := range self.conns {
			if !v.Alive() {
				delete(self.conns, k)
			}
		}
		self.lock.RUnlock()
	}
}

func (self *Server) handle(w http.ResponseWriter, r *http.Request) {
	if path.Base(r.URL.Path) == "crossdomain.xml" {
		w.Header().Set("Content-Type", "application/xml")
		w.Write(crossdomainxml)
		return
	}
	switch path.Ext(r.URL.Path) {
	case ".m3u8":
		key, _ := self.parseM3u8(r.URL.Path)
		self.lock.RLock()
		conn := self.getConn(key)
		if conn == nil {
			http.Error(w, ErrNoPublisher.Error(), http.StatusForbidden)
			return
		}
		tsCache := conn.GetCacheInc()
		body, err := tsCache.GenM3U8PlayList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/x-mpegURL")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body)
	case ".ts":
		key, _ := self.parseTs(r.URL.Path)
		conn := self.getConn(key)
		if conn == nil {
			http.Error(w, ErrNoPublisher.Error(), http.StatusForbidden)
			return
		}
		tsCache := conn.GetCacheInc()
		item, err := tsCache.GetItem(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "video/mp2ts")
		w.Header().Set("Content-Length", strconv.Itoa(len(item.Data)))
		w.Write(item.Data)
	default:
		http.Error(w, ErrInvalidReq.Error(), http.StatusBadRequest)
		return
	}
}

func (self *Server) parseM3u8(pathstr string) (key string, err error) {
	pathstr = strings.TrimLeft(pathstr, "/")
	key = strings.TrimRight(pathstr, path.Ext(pathstr))
	return
}

func (self *Server) parseTs(pathstr string) (key string, err error) {
	pathstr = strings.TrimLeft(pathstr, "/")
	paths := strings.SplitN(pathstr, "/", 3)
	if len(paths) != 3 {
		err = fmt.Errorf("invalid path=%s", pathstr)
		return
	}
	key = paths[0] + "/" + paths[1]

	return
}

const (
	videoHZ                = 90000
	aacSampleLen           = 1024
	h264_default_hz uint64 = 90
)

type Source struct {
	av.RWBaser
	seq       int
	info      av.Info
	bwriter   *bytes.Buffer
	btswriter *bytes.Buffer
	demuxer   *flv.Demuxer
	muxer     *ts.Muxer
	pts, dts  uint64
	stat      *status
	align     *align
	cache     *audioCache
	tsCache   *TSCacheItem
	tsparser  *parser.CodecParser
}

func NewSource(info av.Info) *Source {
	info.Inter = true
	return &Source{
		info:     info,
		align:    &align{},
		stat:     newStatus(),
		RWBaser:  av.NewRWBaser(time.Minute),
		cache:    newAudioCache(),
		demuxer:  flv.NewDemuxer(),
		muxer:    ts.NewMuxer(),
		tsCache:  NewTSCacheItem(info.Key),
		tsparser: parser.NewCodecParser(),
		bwriter:  bytes.NewBuffer(make([]byte, 100*1024)),
	}
}

func (self *Source) GetCacheInc() *TSCacheItem {
	return self.tsCache
}

func (self *Source) Write(p av.Packet) error {
	if p.IsMetadata {
		return nil
	}
	self.SetPreTime()
	err := self.demuxer.Demux(&p)
	if err != nil {
		if err == flv.ErrAvcEndSEQ {
			return nil
		}
		return err
	}

	compositionTime, isSeq, err := self.parse(&p)
	if err != nil || isSeq {
		return err
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

func (self *Source) Close(err error) {
	//
}

func (self *Source) cut() {
	newf := true
	if self.btswriter == nil {
		self.btswriter = bytes.NewBuffer(nil)
	} else if self.btswriter != nil && self.stat.durationMs() >= 5000 {
		self.flushAudio()

		self.seq++
		filename := fmt.Sprintf("%s%s/%d.ts", "/", self.info.Key, time.Now().Unix())
		item := NewTSItem(filename, int(self.stat.durationMs()), self.seq, self.btswriter.Bytes())
		self.tsCache.SetItem(filename, item)

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

func (self *Source) parse(p *av.Packet) (int32, bool, error) {
	var compositionTime int32
	var ah av.AudioPacketHeader
	var vh av.VideoPacketHeader
	if p.IsVideo {
		vh = p.Header.(av.VideoPacketHeader)
		if vh.CodecID() != av.VIDEO_H264 {
			return compositionTime, false, ErrNoSupportVideoCodec
		}
		compositionTime = vh.CompositionTime()
		if vh.IsKeyFrame() && vh.IsSeq() {
			return compositionTime, true, self.tsparser.Parse(p, self.bwriter)
		}
	} else {
		ah = p.Header.(av.AudioPacketHeader)
		if ah.SoundFormat() != av.SOUND_AAC {
			return compositionTime, false, ErrNoSupportAudioCodec
		}
		if ah.AACPacketType() == av.AAC_SEQHDR {
			return compositionTime, true, self.tsparser.Parse(p, self.bwriter)
		}
	}
	self.bwriter.Reset()
	if err := self.tsparser.Parse(p, self.bwriter); err != nil {
		return compositionTime, false, err
	}
	p.Data = self.bwriter.Bytes()

	if p.IsVideo && vh.IsKeyFrame() {
		self.cut()
	}

	return compositionTime, false, nil
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
