package httpflv

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/amf"
	"bzcom/biubiu/media/utils/pio"
	"bzcom/biubiu/media/utils/uid"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	handler av.Handler
}

func NewServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

func (self *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		self.handleConn(w, r)
	})
	http.Serve(l, mux)
	return nil
}

func (self *Server) handleConn(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	u := r.URL.Path
	if pos := strings.LastIndex(u, "."); pos < 0 || u[pos:] != ".flv" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	path := strings.TrimSuffix(strings.TrimLeft(u, "/"), ".flv")
	patths := strings.SplitN(path, "/", 2)
	if len(patths) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	writer := NewFLVWriter(patths[0], patths[1], url, w)
	self.handler.HandleWriter(writer)
	writer.Wait()
}

const (
	headerLen = 11
)

type FLVWriter struct {
	av.RWBaser
	app, title, url string
	buf             []byte
	closed          chan struct{}
	ctx             http.ResponseWriter
}

func NewFLVWriter(app, title, url string, ctx http.ResponseWriter) *FLVWriter {
	ret := &FLVWriter{
		app:     app,
		title:   title,
		url:     url,
		ctx:     ctx,
		RWBaser: av.NewRWBaser(),
		closed:  make(chan struct{}),
		buf:     make([]byte, headerLen),
	}
	ret.ctx.Header().Set("Access-Control-Allow-Origin", "*")
	ret.ctx.Write([]byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09})
	pio.PutI32BE(ret.buf[:4], 0)
	ret.ctx.Write(ret.buf[:4])

	return ret
}

func (self *FLVWriter) Write(p av.Packet) error {
	self.RWBaser.SetPreTime()
	h := self.buf[:headerLen]
	typeID := av.TAG_VIDEO
	if !p.IsVideo {
		if p.IsMetadata {
			var err error
			typeID = av.TAG_SCRIPTDATAAMF0
			p.Data, err = amf.MetaDataReform(p.Data, amf.DEL)
			if err != nil {
				return err
			}
		} else {
			typeID = av.TAG_AUDIO
		}
	}
	dataLen := len(p.Data)
	timestamp := p.TimeStamp
	timestamp += self.BaseTimeStamp()
	self.RWBaser.RecTimeStamp(timestamp, uint32(typeID))

	preDataLen := dataLen + headerLen
	timestampbase := timestamp & 0xffffff
	timestampExt := timestamp >> 24 & 0xff

	pio.PutU8(h[0:1], uint8(typeID))
	pio.PutI24BE(h[1:4], int32(dataLen))
	pio.PutI24BE(h[4:7], int32(timestampbase))
	pio.PutU8(h[7:8], uint8(timestampExt))

	if _, err := self.ctx.Write(h); err != nil {
		return err
	}
	if _, err := self.ctx.Write(p.Data); err != nil {
		return err
	}

	pio.PutI32BE(h[:4], int32(preDataLen))
	if _, err := self.ctx.Write(h[:4]); err != nil {
		return err
	}

	return nil
}

func (self *FLVWriter) Wait() {
	for {
		select {
		case <-self.closed:
			return
		default:
		}
	}
}

func (self *FLVWriter) Close(error) {
	close(self.closed)
}

func (self *FLVWriter) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	ret.URL = self.url
	ret.Key = self.app + "/" + self.title
	return
}
