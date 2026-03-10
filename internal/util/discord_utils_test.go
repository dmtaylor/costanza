package util

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestUserIsServerAdmin(t *testing.T) {
	type args struct {
		member *discordgo.Member
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"regular user",
			args{&discordgo.Member{Permissions: discordgo.PermissionSendMessages | discordgo.PermissionAddReactions}},
			false,
		},
		{
			"server admin",
			args{&discordgo.Member{Permissions: discordgo.PermissionManageGuild | discordgo.PermissionSendMessages}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, UserIsServerAdmin(tt.args.member), "UserIsServerAdmin(%v)", tt.args.member)
		})
	}
}
