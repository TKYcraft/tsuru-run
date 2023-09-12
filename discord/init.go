package discord

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type State struct {
	VoiceConnection *discordgo.VoiceConnection
}

var guildState = map[string]*State{}

func OnReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "tsuru.run!!")
	_, err := s.UserUpdate("tsuru.run", "")
	if err != nil {
		slog.Error("Failed to update bot name:", err)
		return
	}
}
