package cmd

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"sync"
)

// Followup ...
type Followup struct {
	api              API
	appId            discord.AppID
	interactionToken string

	locker sync.Mutex
}

// Create ...
func (f *Followup) Create(response Response) (*discord.Message, error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	return f.api.CreateInteractionFollowup(f.appId, f.interactionToken, *response.marshal().Data)
}

// Edit ...
func (f *Followup) Edit(messageId discord.MessageID, editedData api.EditInteractionResponseData) (*discord.Message, error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	return f.api.EditInteractionFollowup(f.appId, messageId, f.interactionToken, editedData)
}

// Delete ...
func (f *Followup) Delete(messageId discord.MessageID) error {
	f.locker.Lock()
	defer f.locker.Unlock()

	return f.api.DeleteInteractionFollowup(f.appId, messageId, f.interactionToken)
}
