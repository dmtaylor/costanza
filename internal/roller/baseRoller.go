package roller

import (
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type BaseRoller struct {
	rng  *rand.Rand
	lock *sync.Mutex
}

func NewBaseRoller() *BaseRoller {
	src := &rand.PCGSource{}
	src.Seed(uint64(time.Now().UnixNano()))
	return &BaseRoller{
		rng:  rand.New(src),
		lock: &sync.Mutex{},
	}
}

func (r *BaseRoller) DoRoll(num int, base int) BaseRoll {
	result := make(BaseRoll, num)
	for i := 0; i < num; i++ {
		result[i] = r.getRoll(base)
	}
	return result
}

func (r *BaseRoller) getRoll(base int) int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.rng.Intn(base) + 1
}
