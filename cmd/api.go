package cmd

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

// API represents an instance of an object that has connection to the gateway.
type API interface {
	CurrentApplication() (*discord.Application, error)

	AddHandler(handler interface{}) (rm func())

	RespondInteraction(id discord.InteractionID, token string, resp api.InteractionResponse) error
	InteractionResponse(appID discord.AppID, token string) (*discord.Message, error)
	EditInteractionResponse(appID discord.AppID, token string, data api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionResponse(appID discord.AppID, token string) error

	CreateInteractionFollowup(appID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error)
	EditInteractionFollowup(appID discord.AppID, messageID discord.MessageID, token string, data api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionFollowup(appID discord.AppID, messageID discord.MessageID, token string) error

	BulkOverwriteCommands(appID discord.AppID, commands []api.CreateCommandData) ([]discord.Command, error)
	BulkOverwriteGuildCommands(appID discord.AppID, guildID discord.GuildID, commands []api.CreateCommandData) ([]discord.Command, error)
}
