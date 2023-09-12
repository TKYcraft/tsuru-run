package discord

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func PingPong(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendReply(s, "pong", m.Reference())
	sendMessage(s, m.ChannelID, "tmp")
}

func Run(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendReply(s, "Run", m.Reference())
	vcId, err := findVChannelIDWithUser(s, m.Author.ID, m.GuildID)
	if err != nil {
		sendReply(s, "何らかのエラーです。\nBOTはあなたがボイスチャンネルに参加していない可能性を疑っているみたい。", m.Reference())
		slog.Error("BOT seems to have lost track of where the user is")
		return
	}
	enterVoiceChannel(s, m.GuildID, vcId)
}

func Dc(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendReply(s, "Dc", m.Reference())
	exitVoiceChannel(m.GuildID)
}
