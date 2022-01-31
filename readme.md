# go-interactions

A library that aims to make dealing with discord's slash commands easy.
It is designed for [diamondburned/arikawa](https://github.com/diamondburned/arikawa),
and based on the command system from [df-mc/dragonfly](https://github.com/df-mc/dragonfly).

### Note

This library is currently not quite ready for production use.
There might be bugs, and while I do try to test everything there could always be some edge cases.
The library is also not feature complete yet.
Currently, there is no way to reliably use this across multiple nodes.
Other currently missing features include command parameter options and autocompleted params.

## Usage

To use this library you will need to have at least Go 1.18 installed.
To use this in your project, simply run the following command:
```
go get github.com/AndreasHGK/go-interactions
```
The following code is an example command made using this library, which will greet the user provided as a parameter.
```go
import (
    "fmt"
    "github.com/andreashgk/go-interactions/cmd"
    "time"
)

// Greet is an example command executor. It will greet a user after the delay specified, otherwise send it in one
// second. Parameters can be specified using the fields in the struct.
type Greet struct {
    // The discord api allows you to add up to 25 parameters per executor. These parameters can be any int, float,
    // string or bool type and can also be of type cmd.User, cmd.Channel, cmd.Mentionable, cmd.Role. The description
    // should be included for every parameter like shown here.
    Target cmd.User `description:"The person to greet"`
    
    // Optional parameters can be added like shown for this. A cmd.Optional[] needs to be wrapped around the parameter
    // type. It has a few methods to get the underlying value and to return whether the value was provided. All optional
    // parameters have to be provided after all required parameters.
    Delay cmd.Optional[int] `description:"Whether or not to show the response publicly"`
}

// Run will be called when the command is executed by the player. All parameter values will be set inside the struct,
// and a *cmd.Interaction is passed to allow for getting values such as the sender and has methods to send responses.
func (u Greet) Run(interaction *cmd.Interaction) {
    // Sends a "DeferredMessageInteractionWithSource" response. This indicates that the bot has received the command and
    // a message response will follow within the next 15 minutes.
    followup, err := interaction.DeferResponse()
    if err != nil {
        fmt.Printf("Error sending command response: %s", err.Error())
        return
    }
    
    go func() {
        time.Sleep(time.Duration(u.Delay.GetOrFallback(1)) * time.Second)
        
        // The followup can be used to send followup responses. Currently, these can only be messages. They will show
        // as responses to the original response.
        _, err := followup.Create(cmd.MessageResponse{
            Content: fmt.Sprintf("Hello, <@%v>", u.Target),
        })
        if err != nil {
            fmt.Printf("Error sending followup message: %s", err.Error())
        }
    }()
}
```

The command will then have to be registered to a command handler.
This can be done as follows:
```go
func RegisterCommands(botState *state.State) {
	// Create a new command handler.
	h := cmd.NewHandler(nil).WithCommands(
		cmd.New("greet", "Greet another user.").WithExecutor(Greet{}),
	)
	// Will register the commands to a specific guild. This clears the list of commands that are pending to be
	// registered.
	h.RegisterAllGuild(botState, discord.GuildID(12345))
	// Adds the interaction event handler to the bot.
	h.Listen(botState)
}
```
This is all that is needed to add your commands!