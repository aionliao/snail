package core

import (
	"bytes"
	"bzcom/biubiu/media/protocol/amf"
	"io"
	"log"
)

var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

var (
	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

type ConnectResp struct {
	FMSVer       string `amf:"fmsVer"`
	Capabilities int    `amf:"capabilities"`
}

type ConnectEvent struct {
	Level          string `amf:"level"`
	Code           string `amf:"code"`
	Description    string `amf:"description"`
	ObjectEncoding int    `amf:"objectEncoding"`
}

type PublishInfo struct {
	Name string
	Type string
}

type ConnServer struct {
	done          bool
	err           error
	streamID      int
	isPublisher   bool
	conn          *Conn
	ConnInfo      ConnectInfo
	PublishInfo   PublishInfo
	decoder       *amf.Decoder
	transactionID float64
}

func NewConnServer(conn *Conn) *ConnServer {
	return &ConnServer{
		conn:     conn,
		streamID: 1,
		decoder:  &amf.Decoder{},
	}
}

func (self *ConnServer) connect(r io.Reader, amfType amf.Version) error {
	var err error
	var transid, objmap interface{}

	transid, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	self.transactionID = transid.(float64)

	objmap, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	infoMap := objmap.(amf.Object)
	v, ok := infoMap["app"]
	if ok {
		self.ConnInfo.App = v.(string)
	}
	v, ok = infoMap["flashVer"]
	if ok {
		self.ConnInfo.Flashver = v.(string)
	}
	v, ok = infoMap["tcUrl"]
	if ok {
		self.ConnInfo.TcUrl = v.(string)
	}
	v, ok = infoMap["objectEncoding"]
	if ok {
		self.ConnInfo.ObjectEncoding = int(v.(float64))
	}

	// TODO: Optional User  Arguments
	return nil
}

func (self *ConnServer) connectResp(cur *ChunkStream) error {
	c := self.conn.NewWindowAckSize(2500000)
	self.conn.Write(&c)
	c = self.conn.NewSetPeerBandwidth(2500000)
	self.conn.Write(&c)
	c = self.conn.NewSetChunkSize(uint32(1024))
	self.conn.Write(&c)

	w := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = self.ConnInfo.ObjectEncoding

	if _, err := encoder.EncodeBatch(w, amf.AMF0, "_result", self.transactionID, resp, event); err != nil {
		return err
	}
	c = ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnServer) createStream(r io.Reader, amfType amf.Version) error {
	var err error
	var transid interface{}

	transid, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	self.transactionID = transid.(float64)

	return nil
}

func (self *ConnServer) createStreamResp(cur *ChunkStream) error {
	w := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}

	if _, err := encoder.EncodeBatch(w, amf.AMF0, "_result", self.transactionID, nil, self.streamID); err != nil {
		return err
	}
	c := ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnServer) publishOrPlay(r io.Reader, cmdName string, amfType amf.Version) error {
	var err error
	var transid, name, pType interface{}

	transid, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	self.transactionID = transid.(float64)

	_, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	name, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	self.PublishInfo.Name = name.(string)

	if cmdName == cmdPlay {
		return nil
	}

	pType, err = self.decoder.Decode(r, amfType)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	self.PublishInfo.Type = pType.(string)

	return nil
}

func (self *ConnServer) publishResp(cur *ChunkStream) error {
	w := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publising."
	if _, err := encoder.EncodeBatch(w, amf.AMF0, "onStatus", 0, nil, event); err != nil {
		return err
	}
	c := ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnServer) playResp(cur *ChunkStream) error {
	w := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}

	self.conn.SetRecorded()
	self.conn.SetBegin()

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if _, err := encoder.EncodeBatch(w, amf.AMF0, "onStatus", 0, nil, event); err != nil {
		return err
	}
	c := ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if _, err := encoder.EncodeBatch(w, amf.AMF0, "onStatus", 0, nil, event); err != nil {
		return err
	}
	c = ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if _, err := encoder.EncodeBatch(w, amf.AMF0, "onStatus", 0, nil, event); err != nil {
		return err
	}
	c = ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if _, err := encoder.EncodeBatch(w, amf.AMF0, "onStatus", 0, nil, event); err != nil {
		return err
	}
	c = ChunkStream{
		Format:    0,
		CSID:      cur.CSID,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cur.StreamID,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)

	return self.conn.Flush()
}

func (self *ConnServer) handleCmdMsg(c *ChunkStream) error {
	amfType := amf.AMF0
	if c.TypeID == 17 {
		c.Data = c.Data[1:]
	}
	r := bytes.NewReader(c.Data)
	v, err := self.decoder.Decode(r, amf.Version(amfType))
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	switch v.(type) {
	case string:
		switch v.(string) {
		case cmdConnect:
			if err = self.connect(r, amf.Version(amfType)); err != nil {
				return err
			}
			if err = self.connectResp(c); err != nil {
				return err
			}
		case cmdCreateStream:
			if err = self.createStream(r, amf.Version(amfType)); err != nil {
				return err
			}
			if err = self.createStreamResp(c); err != nil {
				return err
			}
		case cmdPublish:
			if err = self.publishOrPlay(r, cmdPublish, amf.Version(amfType)); err != nil {
				return err
			}
			if err = self.publishResp(c); err != nil {
				return err
			}
			self.done = true
			self.isPublisher = true
		case cmdPlay:
			if err = self.publishOrPlay(r, cmdPlay, amf.Version(amfType)); err != nil {
				return err
			}
			if err = self.playResp(c); err != nil {
				return err
			}
			self.done = true
			self.isPublisher = false
		case cmdFcpublish:
		case cmdReleaseStream:
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			log.Println("no support command=", v.(string))
		}
	}

	return nil
}

func (self *ConnServer) ReadMsg() error {
	for {
		var c ChunkStream
		self.err = self.conn.Read(&c)
		if self.err != nil {
			return self.err
		}
		switch c.TypeID {
		case 20, 17:
			if err := self.handleCmdMsg(&c); err != nil {
				return err
			}
		}
		if self.done {
			break
		}
	}
	return nil
}

func (self *ConnServer) IsPublisher() bool {
	return self.isPublisher
}

func (self *ConnServer) Write(c ChunkStream) error {
	return self.conn.Write(&c)
}

func (self *ConnServer) Read(c *ChunkStream) (err error) {
	err = self.conn.Read(c)
	return
}

func (self *ConnServer) Close(err error) {
	self.conn.Close()
}