package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dmtaylor/costanza/internal/model"
)

func TestNewDbChannelCache(t *testing.T) {
	type args struct {
		pool model.DbPool
	}
	tests := []struct {
		name string
		args args
		want *DbChannelCache
	}{
		{
			"basic",
			args{
				nil,
			},
			&DbChannelCache{
				pool:      nil,
				cache:     make(map[uint64]channelCacheItem),
				cacheLock: sync.RWMutex{},
				updating:  sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDbChannelCache(tt.args.pool))
		})
	}
}

func TestDbChannelCache_Clear(t *testing.T) {
	cache := NewDbChannelCache(nil)
	cache.Set(context.Background(), uint64(5), []uint64{5, 10})
	assert.Len(t, cache.cache, 1, "not added")
	cache.Clear(context.Background())
	assert.Len(t, cache.cache, 0, "cache not empty")
	if _, ok := cache.cache[5]; ok {
		t.Error("found entry that should be gone")
	}
}

func TestDbChannelCache_Set(t *testing.T) {
	cache := NewDbChannelCache(nil)
	expectedExpiry := time.Now().Add(defaultEntryDuration)
	cache.Set(context.Background(), uint64(6), []uint64{5, 10})
	if item, ok := cache.cache[6]; ok {
		assert.Equal(t, []uint64{5, 10}, item.value)
		assert.WithinDuration(t, expectedExpiry, item.expiry, time.Millisecond, "expiry way off")
	} else {
		t.Error("item missing from cache")
	}
}

func TestDbChannelCache_GetCacheHit(t *testing.T) {
	cache := NewDbChannelCache(nil)
	cache.cache[5] = channelCacheItem{value: []uint64{5, 10}, expiry: time.Now().Add(time.Minute * 15)}
	got, err := cache.Get(context.Background(), 5)
	if assert.NoError(t, err, "got unexpected error") {
		assert.Equal(t, []uint64{5, 10}, got, "result mismatch")
	}
}

func TestDbChannelCache_ConcurrentGet(t *testing.T) {
	cache := NewDbChannelCache(nil)
	cache.cache[1] = channelCacheItem{
		value:  []uint64{1, 2},
		expiry: time.Now().Add(time.Minute * 15),
	}
	cache.cache[2] = channelCacheItem{
		value:  []uint64{3, 4},
		expiry: time.Now().Add(time.Minute * 15),
	}
	var got1, got2 []uint64
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		got1, err1 = cache.Get(context.Background(), 1)
	}()
	go func() {
		defer wg.Done()
		got2, err2 = cache.Get(context.Background(), 2)
	}()
	wg.Wait()
	if assert.NoError(t, err1, "got inner error 1") {
		assert.Equal(t, []uint64{1, 2}, got1, "mismatch results 1")
	}
	if assert.NoError(t, err2, "got inner error 2") {
		assert.Equal(t, []uint64{3, 4}, got2, "mismatch results 2")
	}
}
