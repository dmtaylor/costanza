package cache

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/hashicorp/go-multierror"

	"github.com/dmtaylor/costanza/internal/model"
)

type StringListCache interface {
	Get(ctx context.Context, key uint64) ([]string, error)
	Set(ctx context.Context, key uint64, value []string)
	Clear(ctx context.Context)
}

type PgxStringListCache struct {
	pool      model.DbPool
	cache     map[uint64]stringListCacheItem
	cacheLock sync.RWMutex
	updating  sync.Mutex
}

type stringListCacheItem struct {
	value  []string
	expiry time.Time
}

func NewPgxStringListCache(pool model.DbPool) *PgxStringListCache {
	return &PgxStringListCache{
		pool:  pool,
		cache: make(map[uint64]stringListCacheItem),
	}
}

func (c *PgxStringListCache) Clear(_ context.Context) {
	c.cache = make(map[uint64]stringListCacheItem)
}

func (c *PgxStringListCache) Set(_ context.Context, key uint64, value []string) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	c.cache[key] = stringListCacheItem{
		value:  value,
		expiry: time.Now().Add(defaultEntryDuration),
	}
}

func (c *PgxStringListCache) Get(ctx context.Context, key uint64) ([]string, error) {
	c.cacheLock.RLock()
	if item, ok := c.cache[key]; ok && time.Now().Before(item.expiry) {
		c.cacheLock.RUnlock()
		return item.value, nil
	} else {
		c.cacheLock.RUnlock()
		c.updating.Lock()
		defer c.updating.Unlock()
		if item, ok := c.cache[key]; !ok || time.Now().After(item.expiry) {
			slog.DebugContext(ctx, "refreshing str list cache for key "+strconv.FormatUint(key, 10))
			dbValues, err := c.fetchDbValues(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("failed to pull into cache: %w", err)
			}
			c.Set(ctx, key, dbValues)
			return dbValues, nil
		} else {
			return item.value, nil
		}
	}
}

func (c *PgxStringListCache) PreloadCache(ctx context.Context, guildIds []uint64) error {
	var err *multierror.Error
	for _, guildId := range guildIds {
		dbValues, e := c.fetchDbValues(ctx, guildId)
		if e != nil {
			err = multierror.Append(err, e)
			continue
		}
		c.Set(ctx, guildId, dbValues)
	}

	return err.ErrorOrNil()
}

func (c *PgxStringListCache) fetchDbValues(ctx context.Context, key uint64) ([]string, error) {
	var results []string
	// hardcode query here, make more general if we need it in the future
	err := pgxscan.Select(ctx, c.pool, &results, "SELECT word FROM cursed_word_list WHERE guild_id = $1", key)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursed word list: %w", err)
	}
	return results, nil
}
