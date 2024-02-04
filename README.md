# Kite - The WebAssembly runtime for Discord Bots

Make your own Discord bot without worrying about hosting!

Kite allows your to write Discord Bots in various languages like Go, JavaScript (WIP), Rust (planned) and many more and deploy them easily without having to worry about hosting and scaling.

In theory any programming language that can compile to WebAssembly, or which interpreter can compile to it, is supported. Right now there is only an official SDK for Go.

The end goal is to have a public instance of Kite running that users can add to their Discord Servers and easily add their own or community made plugins.

## WIP

Kite is in a very early state and under active development. It's not ready to be used for anything meaningful.

Take a look at the [issues](https://github.com/merlinfuchs/kite/issues) to see what is being worked on right now.

## Installation

```shell
# This installs the Kite CLI and Kite API server but doesn't include the web app
go install github.com/merlinfuchs/kite/kite-service@latest
```

## Running the Server

### 1. Configure the server

Create a `kite.toml` file which contains configuration values for the Kite server:

```toml
[server.discord]
token = "" # Your Discord Bot token
client_id = "" # Your Discord Bot client id
client_secret = "" # Your DIscord Bot client secret
```

### 2. Run the server

```shell
kite server start
```

## Authenticating with the Server

To be able to deploy your plugins from the local CLI you first have to authenticate with the Kite server.

```shell
# Authenticate with the locally running server
kite config login

# Authenticate with a specific server
kite config login --server=https://kite.onl
```

You will be prompted to open an URL in your browser which will redirect your to Discord where you authenticate. Once you are redirected back your should be logged in automatically by the CLI. Right now sessions are valid for 30 days.

## Writing a Plugin

### Go

To write plugins in Go you need to have [TinyGo](https://tinygo.org/getting-started/install/) installed.

#### 1. Create a new plugin

```shell
# The key should uniquely identify your plugin, it's good practice follow this format: <plugin–name>@<discord-user-name>
# There can only be one deployment with a certain key per server and only one published plugin on the marketplace
kite plugin init --type go --key my_go_plugin@me ./my_plugin
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
[deployment]
name = 'Ping Plugin'
description = '!ping -> Pong!'

[module]
type = 'go'
```

#### 4. Write some code

Create a `plugin.go` file with your Go code for the plugin. It must contain a `main` function which is executed when the plugin is instantiated and can register event handlers.

```go
package main

import (
    kite "github.com/merlinfuchs/kite/kite-sdk-go"
    "github.com/merlinfuchs/kite/kite-sdk-go/log"
    "github.com/merlinfuchs/kite/kite-types/discord"
    "github.com/merlinfuchs/kite/kite-types/event"
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

### JavaScript

To write plugins in JS you need to first install the custom compiler by following the instructions in the [kite-sdk-js](kite-sdk-js).

#### 1. Create a new plugin

```shell
# The key should uniquely identify your plugin, it's good practice follow this format: <plugin–name>@<discord-user-name>
# There can only be one deployment with a certain key per server and only one published plugin on the marketplace
kite plugin init --type js --key my_js_plugin@me ./my_plugin
```

#### 2. Update the manifest

The `kite.toml` is the manifest for your plugin and can contain various configuration options. You can change the name of your plugin, register commands, etc.

For this example we just have to define that we are using the `DISCORD_MESSAGE_CREATE` event:

```toml
[deployment]
name = 'Ping Plugin'
description = '!ping -> Pong!'

[module]
type = 'js'
```

#### 3. Write some code

There is no high level SDK or typings for JavaScript yet, you therefore have to write raw Kite host calls.

```js
Kite.handle = function (event) {
  if (event.type != "DISCORD_MESSAGE_CREATE") return { success: true };

  const data = event.data;

  if (data.content == "!ping") {
    Kite.call({
      type: "DISCORD_MESSAGE_CREATE",
      data: {
        channel_id: data.channel_id,
        content: "Pong!",
      },
    });
  }

  return { success: true };
};
```

## Deploying a Plugin

### 1. Compile it

Before you can deploy your plugin you first have to compile it to a WASM file. Kite provides commands for the various plugin types which then call the specific compiler.

```shell
kite plugin build
```

### 2. Deploy it

To deploy your plugin you need the id of the guild / server that you want to deploy it to. Make sure the bot is a member of that server!

```shell
# Deploy to the publicly instance (https://api.kite.onl)
kite plugin deploy --guild_id 1234567890

# Deploy to a locally running server
kite plugin deploy --guild_id 1234567890 --server "https://localhost:8080"
```
