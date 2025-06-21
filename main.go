package main

import (
	"bufio"
	"embed"
	_ "embed"
	"fmt"
	"image/color"
	"io"
	"log"
	"net"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	utils "github.com/anthonybliss1/fyne-go-chat/ui/client"
	ui "github.com/anthonybliss1/fyne-go-chat/ui/theme"
)

//go:embed send.svg icon.png connect.svg
var embeddedAssets embed.FS

var sendIcon, appIcon, connectIcon fyne.Resource

func init() {
	send, err := embeddedAssets.ReadFile("send.svg")
	if err != nil {
		log.Panicf("failed to load send.svg: %q", err)
	}
	sendIcon = fyne.NewStaticResource("send.svg", send)

	icon, err := embeddedAssets.ReadFile("icon.png")
	if err != nil {
		log.Panicf("failed to load icon.png: %q", err)
	}
	appIcon = fyne.NewStaticResource("icon.png", icon)

	connect, err := embeddedAssets.ReadFile("connect.svg")
	if err != nil {
		log.Panicf("failed to load connect.svg: %q", err)
	}
	connectIcon = fyne.NewStaticResource("connect.svg", connect)
}

func generateConnectionWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("Connect to Server")

	title := canvas.NewText("Connect to Server", color.White)
	title.TextSize = 30
	//title.TextStyle.Italic = true
	title.Alignment = fyne.TextAlignCenter

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder("Display Name")

	serverAddress := widget.NewEntry()
	serverAddress.SetPlaceHolder("Server Address")

	connectBtn := widget.NewButtonWithIcon("Connect", connectIcon, func() {
		conn, t := dialServer(w, displayName, serverAddress)

		if t {
			w.Hide()
			msgr := generateMessengerWindow(a, displayName.Text, conn)
			msgr.CenterOnScreen()
			msgr.Show()
		}
	})

	w.SetContent(container.NewVBox(
		title,
		layout.NewSpacer(),
		displayName,
		layout.NewSpacer(),
		serverAddress,
		layout.NewSpacer(),
		connectBtn,
		layout.NewSpacer(),
	))

	w.SetOnClosed(func() { a.Quit() })

	w.Resize(fyne.NewSize(400, 200))
	w.SetFixedSize(true)

	return w
}

func generateMessengerWindow(a fyne.App, displayName string, conn net.Conn) fyne.Window {
	var isBanner = false
	w := a.NewWindow("Go Chat Messenger")

	msgArea := container.New(layout.NewVBoxLayout())

	banner := `
 ______     ______        ______     __  __     ______     ______
/\  ___\   /\  __ \      /\  ___\   /\ \_\ \   /\  __ \   /\__  _\
\ \ \__ \  \ \ \/\ \     \ \ \____  \ \  __ \  \ \  __ \  \/_/\ \/
 \ \_____\  \ \_____\     \ \_____\  \ \_\ \_\  \ \_\ \_\    \ \_\
  \/_____/   \/_____/      \/_____/   \/_/\/_/   \/_/\/_/     \/_/

Hi %s
Welcome to Go Chat!

`
	goChatLabel := widget.NewLabelWithStyle(fmt.Sprintf(banner, displayName), fyne.TextAlignCenter, fyne.TextStyle{})
	msgArea.Add(goChatLabel)
	isBanner = true

	scrollArea := container.NewVScroll(msgArea)

	msg := widget.NewEntry()
	msg.SetPlaceHolder("Send a message...")

	send := func() {
		if isBanner {
			msgArea.Remove(goChatLabel)
			isBanner = false
		}
		msgBubble := generateMessageBubble(msg.Text, displayName, true)
		fyne.Do(func() {
			msgArea.Add(msgBubble)
			scrollArea.ScrollToBottom()
		})
		if err := utils.SendMessage(conn, displayName, msg.Text); err != nil {
			dialog.ShowInformation("Error Sending Message", fmt.Sprintf("%s", err), w)
		}
		msg.SetText("")
	}

	msgSend := widget.NewButtonWithIcon("", sendIcon, send)

	msgInput := container.NewBorder(nil, nil, nil, msgSend, msg)

	w.SetContent(container.NewBorder(nil, msgInput, nil, nil, scrollArea))

	msg.OnSubmitted = func(_ string) {
		send()
	}

	w.Resize(fyne.NewSize(900, 600))
	w.SetFixedSize(true)

	w.SetOnClosed(func() { a.Quit() })

	go incomingMessage(conn, msgArea, scrollArea)

	return w
}

