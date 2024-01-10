package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscordReactionScore_FormatResult(t *testing.T) {
	type fields struct {
		GuildId     uint64
		UserId      uint64
		ReportMonth string
		Score       int
	}
	type args struct {
		userString string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"simple",
			fields{
				GuildId:     100,
				UserId:      101,
				ReportMonth: "2024-01",
				Score:       101,
			},
			args{"Dick Halloran"},
			"Dick Halloran with score 101",
		},
		{
			"negative_score",
			fields{
				GuildId:     200,
				UserId:      201,
				ReportMonth: "2024-01",
				Score:       -5,
			},
			args{"Jack Torrance"},
			"Jack Torrance with score -5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DiscordReactionScore{
				GuildId:     tt.fields.GuildId,
				UserId:      tt.fields.UserId,
				ReportMonth: tt.fields.ReportMonth,
				Score:       tt.fields.Score,
			}
			assert.Equalf(t, tt.want, d.FormatResult(tt.args.userString), "FormatResult(%v)", tt.args.userString)
		})
	}
}
