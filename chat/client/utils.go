package utils

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gordonklaus/portaudio"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	opus "gopkg.in/hraban/opus.v2"
)

//go:embed sounds/*.mp3
var soundAssets embed.FS

var room *lksdk.Room

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

func StartVoice(roomName, identity, serverAddress string) error {
	type tokenResponse struct {
		JoinToken string `json:"jwtToken"`
		HostUrl   string `json:"hostURL"`
	}

	url := fmt.Sprintf("http://%s:8080/token?name=%s", serverAddress, identity)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("token request failed: %q", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed: %q", body)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize port audio: %q", err)
	}
	defer portaudio.Terminate()

	roomCB := &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: trackSubscribed,
		},
	}

	room, err = lksdk.ConnectToRoomWithToken(tr.HostUrl, tr.JoinToken, roomCB)
	if err != nil {
		return fmt.Errorf("unable to connect to room: %q", err)
	}
	defer room.Disconnect()

	publishMic(room.LocalParticipant, identity)

	select {}
}

func trackSubscribed(remote *webrtc.TrackRemote, _ *lksdk.RemoteTrackPublication, _ *lksdk.RemoteParticipant) {
	// initialize decoder
	dec, err := opus.NewDecoder(48000, 1)
	if err != nil {
		log.Println("Opus decoder error:", err)
		return
	}

	// open default output
	out := make([]int16, 960)
	streamOut, err := portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
	if err != nil {
		log.Println("output stream error:", err)
		return
	}
	streamOut.Start()
	defer streamOut.Stop()

	for {
		pkt, _, err := remote.ReadRTP()
		if err != nil {
			log.Println("RTP read error:", err)
			return
		}
		// decode
		_, err = dec.Decode(pkt.Payload, out)
		if err != nil {
			log.Println("opus decode error:", err)
			continue
		}
		// play
		if err := streamOut.Write(); err != nil {
			log.Println("output write error:", err)
		}
	}
}

func publishMic(lp *lksdk.LocalParticipant, identity string) {
	enc, err := opus.NewEncoder(48000, 1, opus.Application(opus.AppVoIP))
	if err != nil {
		log.Fatalf("Opus encoder error: %v", err)
	}

	track, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", identity,
	)
	if err != nil {
		log.Fatalf("track create error: %v", err)
	}
	if _, err := lp.PublishTrack(track, nil); err != nil {
		log.Fatalf("publish track error: %v", err)
	}

	// open default input (mono 48kHz)
	in := make([]int16, 960)
	streamIn, err := portaudio.OpenDefaultStream(1, 0, 48000, len(in), &in)
	if err != nil {
		log.Fatalf("input stream error: %v", err)
	}
	streamIn.Start()
	go func() {
		defer streamIn.Stop()
		buf := make([]byte, 4000)
		for {
			if err := streamIn.Read(); err != nil {
				if err != io.EOF {
					log.Println("mic read error:", err)
				}
				return
			}
			// encode 20ms
			n, err := enc.Encode(in, buf)
			if err != nil {
				log.Println("opus encode error:", err)
				continue
			}
			if err := track.WriteSample(media.Sample{Data: buf[:n], Duration: time.Millisecond * 20}); err != nil {
				log.Println("write sample error:", err)
			}
		}
	}()
}

func RoomDisconnect() {
	if room != nil {
		room.Disconnect()
		room = nil
	}
}
