package cache

import (
	"context"
	"time"
)

const maxSize = 20
const defaultEntryDuration = time.Minute * 15

type PreloadableCache interface {
	PreloadCache(ctx context.Context, guildIds []uint64) error
}
