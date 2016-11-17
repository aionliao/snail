package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/container/flv"
	"bzcom/biubiu/media/protocol/rtmp/core"
	"bzcom/biubiu/media/utils/uid"
	"errors"
	"net"
	"time"
)

type Client struct {
	handler av.Handler
}

func NewRtmpClient(h av.Handler) *Client {
	return &Client{
		handler: h,
	}
}

func (self *Client) Dial(url string, method string) error {
	connClient := core.NewConnClient()
	if err := connClient.Start(url, method); err != nil {
		return err
	}
	if method == "publish" {
		writer := NewVirWriter(connClient)
		self.handler.HandleWriter(writer)
	} else if method == "play" {
		reader := NewVirReader(connClient)
		self.handler.HandleReader(reader)
	}
	return nil
}

type Server struct {
	handler av.Handler
}

func NewRtmpServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

func (self *Server) Serve(listener net.Listener) (err error) {
	for {
		var netconn net.Conn
		netconn, err = listener.Accept()
		if err != nil {
			return
		}
		conn := core.NewConn(netconn, 4*1024)
		go self.handleConn(conn)
	}
}

func (self *Server) handleConn(conn *core.Conn) error {
	conn.SetDeadline(time.Now().Add(time.Second * 30))
	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		return err
	}
	conn.SetDeadline(time.Time{})
	connServer := core.NewConnServer(conn)

	if err := connServer.ReadMsg(); err != nil {
		return err
	}
	if connServer.IsPublisher() {
		reader := NewVirReader(connServer)
		self.handler.HandleReader(reader)
	} else {
		writer := NewVirWriter(connServer)
		self.handler.HandleWriter(writer)
	}

	return nil
}

type GetInFo interface {
	GetInfo() (string, string, string)
}

type StreamReadWriteCloser interface {
	GetInFo
	Close(error)
	Write(core.ChunkStream) error
	Read(c *core.ChunkStream) error
}

type RWBaser struct {
	t  time.Time
	RW StreamReadWriteCloser
}

func NewRWBaser(rw StreamReadWriteCloser) RWBaser {
	return RWBaser{
		t:  time.Now(),
		RW: rw,
	}
}

func (self *RWBaser) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	app, title, url := self.RW.GetInfo()
	ret.URL = url
	ret.Key = app + "/" + title
	return
}

func (self *RWBaser) Close(err error) {
	self.RW.Close(err)
}

func (self *RWBaser) Alive() bool {
	if time.Now().Sub(self.t) >= time.Second*30 {
		self.RW.Close(errors.New("read timeout"))
		return false
	}
	return true
}

func (self *RWBaser) SetT() {
	self.t = time.Now()
}

type VirWriter struct {
	RWBaser
	lastVideoTs uint32
	lastAudioTs uint32
	maxTs       uint32
}

func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	return &VirWriter{
		RWBaser: NewRWBaser(conn),
	}
}

func (self *VirWriter) Write(p av.Packet) error {
	var cs core.ChunkStream
	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = 1
	cs.Timestamp = p.TimeStamp
	cs.Timestamp += self.maxTs

	if p.IsVideo {
		self.lastVideoTs = cs.Timestamp
		cs.TypeID = av.TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TAG_SCRIPTDATAAMF0
		} else {
			self.lastAudioTs = cs.Timestamp
			cs.TypeID = av.TAG_AUDIO
		}
	}
	self.SetT()
	return self.RW.Write(cs)
}

func (self *VirWriter) Reset() {
	if self.lastAudioTs > self.lastVideoTs {
		self.maxTs = self.lastAudioTs
	} else {
		self.maxTs = self.lastVideoTs
	}
}

type VirReader struct {
	demuxer *flv.Demuxer
	RWBaser
}

func NewVirReader(conn StreamReadWriteCloser) *VirReader {
	return &VirReader{
		RWBaser: NewRWBaser(conn),
		demuxer: flv.NewDemuxer(),
	}
}

func (self *VirReader) Read(p *av.Packet) (err error) {
	var cs core.ChunkStream
	for {
		err = self.RW.Read(&cs)
		if err != nil {
			return err
		}
		if cs.TypeID == av.TAG_AUDIO ||
			cs.TypeID == av.TAG_VIDEO ||
			cs.TypeID == av.TAG_SCRIPTDATAAMF0 ||
			cs.TypeID == av.TAG_SCRIPTDATAAMF3 {
			break
		}
	}
	self.SetT()
	p.IsVideo = cs.TypeID == av.TAG_VIDEO
	p.IsMetadata = (cs.TypeID == av.TAG_SCRIPTDATAAMF0 || cs.TypeID == av.TAG_SCRIPTDATAAMF3)
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp
	self.demuxer.Demux(p)
	return err
}
