# Kite - The WebAssembly runtime for Discord Bots

Make your own Discord bot without worrying about hosting!

Kite allows your to write Discord Bots in various languages like Go, JavaScript (planned), Rust (planned) and many more and deploy them easily without having to worry about hosting and scaling.

In theory any programming language that can compile to WebAssembly, or which interpreter can compile to it, is supported. Right now there is only an official SDK for Go.

The end goal is to have a public instance of Kite running that users can add to their Discord Servers and easily add their own or community made plugins.

## WIP

Kite is in a very early state and under active development. It's not ready to be used for anything meaningful.

- [x] MVP writing and running plugins
- [x] Model most common Discord types in the Go SDK
- [x] Support most common Discord events
- [x] Support most common Discord HTTP requests
- [ ] Support getting Discord entities from cache
- [x] Support registering Slash Commands
- [ ] Support most common Slash Command responses
- [x] Support KV store with most common actions
- [ ] Support dynamically loading and unloading plugins during runtime

## Installation

```shell
go install github.com/merlinfuchs/kite@latest
```

## Writing a Plugin

### Go

To write plugins in Go you need to have [TinyGo](https://tinygo.org/getting-started/install/) installed.

#### 1. Create a new plugin

```shell
kite plugin --path ./my_plugin init --type go
```

#### 2. Init Go project

```shell
cd my_plugin

go mod init github.com/my_username/my_plugin
```

#### 3. Update the manifest

The `kite.toml` is the manifest for your plugin and can contain various configuration options. You can change the name of your plugin, register commands, etc.

For this example we just have to define that we are using the `DISCORD_MESSAGE_CREATE` event:

```toml
[plugin]
name = 'Ping Plugin'
description = '!ping -> Pong!'
type = 'go'

events = ["DISCORD_MESSAGE_CREATE"]
```

#### 4. Write some code

Create a `plugin.go` file with your Go code for the plugin. It must contain a `main` function which is executed when the plugin is instantiated and can register event handlers.

```go
package main

import (
    kite "github.com/merlinfuchs/kite/go-sdk"
    "github.com/merlinfuchs/kite/go-sdk/log"
    "github.com/merlinfuchs/kite/go-types/discord"
    "github.com/merlinfuchs/kite/go-types/event"
)

func main() {
    kite.Handle(event.DiscordMessageCreate, func(req event.Event) error {
        msg := req.Data.(discord.MessageCreateEvent)

        if msg.Content == "!ping" {
            _, err := kite.CreateMessage(msg.ChannelID, "Pong!")
            if err != nil {
                log.Error("Failed to send message: " + err.Error())
                return err
            }
        }

        return nil
    })
}
```

#### 5. Compile it

To deploy your plugin you first have to compile to a WASM file. Kite provides commands for the various plugin types which then call the specific compiler.

```shell
# Retain debug information in the WASM file which helps with finding issues
kite plugin build --debug

# Create the smallest possible WASM file, good for deployment
kite plugin build
```

## Running the Server

### 1. Configure the server

Create a `kite.toml` file which contains configuration values for the Kite server:

```toml
[server.discord]
token = "" # Your Discord Bot token
client_id = "" # Your Discord Bot client id

[[server.plugins]]
path = "./my_plugin" # Path to your previous created plugin (must be compiled first)
```

### 2. Run the server

```shell
kite server start
```
