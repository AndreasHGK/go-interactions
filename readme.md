# go-interactions

A library that aims to make dealing with discord's slash commands easy.
It is designed for [diamondburned/arikawa](https://github.com/diamondburned/arikawa).

## Usage

To use this library you will need to have at least Go 1.18 installed.
To use this in your project, simply run the following command:
```
go get github.com/AndreasHGK/go-interactions
```
The following code is an example command made using this library. For a more detailed guide, visit the wiki *(todo)*.
```go
// Greet is an example command executor. Parameters can be specified using the fields in the struct.
type Greet struct {
	Target    cmd.User           `description:"The person to greet"`
	Delay cmd.Optional[int] `description:"Whether or not to show the response publicly"`
}

// Run will be called when the command is ran by the player. 
func (u Greet) Run(interaction *cmd.Interaction) {
	_, err := interaction.Respond(cmd.MessageResponse{
		Content: fmt.Sprintf("Hello, <@%v>", u.Target),
	})
	// todo: handle error
}
```