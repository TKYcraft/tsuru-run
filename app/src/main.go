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
	"log" // 標準log
	"path/filepath" // ファイルパスからファイル名を取得するために使ってる

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/dgvoice"
)

var eLog *log.Logger
var DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
var PREFIX = os.Getenv("PREFIX")

// IDのみをstringで管理
var cMessageChannelID = map[string][]string{}

// VoiceConnectionを保持
var cVoiceChannelVC = map[string]*discordgo.VoiceConnection{}

func init() {
	// defaultのLOG設定
	log.SetPrefix("[LOG]")
	log.SetFlags(log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Error時のLOG出力用
	eLog = log.New(os.Stdout, "[ERROR]", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// .envが存在しないもしくは、.env内にDISCORD_TOKENが存在しません。
	if DISCORD_TOKEN == ""{
		panic("DISCORD_TOKEN is not found in env.")
	}

	// .envが存在しないもしくは、.env内にPREFIXが存在しません。
	if os.Getenv("PREFIX") == ""{
		panic("PREFIX is not found in env")
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		eLog.Println("Error creating Discord session: ", err)
		eLog.Println("DISCORD_TOKEN: ", DISCORD_TOKEN)
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
		eLog.Println("Error opening Discord session: ", err, "DISCORD_TOKEN: ", DISCORD_TOKEN)
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
			eLog.Printf("Could not find the channel that the message came from.\n")
			return
		}

		switch command {
			case "run":
				if cMessageChannelID[c.GuildID] == nil || len(cMessageChannelID[c.GuildID]) == 0 || contains(cMessageChannelID[c.GuildID], m.ChannelID) {
					cMessageChannelID[c.GuildID] = append(cMessageChannelID[c.GuildID], m.ChannelID)
				}

				channelID, err := findVChannel_withUser(s, m.Author.ID, c.GuildID)
				if err != nil {
					sendReply(s, "何らかのエラーです。\nBOTはあなたがボイスチャンネルに参加していない可能性を疑っているみたい。", m.Reference())
					eLog.Printf("BOT seems to have lost track of where the user is\n")
					return
				}
				go sendReply(s, "tsuru.runを実行!", m.Reference())
				go enterVoiceChannel(s, c.GuildID, channelID)

			case "airhorn":
				// func化したい
				if cVoiceChannelVC[c.GuildID] == nil{
					sendReply(s, "tsuru-runはどこのボイスチャンネルにも居ないよ。", m.Reference())
					eLog.Printf("BOT does not exist on any channels.\n")
					return
				}

				go sendVoice(cVoiceChannelVC[c.GuildID], "/go_usr/file/airhorn.dca")
				sendReply(s, "airhorrrrrrrrn", m.Reference())

			case "play":
				// func化したい
				if cVoiceChannelVC[c.GuildID] == nil{
					sendReply(s, "tsuru-runはどこのボイスチャンネルにも居ないよ。", m.Reference())
					eLog.Printf("BOT does not exist on any channels.\n")
					return
				}

				path := "/go_usr/file/chino_and_cocoa.mp3"
				go playAudioFile(cVoiceChannelVC[c.GuildID], path)

				sendReply(s, "play:" + path, m.Reference())

			case "exit":
				// func化したい
				if cVoiceChannelVC[c.GuildID] == nil{
					sendReply(s, "tsuru-runはどこのボイスチャンネルにも居ないよ。", m.Reference())
					eLog.Printf("BOT does not exist on any channels.\n")
					return
				}

				sendReply(s, "ばいばい。", m.Reference())
				exitVoiceChannel(c.GuildID)

			case "debug":
				fmt.Printf(strings.Join(dirwalk("/usr/bin"), "\n"))
				fmt.Printf(strings.Join(dirwalk("/go_usr/file"), "\n"))

		}
	}

}

// guildに追加されるたびに呼び出される。(AddHandlerがあるため)
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		eLog.Printf("Guild is unavailable.")
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
		eLog.Print("Could not find channel.")
	}
	return c, err
}

// セッションを利用し、ギルドIDに沿ったギルドを取得する。
func findGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	// Find the guild for that channel.
	g, err := s.State.Guild(guildID)
	if err != nil {
		eLog.Print("Could not find guild.")
	}
	return g, err
}

// セッションを利用し、ユーザが所属するボイスチャンネルを取得する。
func findVChannel_withUser(s *discordgo.Session, userID string, guildId string) (string, error) {
	g, err := findGuild(s, guildId)
	if err != nil{
		eLog.Printf("Could not find channel with user.\n")
		return "", err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID,  err
		}
	}

	return "", errors.New("Could not find channel with user.")
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
		eLog.Println("Error sending message: ", err)
	}
}

