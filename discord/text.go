package discord

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	slog.Debug("receive message: %s", m.Content)

	if strings.HasPrefix(m.Content, Conf.Prefix) {
		command := strings.TrimPrefix(m.Content, Conf.Prefix)
		slog.Info("receive command: %s", m.Content)
		switch command {
		case "ping":
			PingPong(s, m)
		case "run":
			Run(s, m)
		case "dc":
			Dc(s, m)
		}
	}
}

// セッションを利用し、指定のチャンネルに指定のメッセージを送信する。
func sendMessage(s *discordgo.Session, channelID string, msg string) {
	_, err := s.ChannelMessageSend(channelID, msg)
	slog.Info("send message>>> %s\n", msg)
	if err != nil {
		slog.Error("cannnot sending message: ", err)
	}
}

// セッションを利用し、ユーザのメッセージ(reference)に指定のメッセージ(msg)でリプライを送信する。
func sendReply(s *discordgo.Session, msg string, ref *discordgo.MessageReference) {
	slog.Info("send reply>>> %s", msg)
	_, err := s.ChannelMessageSendReply(ref.ChannelID, msg, ref)
	if err != nil {
		slog.Error("cannot sending reply message: %v", err)
		return
	}
}
