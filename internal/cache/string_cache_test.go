package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPgxStringListCache(t *testing.T) {
	cache := NewPgxStringListCache(nil)
	expected := &PgxStringListCache{
		pool:      nil,
		cache:     make(map[uint64]stringListCacheItem),
		cacheLock: sync.RWMutex{},
		updating:  sync.Mutex{},
	}
	assert.Equal(t, expected, cache, "bad default cache")
}

func TestPgxStringListCache_Clear(t *testing.T) {
	cache := NewPgxStringListCache(nil)
	cache.Set(context.Background(), uint64(5), []string{"joan"})
	cache.Clear(context.Background())
	assert.Len(t, cache.cache, 0, "cache not empty")
	if _, ok := cache.cache[5]; ok {
		t.Error("found entry that should be gone")
	}
}

func TestPgxStringListCache_Set(t *testing.T) {
	cache := NewPgxStringListCache(nil)
	expectedExpiry := time.Now().Add(defaultEntryDuration)
	cache.Set(context.Background(), 1, []string{"peter", "joe"})
	if item, ok := cache.cache[1]; ok {
		assert.Equal(t, []string{"peter", "joe"}, item.value)
		assert.WithinDuration(t, expectedExpiry, item.expiry, time.Millisecond, "expiry way off")
	} else {
		t.Error("item missing from cache")
	}
}

func TestPgxStringListCache_GetCacheHit(t *testing.T) {
	cache := NewPgxStringListCache(nil)
	cache.cache[5] = stringListCacheItem{value: []string{"alan", "merrel"}, expiry: time.Now().Add(time.Minute * 15)}
	got, err := cache.Get(context.Background(), 5)
	if assert.NoError(t, err, "got unexpected error") {
		assert.Equal(t, []string{"alan", "merrel"}, got, "result mismatch")
	}
}

func TestPgxStringListCache_GetCacheMiss(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	var testKey uint64 = 100
	expectedResult := []string{"jackie", "bette"}
	rows := pgxmock.NewRows([]string{"word"}).
		AddRow("jackie").
		AddRow("bette")
	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).
		WithArgs(testKey).
		WillReturnRows(rows)
	cache := NewPgxStringListCache(mockDb)
	expectedExpiryEstimate := time.Now().Add(time.Minute * 15)
	got, err := cache.Get(context.Background(), testKey)
	if assert.NoError(t, err, "got error") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		assert.Equal(t, expectedResult, got, "mismatch results")
		item := cache.cache[testKey]
		assert.WithinDuration(t, expectedExpiryEstimate, item.expiry, time.Millisecond, "expiry way off")
		assert.Equal(t, expectedResult, item.value, "correct value not stored")
	}
}

func TestPgxStringListCache_GetExpired(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	var testKey uint64 = 101
	expectedResult := []string{"debbie", "steve", "sophie"}
	rows := pgxmock.NewRows([]string{"word"}).
		AddRow("debbie").
		AddRow("steve").
		AddRow("sophie")
	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).
		WithArgs(testKey).
		WillReturnRows(rows)
	cache := NewPgxStringListCache(mockDb)
	cache.cache[testKey] = stringListCacheItem{
		value:  []string{"debbie", "steve"},
		expiry: time.Now().Add(time.Minute * -1),
	}
	expectedExpiryEstimate := time.Now().Add(time.Minute * 15)
	got, err := cache.Get(context.Background(), testKey)
	if assert.NoError(t, err, "got error") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		assert.Equal(t, expectedResult, got, "mismatch results")
		item := cache.cache[testKey]
		assert.WithinDuration(t, expectedExpiryEstimate, item.expiry, time.Millisecond, "expiry way off")
		assert.NotEqual(t, []string{"debbie", "steve"}, item.value, "old entry not replaced")
		assert.Equal(t, expectedResult, item.value, "correct value not stored")
	}
}

func TestPgxStringListCache_PreloadCache(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build pool")
	defer mockDb.Close()

	guildIds := []uint64{4, 5, 6}
	guild1Expected := []string{"mimi", "steve"}
	var guild2Expected []string
	guild3Expected := []string{"marge"}

	guild1Rows := pgxmock.NewRows([]string{"word"}).
		AddRow("mimi").
		AddRow("steve")
	guild2Rows := pgxmock.NewRows([]string{"word"})
	guild3Rows := pgxmock.NewRows([]string{"word"}).
		AddRow("marge")

	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).
		WithArgs(uint64(4)).
		WillReturnRows(guild1Rows)
	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).
		WithArgs(uint64(5)).
		WillReturnRows(guild2Rows)
	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).
		WithArgs(uint64(6)).
		WillReturnRows(guild3Rows)
	cache := NewPgxStringListCache(mockDb)
	expectedExpiryEstimate := time.Now().Add(time.Minute * 15)
	err = cache.PreloadCache(context.Background(), guildIds)
	if assert.NoError(t, err, "got error") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet expectations") {
		cacheItem1 := cache.cache[4]
		assert.Equal(t, guild1Expected, cacheItem1.value, "item 1 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem1.expiry, time.Millisecond, "expiry 1 drift")
		cacheItem2 := cache.cache[5]
		assert.Equal(t, guild2Expected, cacheItem2.value, "item 2 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem2.expiry, time.Millisecond, "expiry 2 drift")
		cacheItem3 := cache.cache[6]
		assert.Equal(t, guild3Expected, cacheItem3.value, "item 3 value mismatch")
		assert.WithinDuration(t, expectedExpiryEstimate, cacheItem3.expiry, time.Millisecond, "expiry 3 drift")
	}
}
