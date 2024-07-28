package wire

import (
	"math"
	"sync/atomic"
)

//生成序列号

type sequence struct {
	num uint32
}

func (s *sequence) Next() uint32 {
	next := atomic.AddUint32(&s.num, 1)
	if next == math.MaxUint32 {
		if atomic.CompareAndSwapUint32(&s.num, next, 1) {
			return s.Next()
		}
	}
	return next
}

var Seq = &sequence{num: 1}
