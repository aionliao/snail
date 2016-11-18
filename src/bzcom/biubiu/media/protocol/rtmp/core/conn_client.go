package core

import (
	"bytes"
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/amf"
	"fmt"
	"log"
	"math/rand"
	"net"
	neturl "net/url"
	"reflect"
	"strings"
)

var (
	connectSuccess = "NetConnection.Connect.Success"
	publishStart   = "NetStream.Publish.Start"
	playStart      = "NetStream.Play.Start"
)

type ConnClient struct {
	done       bool
	transID    int
	url        string
	tcurl      string
	app        string
	title      string
	query      string
	curcmdName string
	streamid   uint32
	conn       *Conn
	encoder    *amf.Encoder
	decoder    *amf.Decoder
}

func NewConnClient() *ConnClient {
	return &ConnClient{
		transID: 1,
		encoder: &amf.Encoder{},
		decoder: &amf.Decoder{},
	}
}

func (self *ConnClient) handleMsg() error {
	var err error
	var rc ChunkStream
	for {
		if err = self.conn.Read(&rc); err != nil {
			return err
		}
		switch rc.TypeID {
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			v, err := self.decoder.Decode(r, amf.AMF0)
			if err != nil {
				return err
			}
			switch v.(string) {
			case "_result", "onStatus":
				for {
					v, err = self.decoder.Decode(r, amf.AMF0)
					if err != nil {
						break
					}
					switch v.(type) {
					case amf.Object:
						objMap := v.(amf.Object)
						code, ok := objMap["code"]
						if ok {
							if code.(string) == connectSuccess ||
								code.(string) == publishStart ||
								code.(string) == playStart {
								return nil
							}
						}
					case float64:
						if int(v.(float64)) == self.transID {
							if self.curcmdName == cmdCreateStream {
								return nil
							}
						}
					default:
					}
				}
			case "onBWDone":
			default:
				return fmt.Errorf("_error")
			}
		}
	}
}

func (self *ConnClient) writeMsg(args ...interface{}) error {
	w := bytes.NewBuffer(nil)
	for _, v := range args {
		if _, err := self.encoder.Encode(w, v, amf.AMF0); err != nil {
			return err
		}
	}
	c := ChunkStream{
		Format:    0,
		CSID:      3,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  self.streamid,
		Length:    uint32(len(w.Bytes())),
		Data:      w.Bytes(),
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnClient) writeConnectMsg() error {
	event := make(amf.Object)
	event["app"] = self.app
	event["type"] = "nonprivate"
	event["flashVer"] = "FMS.3.1"
	event["tcUrl"] = self.tcurl
	self.curcmdName = cmdConnect
	log.Println("v", reflect.ValueOf(event).Kind())
	if err := self.writeMsg(cmdConnect, self.transID, event); err != nil {
		return err
	}
	return self.handleMsg()
}

func (self *ConnClient) writeCreateStreamMsg() error {
	self.transID++
	self.curcmdName = cmdCreateStream
	if err := self.writeMsg(cmdCreateStream, self.transID, nil); err != nil {
		return err
	}
	return self.handleMsg()
}

func (self *ConnClient) writePublishMsg() error {
	self.transID++
	self.streamid = 1
	self.curcmdName = cmdPublish
	if err := self.writeMsg(cmdPublish, self.transID, nil, self.title, publishLive); err != nil {
		return err
	}
	return self.handleMsg()
}

func (self *ConnClient) writePlayMsg() error {
	self.transID++
	self.streamid = 1
	self.curcmdName = cmdPlay
	if err := self.writeMsg(cmdPlay, 0, nil, self.title); err != nil {
		return err
	}
	return self.handleMsg()
}

func (self *ConnClient) Start(url string, method string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	self.url = url
	path := strings.TrimLeft(u.Path, "/")
	ps := strings.SplitN(path, "/", 2)
	if len(ps) != 2 {
		return fmt.Errorf("u path err: %s", path)
	}
	self.app = ps[0]
	self.title = ps[1]
	self.query = u.RawQuery
	self.tcurl = "rtmp://" + u.Host + "/" + self.app
	port := ":1935"
	host := u.Host
	localIP := ":0"
	var remoteIP string
	if strings.Index(host, ":") != -1 {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return err
		}
		port = ":" + port
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}
	remoteIP = ips[rand.Intn(len(ips))].String()
	if strings.Index(remoteIP, ":") == -1 {
		remoteIP += port
	}

	local, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		return err
	}
	remote, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		return err
	}

	log.Println("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

	self.conn = NewConn(conn, 4*1024)
	if err := self.conn.HandshakeClient(); err != nil {
		return err
	}

	if err := self.writeConnectMsg(); err != nil {
		return err
	}

	if err := self.writeCreateStreamMsg(); err != nil {
		return err
	}
	if method == av.PUBLISH {
		if err := self.writePublishMsg(); err != nil {
			return err
		}
	} else if method == av.PLAY {
		if err := self.writePlayMsg(); err != nil {
			return err
		}
	}

	return nil
}

func (self *ConnClient) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.ADD); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return self.conn.Write(&c)
}

func (self *ConnClient) Read(c *ChunkStream) (err error) {
	err = self.conn.Read(c)
	return err
}

func (self *ConnClient) GetInfo() (app string, name string, url string) {
	app = self.app
	name = self.title
	url = self.url
	return
}

func (self *ConnClient) Close(err error) {
	self.conn.Close()
}
