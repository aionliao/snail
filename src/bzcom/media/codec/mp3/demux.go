package mp3

import "errors"

type Demuxer struct {
	samplingFrequency int
}

func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

// sampling_frequency - indicates the sampling frequency, according to the following table.
// '00' 44.1 kHz
// '01' 48 kHz
// '10' 32 kHz
// '11' reserved
var mp3Rates = []int{44100, 48000, 32000}
var (
	errMp3DataInvalid = errors.New("mp3data  invalid")
	errIndexInvalid   = errors.New("invalid rate index")
)

func (self *Demuxer) Demux(src []byte) error {
	if len(src) < 3 {
		return errMp3DataInvalid
	}
	index := (src[2] >> 2) & 0x3
	if index <= len(mp3Rates)-1 {
		self.samplingFrequency = mp3Rates[index]
		return nil
	}
	return errIndexInvalid
}

func (self *Demuxer) SampleRate() int {
	if self.samplingFrequency == 0 {
		self.samplingFrequency = 44100
	}
	return self.samplingFrequency
}
