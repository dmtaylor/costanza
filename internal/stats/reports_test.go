package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dmtaylor/costanza/internal/model"
)

func TestBuildGameWinReport(t *testing.T) {
	type args struct {
		topWinners []*model.DailyGameWinStat
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"simple",
			args{
				[]*model.DailyGameWinStat{
					{
						Id:            1,
						GuildId:       999,
						UserId:        8888,
						PlayCount:     10,
						GuessCount:    20,
						WinCount:      8,
						CurrentStreak: 5,
						MaxStreak:     5,
					},
				},
			},
			"Top game winners for the month are:\n#1: <@8888> with 8 wins (win rate 80.00%, average guesses 2.00, longest streak 5)\n",
		},
		{
			"multi",
			args{
				[]*model.DailyGameWinStat{
					{
						Id:            5,
						GuildId:       777,
						UserId:        88892,
						PlayCount:     50,
						GuessCount:    60,
						WinCount:      50,
						CurrentStreak: 50,
						MaxStreak:     50,
					},
					{
						Id:            4,
						GuildId:       777,
						UserId:        77703,
						PlayCount:     50,
						GuessCount:    70,
						WinCount:      45,
						CurrentStreak: 10,
						MaxStreak:     35,
					},
					{
						Id:            6,
						GuildId:       777,
						UserId:        99981,
						PlayCount:     50,
						GuessCount:    80,
						WinCount:      30,
						CurrentStreak: 5,
						MaxStreak:     10,
					},
				},
			},
			"Top game winners for the month are:\n" +
				"#1: <@88892> with 50 wins (win rate 100.00%, average guesses 1.20, longest streak 50)\n" +
				"#2: <@77703> with 45 wins (win rate 90.00%, average guesses 1.40, longest streak 35)\n" +
				"#3: <@99981> with 30 wins (win rate 60.00%, average guesses 1.60, longest streak 10)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildGameWinReport(tt.args.topWinners)
			assert.Equalf(t, tt.want, got, "BuildGameWinReport(%v)", tt.args.topWinners)
		})
	}
}

func TestBuildMessageReport(t *testing.T) {
	tests := []struct {
		name  string
		stats []*model.DiscordUsageStat
		want  string
	}{
		{
			"basic",
			[]*model.DiscordUsageStat{
				{
					Id:           990,
					GuildId:      888888,
					UserId:       4523,
					ReportMonth:  "2024-01",
					MessageCount: 101,
				},
				{
					Id:           53,
					GuildId:      888888,
					UserId:       9923,
					ReportMonth:  "2024-01",
					MessageCount: 99,
				},
				{
					Id:           991,
					GuildId:      888888,
					UserId:       1023,
					ReportMonth:  "2024-01",
					MessageCount: 98,
				},
			},
			"Top posters for the month are:\n" +
				"#1: <@4523> with 101 messages\n" +
				"#2: <@9923> with 99 messages\n" +
				"#3: <@1023> with 98 messages\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildMessageReport(tt.stats)
			assert.Equalf(t, tt.want, got, "BuildMessageReport(%v, %v)", tt.stats)
		})
	}
}

func TestBuildReactionScoreReport(t *testing.T) {
	tests := []struct {
		name              string
		topReactionScores []*model.DiscordReactionScore
		want              string
	}{
		{
			"basic",
			[]*model.DiscordReactionScore{
				{
					GuildId:     990,
					UserId:      10240,
					ReportMonth: "2024-01",
					Score:       59,
				},
				{
					GuildId:     990,
					UserId:      7892,
					ReportMonth: "2024-01",
					Score:       44,
				},
				{
					GuildId:     990,
					UserId:      40890,
					ReportMonth: "2024-01",
					Score:       32,
				},
			},
			"Top reaction scores are:\n" +
				"#1: <@10240> with score 59\n" +
				"#2: <@7892> with score 44\n" +
				"#3: <@40890> with score 32\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildReactionScoreReport(tt.topReactionScores)
			assert.Equalf(t, tt.want, got, "BuildReactionScoreReport(%v, %v)", tt.topReactionScores)
		})
	}
}
