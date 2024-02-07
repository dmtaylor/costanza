package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestDbChannelCache_GetCacheMiss(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	var testKey uint64 = 90
	expectedResult := []uint64{5, 10, 15}
	rows := pgxmock.NewRows([]string{"channel_id"}).
		AddRow(uint64(5)).
		AddRow(uint64(10)).
		AddRow(uint64(15))
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).
		WithArgs(testKey).
		WillReturnRows(rows)
	cache := NewDbChannelCache(mockDb)
	expectedExpiryEstimate := time.Now().Add(defaultEntryDuration)
	got, err := cache.Get(context.Background(), testKey)
	if assert.NoError(t, err, "got error") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		assert.Equal(t, expectedResult, got, "mismatch results")
		item := cache.cache[testKey]
		assert.WithinDuration(t, expectedExpiryEstimate, item.expiry, time.Millisecond, "expiry way off")
		assert.Equal(t, expectedResult, item.value, "correct value not stored")
	}
}

func TestDbChannelCache_GetExpiredEntry(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	var testKey uint64 = 90
	expectedResult := []uint64{5, 10, 15}
	rows := pgxmock.NewRows([]string{"channel_id"}).
		AddRow(uint64(5)).
		AddRow(uint64(10)).
		AddRow(uint64(15))
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).
		WithArgs(testKey).
		WillReturnRows(rows)
	cache := NewDbChannelCache(mockDb)
	cache.cache[testKey] = channelCacheItem{
		value:  []uint64{5, 10},
		expiry: time.Now().Add(time.Minute * -1),
	}
	expectedExpiryEstimate := time.Now().Add(defaultEntryDuration)
	got, err := cache.Get(context.Background(), testKey)
	if assert.NoError(t, err, "got error") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		assert.Equal(t, expectedResult, got, "mismatch results")
		item := cache.cache[testKey]
		assert.WithinDuration(t, expectedExpiryEstimate, item.expiry, time.Millisecond, "expiry way off")
		assert.NotEqual(t, []uint64{5, 10}, item.value, "old entry not replaced")
		assert.Equal(t, expectedResult, item.value, "correct value not stored")
	}
}

func TestDbChannelCache_PreloadCache(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	guildIds := []uint64{1, 2, 3}
	guild1Expected := []uint64{5, 10}
	guild2Expected := []uint64{10, 20}
	guild3Expected := []uint64{15, 30}
	guild1Rows := pgxmock.NewRows([]string{"channel_id"}).
		AddRow(uint64(5)).
		AddRow(uint64(10))
	guild2Rows := pgxmock.NewRows([]string{"channel_id"}).
		AddRow(uint64(10)).
		AddRow(uint64(20))
	guild3Rows := pgxmock.NewRows([]string{"channel_id"}).
		AddRow(uint64(15)).
		AddRow(uint64(30))
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(guild1Rows)
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).
		WithArgs(uint64(2)).
		WillReturnRows(guild2Rows)
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).
		WithArgs(uint64(3)).
		WillReturnRows(guild3Rows)
	expectedExpiryEstimate := time.Now().Add(defaultEntryDuration)
	cache := NewDbChannelCache(mockDb)
	err = cache.PreloadCache(context.Background(), guildIds)
	if assert.NoError(t, err, "got error preloading cache") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		cacheItem1 := cache.cache[1]
		assert.Equal(t, guild1Expected, cacheItem1.value, "item 1 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem1.expiry, time.Millisecond, "expiry 1 drift")
		cacheItem2 := cache.cache[2]
		assert.Equal(t, guild2Expected, cacheItem2.value, "item 2 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem2.expiry, time.Millisecond, "expiry 2 drift")
		cacheItem3 := cache.cache[3]
		assert.Equal(t, guild3Expected, cacheItem3.value, "item 3 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem3.expiry, time.Millisecond, "expiry 3 drift")
	}
}
