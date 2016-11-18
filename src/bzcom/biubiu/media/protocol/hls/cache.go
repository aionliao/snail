package hls

import "bytes"

const (
	cache_max_frames byte = 6
	audio_cache_len  int  = 10 * 1024
)

type AudioCache struct {
	soundFormat byte
	num         byte
	offset      int
	pts         uint64
	buf         *bytes.Buffer
}

func NewAudioCache() *AudioCache {
	return &AudioCache{
		buf: bytes.NewBuffer(make([]byte, audio_cache_len)),
	}
}

func (a *AudioCache) Cache(src []byte, pts uint64) bool {
	if a.num == 0 {
		a.offset = 0
		a.pts = pts
		a.buf.Reset()
	}
	a.buf.Write(src)
	a.offset += len(src)
	a.num++

	return false
}

func (a *AudioCache) GetFrame() (int, uint64, []byte) {
	a.num = 0
	return a.offset, a.pts, a.buf.Bytes()
}

func (a *AudioCache) CacheNum() byte {
	return a.num
}
