package quotes

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
)

//go:embed quotes.json
var quotesData []byte

type QuoteEngine struct {
	quoteList []string
	rng       *rand.Rand
	lock      sync.Mutex
}

func NewQuoteEngine() (*QuoteEngine, error) {
	engine := QuoteEngine{
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
		lock: sync.Mutex{},
	}
	err := json.Unmarshal(quotesData, &engine.quoteList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse quotes")
	}

	return &engine, nil
}

func (q *QuoteEngine) GetQuote() string {
	q.lock.Lock()
	idx := q.rng.Intn(len(q.quoteList))
	q.lock.Unlock()
	return q.quoteList[idx]
}
