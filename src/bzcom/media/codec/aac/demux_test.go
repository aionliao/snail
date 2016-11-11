package aac

import (
	"bytes"
	"bzcom/media/container/flv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAACDemux(t *testing.T) {
	at := assert.New(t)
	d := NewDemuxer()
	tag := flv.Tag{
		MT: flv.MediaTag{
			AACPacketType: flv.AAC_SEQHDR,
		},
		Data: []byte{0x11, 0x88, 0x56, 0xe5, 0x00},
	}
	w := bytes.NewBuffer(nil)
	err := d.Demux(&tag, w)
	at.Equal(err, nil)
	at.Equal(w.Len(), 0)
	at.Equal(int(d.cfgInfo.channel), 1)
	at.Equal(int(d.cfgInfo.sampleRate), 3)

	tag = flv.Tag{
		MT: flv.MediaTag{
			AACPacketType: flv.AAC_RAW,
		},
		Data: []byte{0x11, 0x88, 0x56, 0xe5, 0x00},
	}
	err = d.Demux(&tag, w)
	at.Equal(err, nil)
	at.Equal(w.Len(), 12)
}
