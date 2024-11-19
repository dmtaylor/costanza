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

type ChannelCache interface {
	Get(ctx context.Context, key uint64) ([]uint64, error)
	Set(ctx context.Context, key uint64, value []uint64)
	Clear(ctx context.Context)
}

type DbChannelCache struct {
	pool      model.DbPool
	cache     map[uint64]channelCacheItem
	cacheLock sync.RWMutex
	updating  sync.Mutex
}

type channelCacheItem struct {
	value  []uint64
	expiry time.Time
}

func NewDbChannelCache(pool model.DbPool) *DbChannelCache {
	return &DbChannelCache{
		pool:  pool,
		cache: make(map[uint64]channelCacheItem),
	}
}

func (c *DbChannelCache) Clear(_ context.Context) {
	c.cache = make(map[uint64]channelCacheItem)
}

func (c *DbChannelCache) Set(_ context.Context, key uint64, value []uint64) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	c.cache[key] = channelCacheItem{
		value:  value,
		expiry: time.Now().Add(defaultEntryDuration),
	}
}

func (c *DbChannelCache) Get(ctx context.Context, key uint64) ([]uint64, error) {
	c.cacheLock.RLock()
	if item, ok := c.cache[key]; ok && time.Now().Before(item.expiry) {
		c.cacheLock.RUnlock()
		return item.value, nil
	} else {
		c.cacheLock.RUnlock()
		c.updating.Lock()
		defer c.updating.Unlock()
		if item, ok := c.cache[key]; !ok || time.Now().After(item.expiry) {
			slog.DebugContext(ctx, "refreshing cache for key "+strconv.FormatUint(key, 10))
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

func (c *DbChannelCache) PreloadCache(ctx context.Context, guildIds []uint64) error {
	var err *multierror.Error
	for _, guildId := range guildIds {
		dbValues, e := c.fetchDbValues(ctx, guildId)
		if e != nil {
			err = multierror.Append(err, e)
			continue
		}
		c.Set(ctx, guildId, dbValues)
	}

	if err == nil {
		return nil
	}
	return err.ErrorOrNil()
}

func (c *DbChannelCache) fetchDbValues(ctx context.Context, key uint64) ([]uint64, error) {
	var results []uint64
	err := pgxscan.Select(ctx, c.pool, &results, "SELECT channel_id FROM cursed_channels WHERE guild_id = $1", key)
	if err != nil {
		return nil, fmt.Errorf("failed to query cursed channels: %w", err)
	}

	return results, nil
}
