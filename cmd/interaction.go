package cmd

import (
	"errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"go.uber.org/atomic"
)

// Interaction will be passed to the command executor when a command is executed by a user. It contains details about
// the command, such as the user who sent it, the channel in which it got sent and a guild ID if the command was
// executed within a channel in a guild. If that is not the case, a discord.NullGuildID will be provided instead.
// Interaction responses can also be sent with Interaction.Respond and Interaction.DeferResponse.
type Interaction struct {
	api   API
	appId discord.AppID

	interactionId    discord.InteractionID
	interactionToken string

	guildId   discord.GuildID
	channelId discord.ChannelID
	member    *discord.Member
	user      *discord.User

	hasResponded atomic.Bool
}

// API returns the underlying discord API used. The command executor can use this to perform additional actions not
// supported by the interaction instance itself.
func (i *Interaction) API() API {
	return i.api
}

// Respond sends a MessageResponse, which is a normal message response. Multiple message responses can be sent to the
// interaction. When the response has been sent, the message will also be returned and can be edited or deleted.
func (i *Interaction) Respond(response MessageResponse) (*Followup, error) {
	if !i.hasResponded.CAS(false, true) {
		panic("cannot send multiple responses to the same interaction")
	}

	err := i.api.RespondInteraction(i.interactionId, i.interactionToken, response.marshal())
	if err != nil {
		return nil, err
	}

	// Create and return a *cmd.Followup, which can be used to send followup responses to the interaction.
	return &Followup{
		api:              i.api,
		appId:            i.appId,
		interactionToken: i.interactionToken,
	}, nil
}

// Response returns the message sent to the interaction as response. This assumes that the response sent to the discord
// api was a message response, and not a deferred message response.
func (i *Interaction) Response() (*discord.Message, error) {
	if !i.hasResponded.Load() {
		return nil, errors.New("an interaction response has not yet been created")
	}

	return i.api.InteractionResponse(i.appId, i.interactionToken)
}

// EditResponse edits the original response to the command. This will only work if the response was a message response
// and not a deferred response. The message also still needs to exist in order for this to work.
// todo: do not expose arikawa type?
func (i *Interaction) EditResponse(editedResponse api.EditInteractionResponseData) (*discord.Message, error) {
	if !i.hasResponded.Load() {
		return nil, errors.New("an interaction response has not yet been created")
	}

	return i.api.EditInteractionResponse(i.appId, i.interactionToken, editedResponse)
}

// DeleteResponse will delete the original response to the command. This can naturally only be done once and if a
// message response has been sent previously.
func (i *Interaction) DeleteResponse() error {
	if !i.hasResponded.Load() {
		return errors.New("an interaction response has not yet been created")
	}

	return i.api.DeleteInteractionResponse(i.appId, i.interactionToken)
}

// DeferResponse sends an api.DeferredMessageInteractionWithSource to discord, which is a response that acknowledges
// that the command has been received by the bot and a message response will be sent later. The user will keep seeing
// the loading state of the command until that message gets sent. Sending this will allow more time than the standard
// short period you have after a command is sent. DeferResponse must not be called after the first time it is called.
func (i *Interaction) DeferResponse() (*Followup, error) {
	if !i.hasResponded.CAS(false, true) {
		panic("cannot send multiple responses to the same interaction")
	}

	err := i.api.RespondInteraction(i.interactionId, i.interactionToken, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
	})
	if err != nil {
		return nil, err
	}

	// Create and return a *cmd.Followup, which can be used to send followup responses to the interaction.
	return &Followup{
		api:              i.api,
		appId:            i.appId,
		interactionToken: i.interactionToken,
	}, nil
}

// User returns the *discord.User who executed the command.
func (i *Interaction) User() *discord.User {
	return i.user
}

// Member returns the *discord.Member who executed the command. Unlike with User(), this may be equal to nil if the
// command was not executed in a guild.
func (i *Interaction) Member() *discord.Member {
	return i.member
}

// ChannelID returns the channel in which the command was sent. Logically, this will be a channel where typing is
// possible.
func (i *Interaction) ChannelID() discord.ChannelID {
	return i.channelId
}

// InGuild returns whether the interaction originated from within a guild.
func (i *Interaction) InGuild() bool {
	return i.guildId != discord.NullGuildID
}

// GuildID returns the discord.GuildID of the guild where the command has been executed. This will be equal to
// discord.NullGuildID if not applicable.
func (i *Interaction) GuildID() discord.GuildID {
	return i.guildId
}
