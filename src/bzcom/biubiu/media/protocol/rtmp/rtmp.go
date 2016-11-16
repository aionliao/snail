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

type VirWriter struct {
	t    time.Time
	conn *core.ConnServer
}

func NewVirWriter(conn *core.ConnServer) *VirWriter {
	return &VirWriter{
		conn: conn,
	}
}

func (self *VirWriter) Write(p av.Packet) error {
	var cs core.ChunkStream
	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = 1
	if p.IsVideo {
		cs.TypeID = av.TAG_VIDEO
	} else {
		cs.TypeID = av.TAG_AUDIO
	}
	cs.Timestamp = p.TimeStamp
	self.t = time.Now()
	return self.conn.Write(cs)
}

func (self *VirWriter) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	ret.Key = self.conn.ConnInfo.App + "/" + self.conn.PublishInfo.Name
	ret.URL = self.conn.ConnInfo.TcUrl + "/" + self.conn.PublishInfo.Name
	return
}

func (self *VirWriter) Close(err error) {
	self.conn.Close(err)
}

func (self *VirWriter) Alive() bool {
	if time.Now().Sub(self.t) >= time.Second*10 {
		self.conn.Close(errors.New("write timeout"))
		return false
	}
	return true
}

type VirReader struct {
	t       time.Time
	demuxer *flv.Demuxer
	conn    *core.ConnServer
}

func NewVirReader(conn *core.ConnServer) *VirReader {
	return &VirReader{
		conn:    conn,
		demuxer: flv.NewDemuxer(),
	}
}

func (self *VirReader) Read(p *av.Packet) error {
	var cs core.ChunkStream
	err := self.conn.Read(&cs)
	if err != nil {
		return err
	}
	self.t = time.Now()
	p.IsVideo = cs.TypeID == av.TAG_VIDEO
	p.IsMetadata = (cs.TypeID == 0x12 || cs.TypeID == 0xf)
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp
	self.demuxer.Demux(p)
	return err
}

func (self *VirReader) Info() (ret av.Info) {
	ret.UID = uid.NEWID()
	ret.Key = self.conn.ConnInfo.App + "/" + self.conn.PublishInfo.Name
	ret.URL = self.conn.ConnInfo.TcUrl + "/" + self.conn.PublishInfo.Name
	return
}

func (self *VirReader) Close(err error) {
	self.conn.Close(err)
}

func (self *VirReader) Alive() bool {
	if time.Now().Sub(self.t) >= time.Second*10 {
		self.conn.Close(errors.New("read timeout"))
		return false
	}
	return true
}
