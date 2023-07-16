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

// NewBaseRoller creates basic roller with default rng
func NewBaseRoller() *BaseRoller {
	return newRoller(uint64(time.Now().UnixNano()))
}

// NewTestBaseRoller creates basic roller with pinned seed for predictable results for testing
func NewTestBaseRoller(seed uint64) *BaseRoller {
	return newRoller(seed)
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

func newRoller(seed uint64) *BaseRoller {
	src := &rand.PCGSource{}
	src.Seed(seed)
	return &BaseRoller{
		rng:  rand.New(src),
		lock: &sync.Mutex{},
	}
}
