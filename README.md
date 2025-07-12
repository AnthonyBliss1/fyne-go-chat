# Go Chat GUI

A chat room style project utilizing a TCP server (`server.go`) to broadcast messages to connected clients. This application also utilizes WebRTC for realtime voice chat. Use `chat/main.go` to establish a connection to the server, send messages, and voice chat.

The client side code, `chat/main.go`, will prompt the user for a server address to connect to. This server address will be the location of the deployed `server.go` build.

## Server Setup
The `server.go` code creates an HTTP server `(port 8080)` and a TCP server `(port 8000)`. Make sure these ports are not in use or change the port configuration in `chat/main.go` and `server.go`.

The `server.go` file requires a `.env` file with multiple environment variables defined (outlined below).

| Variable | Usage |
| ------- | ----- |
| OPENAI_API_KEY | OpenAI API Key required to use the #chat command |
| LIVEKIT_URL | Livekit URL either pointing to a self-hosted or cloud instance |
| LIVEKIT_API_KEY | Livekit API Key provided by self-hosted or cloud instance |
| LIVEKIT_API_SECRET | Livekit API Secret provided by self-hosted or cloud instance |

>[!IMPORTANT]
>Livekit is a "batteries-included" solution for WebRTC implementation. Go Chat uses Livekit for realtime voice chat which means a Livekit server must be deployed either on your own machine or in the cloud. I recommend using Livekit's free builder plan which will make the Go Chat setup much easier.

## Commands
***Send commands with `#`***

| Command | Usage |
| ------- | ----- |
| #room | Show the current users connected to the server |
| #chat "{prompt}" | Send a message to the AI bot |

- To the use the `#chat` command, you need a `.env` file that includes your `OPENAI_API_KEY` next to the server code.
- Make sure to wrap your prompt in quotes:
    - `#chat "Hello!"`

## Usage

1. **Clone the Repo**
```bash
git clone https://github.com/AnthonyBliss1/go-chat-gui.git
cd go-chat-gui
go mod tidy
```

2. **Install Additional Libraries**
```bash
brew install pkgconf opus opusfile portaudio
```

3. **Run Server and Client Packages**
```bash
go run server/server.go
```

```bash
go run chat/main.go
```

4. **(Optional) Build Server Executable**
```bash
go build -o builds/server ./server/server.go
```
