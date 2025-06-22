# Go Chat GUI

A chat room style project utilizing a TCP server (`server.go`) to broadcast messages to connected clients. Use `chat/main.go` to establish a connection to the server and send messages.

The client side code, `chat/main.go`, will prompt the user for a server address to connect to. Also, `server.go` is currently set up to listen on `Port 8000` so make sure this port is not in use or change the port configuration in `chat/main.go` and `server.go`.

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
```

2. **Install Fyne**
```bash
go get fyne.io/fyne/v2@latest
go install fyne.io/tools/cmd/fyne@latest
```

3. **Run Server and Client Packages**
```bash
go run server/server.go
```

```bash
go run chat/main.go
```

4. **(Optional) Build Executable**
```bash
go build -o builds/server ./server/server.go
```

***For MacOS***
```bash
cd chat
export GOFLAGS="-buildvcs=false"
fyne package -os darwin -icon icon.png
```

***For Windows***
```bash
cd chat
export GOFLAGS="-buildvcs=false"
fyne package -os windows -icon icon.png
```
