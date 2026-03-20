package stats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/hashicorp/go-multierror"
	"github.com/jackc/pgx/v5"

	"github.com/dmtaylor/costanza/internal/cache"
	"github.com/dmtaylor/costanza/internal/model"
)

type CursedHandler struct {
	pool               model.DbPool
	cursedWordCache    *cache.PgxStringListCache
	cursedChannelCache *cache.DbChannelCache
}

func NewCursedHandler(pool model.DbPool) *CursedHandler {
	return &CursedHandler{
		pool:               pool,
		cursedWordCache:    cache.NewPgxStringListCache(pool),
		cursedChannelCache: cache.NewDbChannelCache(pool),
	}
}

func (ch *CursedHandler) PreloadCache(ctx context.Context, guildIds []uint64) error {
	errGroup := multierror.Group{}
	errGroup.Go(func() error {
		return ch.cursedWordCache.PreloadCache(ctx, guildIds)
	})
	errGroup.Go(func() error {
		return ch.cursedChannelCache.PreloadCache(ctx, guildIds)
	})

	return errGroup.Wait().ErrorOrNil()
}

func (ch *CursedHandler) GetCursedWordsList(ctx context.Context, guildId uint64) ([]string, error) {
	return ch.cursedWordCache.Get(ctx, guildId)
}

func (ch *CursedHandler) GetCursedChannelList(ctx context.Context, guildId uint64) ([]uint64, error) {
	return ch.cursedChannelCache.Get(ctx, guildId)
}

func (ch *CursedHandler) AddWordToCursedList(ctx context.Context, guildId uint64, word string) error {
	var wordPresent bool
	tx, err := ch.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rollback transaction: "+err.Error())
		}
	}(tx, ctx)
	err = pgxscan.Select(ctx, tx, &wordPresent, "SELECT EXISTS(SELECT 1 FROM cursed_word_list WHERE guild_id = $1 AND word = $2)", guildId, word)
	if err != nil {
		return fmt.Errorf("failed to check word existence: %w", err)
	}
	if wordPresent {
		return nil
	}
	_, err = tx.Exec(ctx, "INSERT INTO cursed_word_list(guild_id, word) VALUES ($1, $2)", guildId, word)
	if err != nil {
		return fmt.Errorf("failed to add word to cursed list: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	ch.cursedWordCache.InvalidateKey(ctx, guildId)
	return nil
}

func (ch *CursedHandler) RemoveWordFromCursedList(ctx context.Context, guildId uint64, word string) error {
	tx, err := ch.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rollback transaction: "+err.Error())
		}
	}(tx, ctx)
	result, err := tx.Exec(ctx, "DELETE FROM cursed_word WHERE guild_id = $1 AND word = $2", guildId, word)
	if err != nil {
		return fmt.Errorf("failed to remove word from cursed list: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil
	}
	ch.cursedWordCache.InvalidateKey(ctx, guildId)
	return nil
}
