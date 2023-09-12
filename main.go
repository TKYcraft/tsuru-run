package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/TKYcraft/tsuru-run/discord"
	"github.com/bwmarrin/discordgo"
)

func main() {
	var err error
	conf, err := discord.LoadConfig("./config.json")
	if err != nil {
		slog.Error("cannot load config. %v", err)
	}

	if conf.DebugLog {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}

	s, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		slog.Error("cannot creating Discord session. %v", err)
	}

	s.AddHandler(discord.OnReady)
	s.AddHandler(discord.OnMessageCreate)

	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	if err = s.Open(); err != nil {
		slog.Error("cannot opening Discord session. %v TOKEN: %s", err, conf.Token)
	}

	// Ctrl+Cで終了する様にシグナルの取得
	slog.Info("tsuru is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Info("Graceful shutdown")
	s.Close()
}
