package cmd

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

// Response is an interface for responses that can be sent to a user when executing a command.
type Response interface {
	// marshal returns an api.InteractionResponse to send to the discord API.
	marshal() api.InteractionResponse
}

// MessageResponse will respond to the command with a message. The message will indicate that it is a response to a
// certain command ran by the user. It can also be made ephemeral to make it show up only for the sender.
type MessageResponse struct {
	// Content is the text content of a message. It is displayed as regular text. This can be left empty to omit the
	// text, but a message will need at least one of: content, embeds or files.
	Content string
	// Embeds is a slice of the embeds to send with the message. Leave this empty to omit. See: discord.Embed
	Embeds []discord.Embed
	// Files is a slice of files to upload with the message.
	Files []sendpart.File
	// Components is a list of components, such as buttons, that will be under the message itself and can generally be
	// interacted with.
	Components discord.ContainerComponents
	// AllowedMentions controls which users/roles will be mentioned in this message.
	AllowedMentions *api.AllowedMentions
	// Ephemeral decides whether the response is visible for the entire server or just the user. If this is true, only
	// the user will be able to see the message and if false, everyone will be able to see the response. This cannot be
	// edited after a response has been sent.
	Ephemeral bool
	// TTS decides whether the message will be sent as a text-to-speech message.
	TTS bool
}

// ModalResponse is a command response that sends a form to the user that runs this command. This is currently not a
// publicly available option.
// todo: work on this when forms become public
type ModalResponse struct{}

func (m MessageResponse) marshal() (r api.InteractionResponse) {
	if m.Content == "" && len(m.Embeds) == 0 && len(m.Files) == 0 {
		panic("Can't send an empty message response")
	}
	r = api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Files:           m.Files,
			AllowedMentions: m.AllowedMentions,
		},
	}
	if m.Content != "" {
		r.Data.Content = option.NewNullableString(m.Content)
	}
	if len(m.Embeds) > 0 {
		r.Data.Embeds = &m.Embeds
	}
	if m.Ephemeral {
		r.Data.Flags |= api.EphemeralResponse
	}
	return
}
