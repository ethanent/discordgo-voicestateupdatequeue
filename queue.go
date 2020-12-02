package main

import (
	"github.com/bwmarrin/discordgo"
	"sync"
)

type VoiceStateEventType int

const (
	// A user has joined a voice channel.
	VoiceChannelJoin VoiceStateEventType = iota

	// A user has left a tracked voice channel.
	VoiceChannelLeave

	// A user has left an untracked channel. ChannelID will not be available, but GuildID still will be.
	VoiceChannelLeaveUnknownChannel

	// A user has changed a setting (eg. mute) without switching channels.
	VoiceChannelSettingUpdate
)

// VoiceStateEvent is adapted from a VoiceStateUpdate received from discordgo.
// It is enhanced with a cache. It makes the nature of events much more clear to the consumer.
// It also handles channel moves seamlessly, converting them to separate leave and join events.
type VoiceStateEvent struct {
	Event          VoiceStateEventType
	GuildID        string
	ChannelID      string
	OriginalUpdate *discordgo.VoiceStateUpdate
}

type userVoiceState struct {
	guildID   string
	channelID string
}

type VoiceStateUpdateQueue struct {
	userVoiceStates    map[string]*userVoiceState
	userVoiceStatesMux sync.Mutex
	Out                chan *VoiceStateEvent
}

func (q *VoiceStateUpdateQueue) Handler(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	q.userVoiceStatesMux.Lock()
	defer q.userVoiceStatesMux.Unlock()

	if e.ChannelID == "" {
		// This is a VoiceChannelLeave or VoiceChannelLeaveUnknownChannel

		userState, ok := q.userVoiceStates[e.UserID]

		if !ok {
			// VoiceChannelLeaveUnknownChannel

			q.Out <- &VoiceStateEvent{
				Event:          VoiceChannelLeaveUnknownChannel,
				GuildID:        e.GuildID,
				ChannelID:      e.ChannelID,
				OriginalUpdate: e,
			}

			return
		}

		// VoiceChannelLeave

		q.Out <- &VoiceStateEvent{
			Event:          VoiceChannelLeave,
			GuildID:        e.GuildID,
			ChannelID:      userState.channelID,
			OriginalUpdate: e,
		}

		// Remove user voice state

		delete(q.userVoiceStates, e.UserID)

		return
	}

	// User has joined a channel, moved channels, or updated a setting.

	userState, ok := q.userVoiceStates[e.UserID]

	if !ok {
		// This is a VoiceChannelJoin.

		q.Out <- &VoiceStateEvent{
			Event:          VoiceChannelJoin,
			GuildID:        e.GuildID,
			ChannelID:      e.ChannelID,
			OriginalUpdate: e,
		}

		q.userVoiceStates[e.UserID] = &userVoiceState{
			guildID:   e.GuildID,
			channelID: e.ChannelID,
		}

		return
	}

	if userState.channelID == e.ChannelID {
		// This is a VoiceChannelSettingUpdate

		q.Out <- &VoiceStateEvent{
			Event:          VoiceChannelSettingUpdate,
			GuildID:        e.GuildID,
			ChannelID:      e.ChannelID,
			OriginalUpdate: e,
		}

		return
	}

	// This is a move. Send a VoiceChannelLeave event, then a VoiceChannelJoin event.

	q.Out <- &VoiceStateEvent{
		Event: VoiceChannelLeave,

		// Send previous guild. (Cached)
		GuildID: userState.guildID,

		// Send previous channel. (Cached)
		ChannelID: userState.channelID,

		// Construct an artificial update event. Not perfect, but has the right ChannelID, GuildID, and UserID.
		OriginalUpdate: &discordgo.VoiceStateUpdate{
			VoiceState: &discordgo.VoiceState{
				UserID:    e.UserID,
				ChannelID: userState.channelID,
				GuildID:   userState.guildID,
			},
		},
	}

	q.Out <- &VoiceStateEvent{
		Event:          VoiceChannelJoin,
		GuildID:        e.GuildID,
		ChannelID:      e.ChannelID,
		OriginalUpdate: e,
	}

	q.userVoiceStates[e.UserID] = &userVoiceState{
		guildID:   e.GuildID,
		channelID: e.ChannelID,
	}
}
