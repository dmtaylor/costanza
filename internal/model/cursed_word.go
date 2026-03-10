package model

type CursedWord struct {
	Id      uint   `json:"id"`
	GuildId uint64 `json:"guildId"`
	Word    string `json:"word"`
}
