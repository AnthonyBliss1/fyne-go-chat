package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"image/color"
	"io"
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

//go:embed send.svg
var sendPng []byte

var sendIcon = fyne.NewStaticResource("send.svg", sendPng)

func generateConnectionWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("Connect to Server")

	title := canvas.NewText("Connect to Server", color.White)
	title.TextSize = 30
	title.TextStyle.Italic = true
	title.Alignment = fyne.TextAlignCenter

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder("Enter Display Name")

	serverAddress := widget.NewEntry()
	serverAddress.SetPlaceHolder("Enter Server Address")

	connectBtn := widget.NewButton("Connect", func() {
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
	))

	w.SetOnClosed(func() { a.Quit() })

	w.Resize(fyne.NewSize(400, 200))
	w.SetFixedSize(true)

	return w
}

func generateMessengerWindow(a fyne.App, displayName string, conn net.Conn) fyne.Window {
	w := a.NewWindow("Go Chat Messenger")

	msgArea := container.New(layout.NewVBoxLayout())

	scrollArea := container.NewVScroll(msgArea)

	msg := widget.NewEntry()
	msg.SetPlaceHolder("Send a message...")

	send := func() {
		msgBubble := generateMessageBubble(msg.Text, displayName, true)
		msgArea.Add(msgBubble)
		scrollArea.ScrollToBottom()
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

	utils.PlaySound("assets/zelda_secret.mp3")
	return conn, true
}

func incomingMessage(conn net.Conn, msgArea *fyne.Container, scrollArea *container.Scroll) {
	var msgBubble *fyne.Container
	rd := bufio.NewReader(conn)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				msgBubble = generateMessageBubble("<Server Disconnected>", "Server", false)
				utils.PlaySound("assets/noti.mp3")
			} else {
				msgBubble = generateMessageBubble(fmt.Sprintf("%q", err), "Server", false)
				utils.PlaySound("assets/noti.mp3")
			}
		}

		if t, senderName, text := utils.ExtractName(strings.TrimRight(line, "\r\n")); t {
			msgBubble = generateMessageBubble(text, senderName, false)
			utils.PlaySound("assets/noti.mp3")
		} else {
			msgBubble = generateMessageBubble(strings.TrimRight(line, "\r\n"), "Server", false)
			utils.PlaySound("assets/noti.mp3")
		}
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title: "New Message",
		})

		msgArea.Add(msgBubble)
		scrollArea.ScrollToBottom()
	}
}

func main() {
	a := app.New()

	base := theme.DefaultTheme()
	a.Settings().SetTheme(&ui.ForcedVariant{
		Theme:   base,
		Variant: theme.VariantDark,
	})

	connection := generateConnectionWindow(a)
	connection.Show()

	a.Run()
}