// セッションを利用し、ユーザのメッセージ(reference)に指定のメッセージ(msg)でリプライを送信する。
func sendReply(s *discordgo.Session, msg string, reference *discordgo.MessageReference) {
	_, err := s.ChannelMessageSendReply(reference.ChannelID, msg, reference)
	fmt.Printf("send_reply>>> %s\n", msg)
	if err != nil {
		eLog.Println("Error sending reply message: ", err)
	}
}

// 指定のギルドの指定のボイスチャンネルへ所属する
func enterVoiceChannel(s *discordgo.Session, guildID string, channelID string) {
	if cVoiceChannelVC[guildID] != nil && cVoiceChannelVC[guildID].ChannelID == channelID {
		eLog.Printf("すでにボイスチャンネルに所属している可能性あり\n")
		return
	}

	// 実際に所属し、コネクションをvcへ
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		eLog.Println("Error enter VoiceChannel:", err)
		vc.Disconnect()
		enterVoiceChannel(s, guildID, channelID)
		return
	}

	// 保持しているコネクションをfunc外から取得できるようにグローバル変数へ
	cVoiceChannelVC[guildID] = vc

	return
}

// 指定のギルドのボイスチャンネルから離脱
func exitVoiceChannel(guildID string) {
	if cVoiceChannelVC[guildID] == nil {
		eLog.Printf("cVoiceChannelVC is nil.\n")
		return
	}

	// 一応、スピーキングをfalseにしている()
	cVoiceChannelVC[guildID].Speaking(false)

	// 実際に切断
	cVoiceChannelVC[guildID].Disconnect()

	// グローバル変数に存在するコネクションを明示的にnilへ
	// これがないとコネクションが有効かを毎回確かめる必要あるかも
	// (どっちみち、コネクションが有効か否かは実装する必要ありそう)
	cVoiceChannelVC[guildID] = nil
	return
}


// dca音源送信のため、loadSound()及びplaySound()の呼び出し
func sendVoice(vc *discordgo.VoiceConnection, dca_path string){
	// ボイスコネクションが存在するか確認
	if vc == nil {
		eLog.Printf("VoiceConnection is nil.\n")
		return
	}

	buffer, err := loadSound(dca_path)
	if err != nil {
		eLog.Printf("Error loading sound: %s\n", dca_path)
		return
	}

	playSound(vc, buffer)
}

// pathからサウンドをロードし、bufferへ
func loadSound(path string) ([][]byte, error) {
	// 再生する音源のバッファー(現状は複数guildに対応していない)
	var buffer = make([][]byte, 0)
	// ファイルパスが有効か確認
	file, err := os.Open(path)
	if err != nil {
		eLog.Println("Error opening dca file :", err)
		return buffer, err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return buffer, err
			}
			return buffer, err
		}

		if err != nil {
			eLog.Println("Error reading from dca file :", err, "\nfile path:", path)
			return buffer, err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			eLog.Println("Error reading from dca file :", err, "\nfile path:", path)
			return buffer, err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
	return buffer, err
}

// bufferから音源を再生
func playSound(vc *discordgo.VoiceConnection, buffer [][]byte) {
	// Speakingを有効化
	vc.Speaking(true)

	// bufferから音声を送信
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Speakingを無効化
	vc.Speaking(false)
}

// Takes inbound audio and sends it right back out.
func playAudioFile(vc *discordgo.VoiceConnection, fPath string) {
	// Speakingを有効化
	vc.Speaking(true)

	// Start loop and attempt to play all files in the given folder
	fmt.Println("Reading Folder: ", fPath)
	_, err := os.ReadFile(fPath)
	if err != nil {
		eLog.Printf("ファイルの読み込みに失敗しました:%s", fPath)
		return
	}

	fmt.Println("PlayAudioFile:", fPath)
	dgvoice.PlayAudioFile(vc, fPath, make(chan bool))

	// Speakingを無効化
	vc.Speaking(false)
}

func dirwalk(dir string) []string {
    files, err := os.ReadDir(dir)
    if err != nil {
        panic(err)
    }

    var paths []string
    for _, file := range files {
        if file.IsDir() {
            paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
            continue
        }
        paths = append(paths, (filepath.Join(dir, file.Name()) + "\n"))
    }
    return paths
}
