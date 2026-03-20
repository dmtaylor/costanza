package stats

import (
	"context"
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmtaylor/costanza/internal/cache"
	"github.com/dmtaylor/costanza/internal/model"
)

func TestCursedHandler_PreloadCache(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to create pool")
	defer mockDb.Close()
	mockDb.MatchExpectationsInOrder(false)

	guildIds := []uint64{1}
	guild1ExpectedWords := []string{"fozzie", "ratso"}
	guild1WordRows := pgxmock.NewRows([]string{"words"}).AddRow("fozzie").AddRow("ratso")
	mockDb.ExpectQuery(`SELECT word FROM cursed_word_list WHERE guild_id = \$1`).WithArgs(uint64(1)).WillReturnRows(guild1WordRows)
	guild1ExpectedChannels := []uint64{500}
	guild1ChannelRows := pgxmock.NewRows([]string{"channel_id"}).AddRow(uint64(500))
	mockDb.ExpectQuery(`SELECT channel_id FROM cursed_channels WHERE guild_id = \$1`).WithArgs(uint64(1)).WillReturnRows(guild1ChannelRows)

	handler := NewCursedHandler(mockDb)
	err = handler.PreloadCache(context.Background(), guildIds)
	if assert.NoError(t, err, "error preloading cache") && assert.NoError(t, mockDb.ExpectationsWereMet(), "unmet db expectations") {
		wordCacheList, err := handler.GetCursedWordsList(context.Background(), guildIds[0])
		assert.NoError(t, err, "error getting words cache")
		assert.ElementsMatch(t, guild1ExpectedWords, wordCacheList, "mismatched words")

		channelCacheList, err := handler.GetCursedChannelList(context.Background(), guildIds[0])
		assert.NoError(t, err, "error getting channels cache")
		assert.ElementsMatch(t, guild1ExpectedChannels, channelCacheList, "mismatched channels")
	}
}

func TestCursedHandler_RemoveWordFromCursedList(t *testing.T) {
	type fields struct {
		pool               model.DbPool
		cursedWordCache    *cache.PgxStringListCache
		cursedChannelCache *cache.DbChannelCache
	}
	type args struct {
		ctx     context.Context
		guildId uint64
		word    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &CursedHandler{
				pool:               tt.fields.pool,
				cursedWordCache:    tt.fields.cursedWordCache,
				cursedChannelCache: tt.fields.cursedChannelCache,
			}
			tt.wantErr(t, ch.RemoveWordFromCursedList(tt.args.ctx, tt.args.guildId, tt.args.word), fmt.Sprintf("RemoveWordFromCursedList(%v, %v, %v)", tt.args.ctx, tt.args.guildId, tt.args.word))
		})
	}
}

func TestNewCursedHandler(t *testing.T) {
	handler := NewCursedHandler(nil)
	expectedHandler := &CursedHandler{
		pool:               nil,
		cursedWordCache:    cache.NewPgxStringListCache(nil),
		cursedChannelCache: cache.NewDbChannelCache(nil),
	}
	assert.Equalf(t, expectedHandler, handler, "bad handler")
}
