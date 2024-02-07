package roller

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type BaseRoller struct {
	rng  *rand.Rand
	lock *sync.Mutex
}

// NewBaseRoller creates basic roller with default rng
func NewBaseRoller() (*BaseRoller, error) {
	buf := make([]byte, 8)
	_, err := crand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to get rand seed: %w", err)
	}
	return newRoller(uint64(time.Now().UnixNano()), binary.NativeEndian.Uint64(buf)), nil
}

// NewTestBaseRoller creates basic roller with pinned seed for predictable results for testing
func NewTestBaseRoller(seed1, seed2 uint64) *BaseRoller {
	return newRoller(seed1, seed2)
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
	return r.rng.IntN(base) + 1
}

func newRoller(seed1, seed2 uint64) *BaseRoller {
	src := rand.NewPCG(seed1, seed2)
	return &BaseRoller{
		rng:  rand.New(src),
		lock: &sync.Mutex{},
	}
}
