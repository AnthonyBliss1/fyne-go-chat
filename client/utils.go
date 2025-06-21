package utils

import (
	"embed"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed assets/*.mp3
var soundAssets embed.FS

func EstablishConnection(displayName string, raw_addy string) (net.Conn, error) {
	serverAddress := strings.TrimSpace(raw_addy) + ":8000"

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %q", err)
	}

	_, err = conn.Write([]byte(displayName + "\n"))
	if err != nil {
		return nil, fmt.Errorf("error sending display name to server: %q", err)
	}

	return conn, nil
}

func SendMessage(conn net.Conn, displayName string, msg string) error {
	_, err := conn.Write([]byte(displayName + ": " + msg + "\n"))
	if err != nil {
		return fmt.Errorf("error sending message to server: %q", err)
	}

	return nil
}

func ExtractName(msg string) (t bool, name, text string) {
	name_index := strings.Index(msg, ":")

	if name_index != -1 {
		name = msg[:name_index]
		text = msg[name_index+2:]
	} else {
		return false, "", ""
	}

	return true, name, text
}

func PlaySound(sound string) {
	f, err := soundAssets.Open(sound)
	if err != nil {
		log.Fatal("sound open:", err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal("mp3 decode:", err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(beep.Seq(
		streamer,
		beep.Callback(func() {
			streamer.Close()
		}),
	))
}
