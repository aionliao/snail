package aac

import (
	"bzcom/media/container/flv"
	"errors"
	"io"
)

type mpegExtension struct {
	objectType byte
	sampleRate byte
}

type mpegCfgInfo struct {
	objectType     byte
	sampleRate     byte
	channel        byte
	sbr            byte
	ps             byte
	frameLen       byte
	exceptionLogTs int64
	extension      *mpegExtension
}

var aacRates = []int{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}

var (
	specificBufInvalid = errors.New("audio mpegspecific error")
	audioBufInvalid    = errors.New("audiodata  invalid")
)

const (
	adtsHeaderLen = 7
)

type Demuxer struct {
	gettedSpecific bool
	adtsHeader     []byte
	cfgInfo        *mpegCfgInfo
}

func NewDemuxer() *Demuxer {
	return &Demuxer{
		gettedSpecific: false,
		cfgInfo:        &mpegCfgInfo{},
		adtsHeader:     make([]byte, adtsHeaderLen),
	}
}

func (self *Demuxer) specificInfo(src []byte) error {
	if len(src) < 2 {
		return specificBufInvalid
	}
	self.gettedSpecific = true
	self.cfgInfo.objectType = (src[0] >> 3) & 0xff
	self.cfgInfo.sampleRate = ((src[0] & 0x07) << 1) | src[1]>>7
	self.cfgInfo.channel = (src[1] >> 3) & 0x0f

	return nil
}

func (self *Demuxer) adts(src []byte, w io.Writer) error {
	if len(src) <= 0 || !self.gettedSpecific {
		return audioBufInvalid
	}

	frameLen := uint16(len(src)) + 7

	//first write adts header
	self.adtsHeader[0] = 0xff
	self.adtsHeader[1] = 0xf1

	self.adtsHeader[2] &= 0x00
	self.adtsHeader[2] = self.adtsHeader[2] | (self.cfgInfo.objectType-1)<<6
	self.adtsHeader[2] = self.adtsHeader[2] | (self.cfgInfo.sampleRate)<<2

	self.adtsHeader[3] &= 0x00
	self.adtsHeader[3] = self.adtsHeader[3] | (self.cfgInfo.channel<<2)<<4
	self.adtsHeader[3] = self.adtsHeader[3] | byte((frameLen<<3)>>14)

	self.adtsHeader[4] &= 0x00
	self.adtsHeader[4] = self.adtsHeader[4] | byte((frameLen<<5)>>8)

	self.adtsHeader[5] &= 0x00
	self.adtsHeader[5] = self.adtsHeader[5] | byte(((frameLen<<13)>>13)<<5)
	self.adtsHeader[5] = self.adtsHeader[5] | (0x7C<<1)>>3
	self.adtsHeader[6] = 0xfc

	if _, err := w.Write(self.adtsHeader[0:]); err != nil {
		return err
	}
	if _, err := w.Write(src); err != nil {
		return err
	}
	return nil
}

func (self *Demuxer) SampleRate() int {
	rate := 44100
	if self.cfgInfo.sampleRate <= byte(len(aacRates)-1) {
		rate = aacRates[self.cfgInfo.sampleRate]
	}
	return rate
}

func (self *Demuxer) Demux(tag *flv.Tag, w io.Writer) (err error) {
	switch tag.MT.AACPacketType {
	case flv.AAC_SEQHDR:
		err = self.specificInfo(tag.Data)
	case flv.AAC_RAW:
		err = self.adts(tag.Data, w)
	}
	return
}
