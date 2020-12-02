# discordgo_voicestateupdatequeue
> Event enhancer and queue for the DiscordGo VoiceStateUpdate event

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ethanent/discordgo_voicestateupdatequeue)](https://pkg.go.dev/github.com/ethanent/discordgo_voicestateupdatequeue)

## Features

- Cleans up raw events from Discord using an internal cache.
- Transparently handles voice channel moves, providing them to the consumer as a leave then join event.
- Consistently sends events with both GuildID and ChannelID. (Except for the [VoiceChannelLeaveUnknownChannel](https://pkg.go.dev/github.com/ethanent/discordgo_voicestateupdatequeue#VoiceStateEventType) event.)

## Install

```go
go get github.com/ethanent/discordgo_voicestateupdatequeue
```

## Use

Create a channel through which to receive [VoiceStateEvent](https://pkg.go.dev/github.com/ethanent/discordgo_voicestateupdatequeue#VoiceStateEvent)s.

```go
c := make(chan *queue.VoiceStateEvent)
```

Create a new VoiceStateEventQueue and give it the channel.

```go
q := queue.NewVoiceStateEventQueue(c)
```

Add the queue's handler to your session.

```go
// Assuming your *discordgo.Session is called s:

s.AddHandler(q.Handler)
```

Then you can use the channel to receive [VoiceStateEvent](https://pkg.go.dev/github.com/ethanent/discordgo_voicestateupdatequeue#VoiceStateEvent)s while the bot is running.

The events will contain a type (join, leave, leave from untracked channel, settings update) and other associated fields.
