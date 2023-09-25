package listen

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
		want          DailyGamePlay
		expectedError error
	}{
		// TODO: Add test cases.
		{
			"invalid_game_type",
			args{
				4,
				5,
				"invalid",
				"this is an invalid game type",
			},
			DailyGamePlay{},
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
			DailyGamePlay{
				101,
				102,
				3,
				true,
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
			DailyGamePlay{
				201,
				202,
				1,
				true,
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
		// TODO add
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
