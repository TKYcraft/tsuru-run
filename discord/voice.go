package discord

import (
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// 指定のギルドの指定のボイスチャンネルへ所属する
func enterVoiceChannel(s *discordgo.Session, guildID string, channelID string) {
	if guildState[guildID] != nil && guildState[guildID].VoiceConnection != nil && guildState[guildID].VoiceConnection.ChannelID == channelID {
		slog.Warn("すでにボイスチャンネルに所属している可能性あり")
		return
	}

	// 実際に所属し、コネクションをvcへ
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		slog.Error("cannot enter VoiceChannel:", err)
		vc.Disconnect()
		enterVoiceChannel(s, guildID, channelID)
		return
	}

	// 保持しているコネクションをfunc外から取得できるようにグローバル変数へ
	guildState[guildID] = &State{
		VoiceConnection: vc,
	}
}

// 指定のギルドのボイスチャンネルから離脱
func exitVoiceChannel(guildID string) {
	if guildState[guildID] == nil || guildState[guildID].VoiceConnection == nil {
		slog.Error("cVoiceChannelVC is nil.\n")
		return
	}

	// 一応、スピーキングをfalseにしている()
	guildState[guildID].VoiceConnection.Speaking(false)

	// 実際に切断
	guildState[guildID].VoiceConnection.Disconnect()

	// グローバル変数に存在するコネクションを明示的にnilへ
	// これがないとコネクションが有効かを毎回確かめる必要あるかも
	// (どっちみち、コネクションが有効か否かは実装する必要ありそう)
	guildState[guildID].VoiceConnection = nil
}

// セッションを利用し、ギルドIDに沿ったギルドを取得する。
func findGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	// Find the guild for that channel.
	g, err := s.State.Guild(guildID)
	if err != nil {
		slog.Error("could not find guild.")
	}

	slog.Debug("find guild: %v", g.ID)
	return g, err
}

// セッションを利用し、ユーザが所属するボイスチャンネルを取得する。
func findVChannelIDWithUser(s *discordgo.Session, userID string, guildId string) (string, error) {
	g, err := findGuild(s, guildId)
	if err != nil {
		slog.Error("could not find channel with user.")
		return "", err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			slog.Debug("user voice channel: %v", vs.ChannelID)
			return vs.ChannelID, nil
		}
	}

	return "", errors.New("could not find channel with user.")
}