func generateMessageBubble(msg string, displayName string, isUser bool) *fyne.Container {
	var bubble *canvas.Rectangle

	msgLabel := widget.NewLabel(msg)
	msgLabel.Wrapping = fyne.TextWrapWord

	nameLabel := canvas.NewText(" "+"<"+displayName+">", color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	nameLabel.TextSize = 12

	switch displayName {
	case "Server":
		orange := color.NRGBA{R: 224, G: 51, B: 11, A: 100}
		bubble = canvas.NewRectangle(orange)

	case "AI":
		blue := color.NRGBA{R: 11, G: 109, B: 224, A: 100}
		bubble = canvas.NewRectangle(blue)

	default:
		purple := color.NRGBA{R: 102, G: 12, B: 225, A: 100}
		bubble = canvas.NewRectangle(purple)
	}

	bubble.CornerRadius = 12
	bubble.SetMinSize(fyne.NewSize(400, 20))

	content := container.NewBorder(nil, nameLabel, nil, nil, msgLabel)

	if isUser {
		return container.New(layout.NewHBoxLayout(),
			layout.NewSpacer(),
			container.NewStack(
				bubble,
				container.NewPadded(content),
			),
		)
	} else {
		return container.New(layout.NewHBoxLayout(),
			container.NewStack(
				bubble,
				container.NewPadded(content),
			),
		)
	}
}

func dialServer(window fyne.Window, displayName, serverAddress *widget.Entry) (net.Conn, bool) {
	conn, err := utils.EstablishConnection(displayName.Text, serverAddress.Text)
	if err != nil {
		dialog.ShowInformation("Error Connecting to Server", fmt.Sprintf("%s", err), window)
		return nil, false
	}

	utils.PlaySound("sounds/zelda_secret.mp3")
	return conn, true
}

func incomingMessage(conn net.Conn, msgArea *fyne.Container, scrollArea *container.Scroll) {
	var msgBubble *fyne.Container
	rd := bufio.NewReader(conn)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				msgBubble = generateMessageBubble(fmt.Sprintf("%q", err), "Server", false)
				utils.PlaySound("sounds/noti.mp3")
				fyne.Do(func() {
					msgArea.Add(msgBubble)
					scrollArea.ScrollToBottom()
				})
				break
			} else {
				msgBubble = generateMessageBubble("<Server Disconnected>", "Server", false)
				utils.PlaySound("sounds/noti.mp3")
				fyne.Do(func() {
					msgArea.Add(msgBubble)
					scrollArea.ScrollToBottom()
				})
				break
			}
		}

		if t, senderName, text := utils.ExtractName(strings.TrimRight(line, "\r\n")); t {
			msgBubble = generateMessageBubble(text, senderName, false)
			utils.PlaySound("sounds/noti.mp3")
		} else {
			msgBubble = generateMessageBubble(strings.TrimRight(line, "\r\n"), "Server", false)
			utils.PlaySound("sounds/noti.mp3")
		}
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title: "New Message",
		})

		fyne.Do(func() {
			msgArea.Add(msgBubble)
			scrollArea.ScrollToBottom()
		})
	}
}

func main() {
	a := app.New()

	base := theme.DefaultTheme()
	a.Settings().SetTheme(&ui.ForcedVariant{
		Theme:   base,
		Variant: theme.VariantDark,
	})
	a.SetIcon(appIcon)

	connection := generateConnectionWindow(a)
	connection.Show()

	a.Run()
}
