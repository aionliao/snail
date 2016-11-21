package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/container/flv"
	"bzcom/biubiu/media/protocol/rtmp/core"
	"bzcom/biubiu/media/utils/uid"
	"net"
	"time"
)

type Client struct {
	handler av.Handler
	getter  av.GetWriter
}

func NewRtmpClient(h av.Handler, getter av.GetWriter) *Client {
	return &Client{
		handler: h,
		getter:  getter,
	}
}

func (self *Client) Dial(url string, method string) error {
	connClient := core.NewConnClient()
	if err := connClient.Start(url, method); err != nil {
		return err
	}
	if method == av.PUBLISH {
		writer := NewVirWriter(connClient)
		self.handler.HandleWriter(writer)
	} else if method == av.PLAY {
		reader := NewVirReader(connClient)
		self.handler.HandleReader(reader)
		if self.getter != nil {
			writer := self.getter.GetWriter(reader.Info())
			self.handler.HandleWriter(writer)
		}
	}
	return nil
}

type Server struct {
	handler av.Handler
	getter  av.GetWriter
}

func NewRtmpServer(h av.Handler, getter av.GetWriter) *Server {
	return &Server{
		handler: h,
		getter:  getter,
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
		if self.getter != nil {
			writer := self.getter.GetWriter(reader.Info())
			self.handler.HandleWriter(writer)
		}
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

type VirWriter struct {
	av.RWBaser
	conn StreamReadWriteCloser
}

func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	return &VirWriter{
		conn:    conn,
		RWBaser: av.NewRWBaser(time.Second * 10),
	}
}

func (self *VirWriter) Write(p av.Packet) error {
	var cs core.ChunkStream
	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = 1
	cs.Timestamp = p.TimeStamp
	cs.Timestamp += self.BaseTimeStamp()

	if p.IsVideo {
		cs.TypeID = av.TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TAG_SCRIPTDATAAMF0
		} else {
			cs.TypeID = av.TAG_AUDIO
		}
	}

	self.SetPreTime()
	self.RecTimeStamp(cs.Timestamp, cs.TypeID)
	return self.conn.Write(cs)
}

func (self *VirWriter) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	app, title, url := self.conn.GetInfo()
	ret.URL = url
	ret.Key = app + "/" + title
	return
}

func (self *VirWriter) Close(err error) {
	self.conn.Close(err)
}

type VirReader struct {
	av.RWBaser
	demuxer *flv.Demuxer
	conn    StreamReadWriteCloser
}

func NewVirReader(conn StreamReadWriteCloser) *VirReader {
	return &VirReader{
		conn:    conn,
		RWBaser: av.NewRWBaser(time.Second * 10),
		demuxer: flv.NewDemuxer(),
	}
}

func (self *VirReader) Read(p *av.Packet) (err error) {
	var cs core.ChunkStream
	for {
		err = self.conn.Read(&cs)
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
	self.SetPreTime()
	p.IsVideo = cs.TypeID == av.TAG_VIDEO
	p.IsMetadata = (cs.TypeID == av.TAG_SCRIPTDATAAMF0 || cs.TypeID == av.TAG_SCRIPTDATAAMF3)
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp
	self.demuxer.DemuxH(p)
	return err
}

func (self *VirReader) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	app, title, url := self.conn.GetInfo()
	ret.URL = url
	ret.Key = app + "/" + title
	return
}

func (self *VirReader) Close(err error) {
	self.conn.Close(err)
}
