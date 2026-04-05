# TextTube MCP Server

This is an MCP (Model Context Protocol) server that exposes TextTube video service tools to LLMs.

## Features

The server provides the following tools:

- `search_channel`: Search for a YouTube channel by name and get its latest videos.
- `get_channel_videos`: Get the latest videos from a specific YouTube channel ID.
- `get_video_details`: Get detailed information about a specific YouTube video.
- `get_video_transcript`: Fetch the transcript for a given YouTube video.
- `summarize_video`: Generate an AI summary for a YouTube video based on its transcript.

## Prerequisites

- [Go](https://golang.org/doc/install) (1.25 or higher)
- Running [TextTube Video Service](file:///Users/shivendra/Projects/text-tube/video-service) (Default: `localhost:50052`)

## Configuration

You can configure the video service address via the `VIDEO_SERVICE_ADDR` environment variable.

## Usage

The MCP server runs over standard I/O. For production-like use and portability, you should first build the binary.

### Build the binary

From the project root:
```bash
make build
```
This creates the binary at `bin/mcp-server`.

### Running for local development
```bash
cd mcp-server
go run main.go
```

### Configuring for MCP Clients (e.g. Claude Desktop)

Point the `command` to the built binary. This is more stable and faster than `go run`.

```json
{
  "mcpServers": {
    "text-tube": {
      "command": "/Users/shivendra/Projects/text-tube/bin/mcp-server",
      "args": [],
      "env": {
        "VIDEO_SERVICE_ADDR": "localhost:50052"
      }
    }
  }
}
```

> [!TIP]
> You can replace `/Users/shivendra/Projects/text-tube/bin/mcp-server` with the absolute path to wherever you move the binary, making the setup portable across your environment.
