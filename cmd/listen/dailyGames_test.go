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
			name: "framed_win",
			args: args{
				101,
				102,
				"Framed",
				`Framed #535
🎥 🟥 🟥 🟩 ⬛ ⬛ ⬛

https://framed.wtf/`,
			},
			want: model.DailyGamePlay{
				GuildId: 101,
				UserId:  102,
				Tries:   3,
				Win:     true,
			},
			expectedError: nil,
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
				GuildId: 111,
				UserId:  112,
				Tries:   6,
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
				GuildId: 201,
				UserId:  202,
				Tries:   1,
				Win:     true,
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
				GuildId: 301,
				UserId:  302,
				Tries:   2,
				Win:     true,
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
				GuildId: 401,
				UserId:  402,
				Tries:   6,
			},
			nil,
		},
		{
			"flashback_win",
			args{
				501,
				502,
				"Flashback",
				`Flashback for October 29, 2023

21 points
🟩🟩🟩🟥🟩🟩🟩🟥

Play here: 
        https://www.nytimes.com/interactive/2023/10/27/upshot/flashback.html?hide-chrome=1`,
			},
			model.DailyGamePlay{
				GuildId: 501,
				UserId:  502,
				Tries:   3,
				Win:     true,
			},
			nil,
		},
		{
			"guessTheGame_yellow_square",
			args{
				guildId:  601,
				userId:   602,
				gameType: "GuessTheGame",
				message:  "#GuessTheGame #548\n\n🎮 🟥 🟥 🟨 🟩 ⬜ ⬜\n\n#InsightfulGuesser\nhttps://guessthe.game/",
			},
			model.DailyGamePlay{
				GuildId: 601,
				UserId:  602,
				Tries:   4,
				Win:     true,
			},
			nil,
		},
		{
			"worldle_loss",
			args{
				guildId:  701,
				userId:   702,
				gameType: "Worldle",
				message: `#Worldle #670 X/6 (99%)
🟩⬛⬛⬛⬛➡️
🟩🟩🟩🟨⬛↘️
🟩🟩⬛⬛⬛↖️
🟩🟩🟨⬛⬛⬅️
🟩🟩🟩🟩🟨↘️
🟩🟩🟩🟩🟨⬆️
https://worldle.teuteuf.fr`,
			},
			model.DailyGamePlay{
				GuildId: 701,
				UserId:  702,
				Tries:   6,
				Win:     false,
			},
			nil,
		},
		{
			"costcodle",
			args{
				guildId:  801,
				userId:   802,
				gameType: "Costcodle",
				message: `Costcodle #136 3/6
⬆️🟥
⬇️🟥
✅
 https://costcodle.com/`,
			},
			model.DailyGamePlay{
				GuildId: 801,
				UserId:  802,
				Tries:   3,
				Win:     true,
			},
			nil,
		},
		{
			"costcodle_2",
			args{
				guildId:  801,
				userId:   802,
				gameType: "Costcodle",
				message: `Costcodle #182 1/6
✅
https://costcodle.com/`,
			},
			model.DailyGamePlay{
				GuildId: 801,
				UserId:  802,
				Tries:   1,
				Win:     true,
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
		{
			"flashback",
			`Flashback for October 29, 2023

21 points
🟩🟩🟩🟥🟩🟩🟩🟥

Play here: 
        https://www.nytimes.com/interactive/2023/10/27/upshot/flashback.html?hide-chrome=1`,
			true,
		},
		{
			"guessTheGame_yellowSquare",
			`#GuessTheGame #548

🎮 🟥 🟥 🟨 🟩 ⬜ ⬜

#InsightfulGuesser
https://guessthe.game/`,
			true,
		},
		{
			"worldle",
			`#Worldle #670 X/6 (99%)
🟩⬛⬛⬛⬛➡️
🟩🟩🟩🟨⬛↘️
🟩🟩⬛⬛⬛↖️
🟩🟩🟨⬛⬛⬅️
🟩🟩🟩🟩🟨↘️
🟩🟩🟩🟩🟨⬆️
https://worldle.teuteuf.fr`,
			true,
		},
		{
			"costcodle",
			`Costcodle #136 3/6
⬆️🟥
⬇️🟥
✅
 https://costcodle.com/`,
			true,
		},
		{
			"costcodle_2",
			`Costcodle #182 1/6
✅
https://costcodle.com/`,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isGameMessage(tt.message), "isGameMessage(%v)", tt.message)
		})
	}
}
