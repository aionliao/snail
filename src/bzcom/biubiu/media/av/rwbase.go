package av

import "time"

type RWBaser struct {
	timeout            time.Duration
	PreTime            time.Time
	BaseTimestamp      uint32
	LastVideoTimestamp uint32
	LastAudioTimestamp uint32
}

func NewRWBaser(duration time.Duration) RWBaser {
	return RWBaser{
		timeout: duration,
		PreTime: time.Now(),
	}
}

func (self *RWBaser) BaseTimeStamp() uint32 {
	return self.BaseTimestamp
}

func (self *RWBaser) CalcBaseTimestamp() {
	if self.LastAudioTimestamp > self.LastVideoTimestamp {
		self.BaseTimestamp = self.LastAudioTimestamp
	} else {
		self.BaseTimestamp = self.LastVideoTimestamp
	}
}

func (self *RWBaser) RecTimeStamp(timestamp, typeID uint32) {
	if typeID == TAG_VIDEO {
		self.LastVideoTimestamp = timestamp
	} else if typeID == TAG_AUDIO {
		self.LastAudioTimestamp = timestamp
	}
}

func (self *RWBaser) SetPreTime() {
	self.PreTime = time.Now()
}

func (self *RWBaser) Alive() bool {
	return !(time.Now().Sub(self.PreTime) >= self.timeout)
}
