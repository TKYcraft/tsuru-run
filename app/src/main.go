package main

import (
	"encoding/binary" // dcaファイルのbinary読み込み用
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"errors"
	"time" // sleep用
	"log" // 標準log

	"github.com/bwmarrin/discordgo"
)

var elog *log.Logger
var DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
var PREFIX = os.Getenv("PREFIX")

// IDのみをstringで管理
var c_message_channelID = map[string][]string{}

// VoiceConnectionを保持
var c_voice_channelVC = map[string]*discordgo.VoiceConnection{}

func init() {
	// defaultのLOG設定
	log.SetPrefix("[LOG]")
	log.SetFlags(log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Error時のLOG出力用
	elog = log.New(os.Stdout, "[ERROR]", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// .envが存在しないもしくは、.env内にDISCORD_TOKENが存在しません。
	if DISCORD_TOKEN == ""{
		panic("DISCORD_TOKEN is not found in env.")
	}

	// .envが存在しないもしくは、.env内にPREFIXが存在しません。
	if os.Getenv("PREFIX") == ""{
		panic("PREFIX is not found in env")
	}
}

// 再生する音源のバッファー(現状は複数guildに対応していない)
var buffer = make([][]byte, 0)

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		elog.Println("Error creating Discord session: ", err)
		elog.Println("DISCORD_TOKEN: ", DISCORD_TOKEN)
		return
	}

	// onReadyをreadyのコールバック関数として登録
	dg.AddHandler(onReady)

	// messageCreateをmessageCreateのコールバック関数として登録
	dg.AddHandler(onMessageCreate)

	// onGuildCreateをguildCreateのコールバック関数として登録
	dg.AddHandler(onGuildCreate)

	// Intentsの登録(権限の登録)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// WebSocketのListen開始
	err = dg.Open()
	if err != nil {
		elog.Println("Error opening Discord session: ", err, "DISCORD_TOKEN: ", DISCORD_TOKEN)
	}

	// Ctrl+Cで終了する様にシグナルの取得
	fmt.Println("tsuru-run is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// 上記シグナルの取得後以下が実行
	<-sc

	log.Println("Graceful shutdown")

	// Discordとのセッションを終了
	dg.Close()
}

// Discord上でready状態になった際に呼び出される。(AddHandlerがあるため)
func onReady(s *discordgo.Session, event *discordgo.Ready) {
	// Discord上のプレイステータスを設定
	s.UpdateGameStatus(0, "tsuru.run")
}

// BOTがアクセスできるチャンネルで新しいメッセージが送られるたびに呼び出される。(AddHandlerがあるため)
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is PREFIX
	if strings.HasPrefix(m.Content, PREFIX) {

		command := strings.TrimPrefix(m.Content, PREFIX)
		fmt.Printf("受信したメッセージ内容:%s\n", m.Content)
		fmt.Printf("受信したコマンド:%s\n", command)

		// Find the channel that the message came from.
		c, err := findChannel(s, m.ChannelID)
		if err != nil {
			return
		}

		// Find the guild for that channel.
		//g, err := findChannel(s, c.guildID)
		if err != nil {
			return
		}

		switch command {
			case "run":
				if c_message_channelID[c.GuildID] == nil || len(c_message_channelID[c.GuildID]) == 0 || contains(c_message_channelID[c.GuildID], m.ChannelID) {
					c_message_channelID[c.GuildID] = append(c_message_channelID[c.GuildID], m.ChannelID)
				}

				// log.Printf("c_message_channelID[%s]:%s", c.GuildID, c_message_channelID[c.GuildID])

				channelID, err := findVChannel_withUser(s, m.Author.ID, c.GuildID)
				if err != nil {
					return
				}
				enterVoiceChannel(s, c.GuildID, channelID)

				sendReply(s, "tsuru.runを実行!", m.Reference())

			case "airhorn":
				if c_voice_channelVC[c.GuildID] == nil{
					sendReply(s, "tsuru-runはどこのボイスチャンネルにも居ないよ。", m.Reference())
					elog.Printf("BOT does not exist on any channel")
					return
				}
				go sendVoice(c_voice_channelVC[c.GuildID], "/go_usr/file/airhorn.dca")
				sendReply(s, "airhorrrrrrrrn", m.Reference())

			case "exit":
				if c_voice_channelVC[c.GuildID] == nil{
					sendReply(s, "tsuru-runはどこのボイスチャンネルにも居ないよ。", m.Reference())
					elog.Printf("BOT does not exist on any channel")
					return
				}
				sendReply(s, "ばいばい。", m.Reference())
				exitVoiceChannel(c.GuildID)
		}
	}

}

