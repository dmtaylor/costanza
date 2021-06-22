package roller

import (
	"math/rand"
	"sync"
	"time"
)

type BaseRoller struct {
	rng  *rand.Rand
	lock *sync.Mutex
}

func New() *BaseRoller {
	return &BaseRoller{
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
		lock: &sync.Mutex{},
	}
}

func (r *BaseRoller) DoRoll(num int, base int) BaseRoll {
	result := make(BaseRoll, num)
	for i := 0; i < num; i++ {
		result = append(result, r.getRoll(base))
	}
	return result
}

func (r *BaseRoller) getRoll(base int) int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.rng.Intn(base) + 1
}
