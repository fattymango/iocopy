# Screen Streaming System

This implements a screen capture streaming system with separate server and client applications.

## Architecture

```
Screen Capture → TCP Server (:9000) → TCP Client → WebSocket Server (:8080/ws) → Wails Desktop Client
```

## Building

### Server (Standalone, no Wails)

```bash
go build -o stream-server.exe ./cmd/server
```

### Client (Wails Desktop App)

1. First, you need to set up the frontend. For the client, you need to use `stream_index.html` as the main page:

```bash
# Option 1: Rename stream_index.html to index.html temporarily
copy frontend\stream_index.html frontend\index.html.backup
copy frontend\index.html frontend\index.html.original
copy frontend\stream_index.html frontend\index.html

# Build the client (this will generate Wails bindings)
wails build -o stream-client.exe

# Restore original index.html if needed
copy frontend\index.html.original frontend\index.html
```

2. Or build directly:
```bash
wails build -o stream-client.exe
```

Then manually swap `stream_index.html` to `index.html` before building, or use a build script.

## Running

### 1. Start the Server

```bash
stream-server.exe
# Or with custom options:
stream-server.exe -tcp :9000 -ws :8080 -fps 25
```

The server will:
- Capture the screen
- Stream frames over TCP on port 9000
- Bridge to WebSocket on port 8080/ws

### 2. Start the Client

```bash
stream-client.exe
# Or with custom WebSocket address:
stream-client.exe -ws localhost:8080
```

The client will:
- Connect to the WebSocket server
- Display frames in a Wails desktop window
- Show connection status and frame statistics

## Features

- **Pure Go** implementation
- **JPEG encoding** (quality 65, configurable)
- **Length-prefixed binary frames** over TCP
- **WebSocket binary messages** (no Base64)
- **Multiple client support** (one-to-many)
- **Auto-reconnection** with exponential backoff
- **Frame dropping** to prevent blocking
- **FPS control** (default 25 FPS)

## File Structure

```
stream/
  ├── tcp_server.go      # TCP server that captures and streams
  ├── tcp_client.go      # TCP client that reads frames
  ├── websocket_server.go # WebSocket broadcaster
  └── stream.go          # Main orchestrator

cmd/
  ├── server/            # Standalone server app
  └── client/           # Wails desktop client app

frontend/
  └── stream_index.html # Client UI (needs to be index.html for client build)
```

## Notes

- The server runs independently (no Wails)
- The client is a Wails desktop application
- Wails bindings are auto-generated when you build the client
- Make sure `stream_index.html` is used as `index.html` when building the client




