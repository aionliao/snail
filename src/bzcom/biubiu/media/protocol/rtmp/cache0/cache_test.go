package cache0

import (
  "testing"
  "bzcom/biubiu/media/av"
  "github.com/stretchr/testify/assert"
)

type mockPacketHeader struct{
  iskey bool
  isseq bool
  codid uint8
  ct   int32
}

func newMockPacketHeader(iskey,isseq bool,codid uint8,ct  int32 ) *mockPacketHeader{
  return &mockPacketHeader{
    iskey :iskey,
    isseq:isseq,
    codid:codid,
    ct:ct,
  }
}
func(self*mockPacketHeader)SoundFormat() uint8{
  return av.SOUND_AAC
}

func(self*mockPacketHeader)AACPacketType() uint8{
  return av.AAC_SEQHDR
}

func(self*mockPacketHeader)IsKeyFrame() bool{
  return self.iskey
}
func(self*mockPacketHeader)IsSeq() bool{
  return self.isseq
}

func(self*mockPacketHeader)CodecID() uint8{
  return self.codid
}

func(self*mockPacketHeader)CompositionTime() int32{
  return self.ct
}

func TestCache(t *testing.T){
    at := assert.New(t)
    cache := NewCache()
    // IsVideo    bool
  	// IsMetadata bool
  	// TimeStamp  uint32 // dts
  	// Header     PacketHeader
  	// Data       []byte

    // metadata
    p1 := av.Packet{
      IsVideo:false,
      IsMetadata:true,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p1)

    // video seq
    p2 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(true, true,7,0),
    }
    cache.Write(p2)

    // audio seq
    p3 := av.Packet{
      IsVideo:false,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p3)

    // key
    p4 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(true, false,7,0),
    }
    p4.Data[0]=0x1
    cache.Write(p4)

    p5 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p5)

    p6 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p6)

    p7 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p7)

    // 2. key
    p8 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(true, false,7,0),
    }
    cache.Write(p8)

    p9 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p9)

    p10 := av.Packet{
      IsVideo:true,
      IsMetadata:false,
      TimeStamp:0,
      Data:make([]byte,8),
      Header:newMockPacketHeader(false, false,7,0),
    }
    cache.Write(p10)

    pos := -1
    flag := 3
    var err error
    var id int64
    var ret av.Packet
    ret,pos, id,  err = cache.Read(pos, flag, id)
    at.Equal(err,nil)
    at.Equal(ret.IsMetadata,true)

    flag = 2
    ret,pos, id,  err = cache.Read(pos, flag, id)
    at.Equal(err,nil)
    at.Equal(ret.IsVideo,true)
    h := ret.Header.(av.VideoPacketHeader)
    at.Equal(h.IsKeyFrame(),true)

    flag = 1
    ret,pos, id,  err = cache.Read(pos, flag, id)
    at.Equal(err,nil)
    at.Equal(ret.IsVideo,false)
    ah := ret.Header.(av.AudioPacketHeader)
    at.Equal(int(ah.AACPacketType()),av.AAC_SEQHDR)

    flag = 0
    ret,pos, id,  err = cache.Read(pos, flag, id)
    at.Equal(err,nil)
    at.Equal(ret.IsVideo,true)
    h = ret.Header.(av.VideoPacketHeader)
    at.Equal(h.IsKeyFrame(),true)
}