// guildに追加されるたびに呼び出される。(AddHandlerがあるため)
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "ぜひ、tsuru.runを使ってね")
			return
		}
	}
}

// セッションを利用し、チャンネルIDに沿ったチャンネルを取得する。
func findChannel(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(channelID)
	if err != nil {
		elog.Print("Could not find channel.")
	}
	return c, err
}

// セッションを利用し、ギルドIDに沿ったギルドを取得する。
func findGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	// Find the guild for that channel.
	g, err := s.State.Guild(guildID)
	if err != nil {
		elog.Print("Could not find guild.")
	}
	return g, err
}

// セッションを利用し、ユーザが所属するボイスチャンネルを取得する。
func findVChannel_withUser(s *discordgo.Session, userID string, guildId string) (string, error) {
	g, err := findGuild(s, guildId)
	if err != nil{
		return "", err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID,  err
		}
	}

	return "", errors.New("Could not find Channel with user.")
}

// スライスに文字列が含まれるか
func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

// セッションを利用し、指定のチャンネルに指定のメッセージを送信する。
func sendMessage(s *discordgo.Session, channelID string, msg string) {
	_, err := s.ChannelMessageSend(channelID, msg)
	fmt.Printf("send_message>>> %s\n", msg)
	if err != nil {
		elog.Println("Error sending message: ", err)
	}
}

// セッションを利用し、ユーザのメッセージ(reference)に指定のメッセージ(msg)でリプライを送信する。
func sendReply(s *discordgo.Session, msg string, reference *discordgo.MessageReference) {
	_, err := s.ChannelMessageSendReply(reference.ChannelID, msg, reference)
	fmt.Printf("send_reply>>> %s\n", msg)
	if err != nil {
		elog.Println("Error sending reply message: ", err)
	}
}

// 指定のギルドの指定のボイスチャンネルへ所属する
func enterVoiceChannel(s *discordgo.Session, guildID string, channelID string) {
	log.Printf("c_voice_channelVC[%s]:%s", guildID, c_voice_channelVC[guildID])
	if c_voice_channelVC[guildID] != nil && c_voice_channelVC[guildID].ChannelID == channelID {
		elog.Printf("すでにボイスチャンネルに所属している可能性あり")
		return
	}

	// 実際に所属し、コネクションをvcへ
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		elog.Println("Error enter VoiceChannel:", err)
		vc.Disconnect()
		enterVoiceChannel(s, guildID, channelID)
		return
	}

	// 実際にコネクションが確率されるまでの間sleepをかける。(時間は適当)
	time.Sleep(250 * time.Millisecond)

	// 保持しているコネクションをfunc外から取得できるようにグローバル変数へ
	c_voice_channelVC[guildID] = vc

	return
}

// 指定のギルドのボイスチャンネルから離脱
func exitVoiceChannel(guildID string) {
	if c_voice_channelVC[guildID] == nil {
		return
	}

	// 一応、スピーキングをfalseにしている()
	c_voice_channelVC[guildID].Speaking(false)

	// 実際に切断
	c_voice_channelVC[guildID].Disconnect()

	// グローバル変数に存在するコネクションを明示的にnilへ
	// これがないとコネクションが有効かを毎回確かめる必要あるかも
	// (どっちみち、コネクションが有効か否かは実装する必要ありそう)
	c_voice_channelVC[guildID] = nil
	return
}


// dca音源送信のため、loadSound()及びplaySoundの呼び出し
func sendVoice(vc *discordgo.VoiceConnection, dca_path string){
	// ボイスコネクションが存在するか確認
	if vc == nil {
		elog.Printf("VoiceConnection is nil.")
		return
	}
	err := loadSound(dca_path)
	if err != nil {
		elog.Printf("Error loading sound: %s", dca_path)
	}
	playSound(vc)
}


// pathからサウンドをロードし、bufferへ
func loadSound(path string) error {
	// ファイルパスが有効か確認
	file, err := os.Open(path)
	if err != nil {
		elog.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			elog.Println("Error reading from dca file :", err, "\nfile path:", path)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			elog.Println("Error reading from dca file :", err, "\nfile path:", path)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

// bufferから音源を再生
func playSound(vc *discordgo.VoiceConnection) {
	// Speakingを有効化
	vc.Speaking(true)

	// bufferから音声を送信
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Speakingを無効化
	vc.Speaking(false)
}
