package pool

type Pool struct {
	buf []byte
}

const maxpoolsize = 2 * 1024 * 1024

func (self *Pool) Get(size int) []byte {
	if len(self.buf) < size {
		self.buf = make([]byte, maxpoolsize)
	}
	b := self.buf[:size]
	self.buf = self.buf[size:]
	return b
}

func NewPool() *Pool {
	return &Pool{
		buf: make([]byte, maxpoolsize),
	}
}
