package listen

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dmtaylor/costanza/internal/model"
)

func Test_createGameResult(t *testing.T) {
	type args struct {
		guildId  uint64
		userId   uint64
		gameType string
		message  string
	}
	tests := []struct {
		name          string
		args          args
		want          model.DailyGamePlay
		expectedError error
	}{
		{
			"invalid_game_type",
			args{
				4,
				5,
				"invalid",
				"this is an invalid game type",
			},
			model.DailyGamePlay{},
			errors.New("invalid game type: invalid"),
		},
		{
			"framed_win",
			args{
				101,
				102,
				"Framed",
				`Framed #535
🎥 🟥 🟥 🟩 ⬛ ⬛ ⬛

https://framed.wtf/`,
			},
			model.DailyGamePlay{
				101,
				102,
				3,
				true,
			},
			nil,
		},
		{
			"framed_loss",
			args{
				111,
				112,
				"Framed",
				`Framed #566
🎥 🟥 🟥 🟥 🟥 🟥 🟥

https://framed.wtf/`,
			},
			model.DailyGamePlay{
				111,
				112,
				6,
				false,
			},
			nil,
		},
		{
			"GuessTheGame_win",
			args{
				201,
				202,
				"GuessTheGame",
				`#GuessTheGame #477

🎮 🟩 ⬜ ⬜ ⬜ ⬜ ⬜

#ScreenshotSleuth
https://guessthe.game/`,
			},
			model.DailyGamePlay{
				201,
				202,
				1,
				true,
			},
			nil,
		},
		{
			"wordle_win",
			args{
				301,
				302,
				"Wordle",
				`Wordle 559 2/6

⬛🟨🟩⬛⬛
🟩🟩🟩🟩🟩`,
			},
			model.DailyGamePlay{
				301,
				302,
				2,
				true,
			},
			nil,
		},
		{
			"wordle_loss",
			args{
				401,
				402,
				"Wordle",
				`Wordle 576 X/6

⬛⬛⬛🟨⬛
⬛⬛⬛⬛🟨
⬛🟩🟩🟩🟩
⬛🟩🟩🟩🟩
⬛🟩🟩🟩🟩
⬛🟩🟩🟩🟩`,
			},
			model.DailyGamePlay{
				401,
				402,
				6,
				false,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createGameResult(tt.args.guildId, tt.args.userId, tt.args.gameType, tt.args.message)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func Test_isGameMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			"empty",
			"",
			false,
		},
		{
			"regular_message",
			"this is a normal message someone would send",
			false,
		},
		{
			"framed",
			`Framed #541
🎥 🟥 🟥 🟥 🟥 🟥 🟥

https://framed.wtf/`,
			true,
		},
		{
			"guessTheGame",
			`#GuessTheGame #477

🎮 🟥 🟥 🟩 ⬜ ⬜ ⬜

#GameNavigator
https://guessthe.game/`,
			true,
		},
		{
			"tradle",
			`#Tradle (🇺🇸 Edition) #278 4/6
🟩🟩🟩🟩⬜
🟩🟩🟩🟩🟨
🟩🟩🟩🟩🟨
🟩🟩🟩🟩🟩
https://oec.world/en/tradle`,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isGameMessage(tt.message), "isGameMessage(%v)", tt.message)
		})
	}
}
