package cmd

import (
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"strings"
)

const (
	// nameMinLength, nameMaxLength are the minimum and maximum lengths a command, subcommand or parameter name can be.
	nameMinLength = 1
	nameMaxLength = 32
	// descMinLength, descMaxLength are the minimum and maximum lengths a command, subcommand or parameter description
	// can be.
	descMinLength = 1
	descMaxLength = 100
)

// Command is a slash command that can be registered and executed.
type Command struct {
	name, description string
	defaultEnabled    bool
	guild             discord.GuildID

	executor    Executor
	subcommands map[string]Subcommand
	subGroups   map[string]string // name: description

	registered bool
}

// New creates a new slash command. By itself it will not do anything, and needs executors to be runnable.
func New(name, description string) Command {
	mustMatch(name)
	if len(name) < nameMinLength || len(name) > nameMaxLength {
		panic(fmt.Sprintf("command name must be equal to or between %v and %v characters in length", nameMinLength, nameMaxLength))
	} else if len(description) < descMinLength || len(description) > descMaxLength {
		panic(fmt.Sprintf("command description must be equal to or between %v and %v characters in length", descMinLength, descMaxLength))
	}

	return Command{
		name:        name,
		description: description,
		subcommands: map[string]Subcommand{},
		subGroups:   map[string]string{},

		guild: discord.NullGuildID,

		defaultEnabled: true,
	}
}

// WithExecutor returns the command with the executor provided. This will be the main executor for the command. If you
// only want subcommands, this does not need to be provided.
func (c Command) WithExecutor(e Executor) Command {
	if len(c.subcommands) > 0 {
		panic("Subcommands and main executor are mutually exclusive")
	}
	c.executor = e
	return c
}

// WithSubcommandGroup adds a new subcommand group to the command. This is essentially a folder for subcommands, and
// allow for "double" subcommands: /<command> <subcommandGroup> <subcommand>
func (c Command) WithSubcommandGroup(name, description string) Command {
	c.subGroups[name] = description
	return c
}

// WithSubcommand adds a new subcommand to the command. It can be either a subcommand directly within the command
// itself, or a subcommand within a subcommand group. To do the latter, you will need to enter the name of your
// subcommand as a single string: "subcommand_group subcommand". To use subcommand groups, they must first be added
// using Command.WithSubcommandGroup.
func (c Command) WithSubcommand(fullName, description string, e Executor) Command {
	if c.executor != nil {
		panic("Subcommands and main executor are mutually exclusive")
	}

	if len(description) < descMinLength || len(description) > descMaxLength {
		panic(fmt.Sprintf("subcommand description must be equal to or between %v and %v characters in length", descMinLength, descMaxLength))
	}

	name := fullName
	var groupName string

	// Check if the full subcommand name contains both a subcommand name and subcommand group
	if names := strings.SplitN(fullName, " ", 2); len(names) > 1 {
		name = names[1]
		groupName = names[0]
		mustMatch(groupName)

		// The subcommand group must first be registered
		if _, ok := c.subGroups[groupName]; !ok {
			panic(fmt.Sprintf("Non-existent subcommand group: %s", groupName))
		}
	}
	mustMatch(name)
	if len(name) < nameMinLength || len(name) > nameMaxLength {
		panic(fmt.Sprintf("subcommand name must be equal to or between %v and %v characters in length", nameMinLength, nameMaxLength))
	} else if groupName != "" && (len(groupName) < nameMinLength || len(groupName) > nameMaxLength) {
		panic(fmt.Sprintf("subcommand group name must be equal to or between %v and %v characters in length", nameMinLength, nameMaxLength))
	}

	c.subcommands[fullName] = Subcommand{
		name:        name,
		description: description,
		group:       groupName,
		executor:    e,
	}
	return c
}

// WithoutDefaultEnabled will make it so users do not have permission to execute this command by default when the bot is
// added to a guild. It will still be available to admins, and you are able to give permission
func (c Command) WithoutDefaultEnabled() Command {
	c.defaultEnabled = false
	return c
}

// Name is what the player will type to execute the slash command: /<name>
func (c Command) Name() string {
	return c.name
}

// Description is a short but descriptive message about what the command is supposed to do.
func (c Command) Description() string {
	return c.description
}

// Guild returns the ID of the guild to which the command is registered. If it is a global command, this will be equal
// to discord.NullGuildID
func (c Command) Guild() discord.GuildID {
	return c.guild
}

// Subcommands returns all registered subcommands of the command.
func (c Command) Subcommands() (subcommands []Subcommand) {
	for _, sub := range c.subcommands {
		subcommands = append(subcommands, sub)
	}
	return
}

// marshal will generate the command with all it's parameters, so it is ready to be sent through the discord API.
func (c Command) marshal() api.CreateCommandData {
	options := discord.CommandOptions{}
	groups := map[string]*discord.SubcommandGroupOption{}
	if c.executor != nil {
		for _, opt := range makeCommandOptions(c.executor) {
			options = append(options, opt)
		}
	} else {
		for _, subcommand := range c.subcommands {
			sub := &discord.SubcommandOption{
				OptionName:  subcommand.name,
				Description: subcommand.description,
				Required:    false, // ???
				Options:     makeCommandOptions(subcommand.executor),
			}

			if subcommand.group != "" {
				if group, ok := groups[subcommand.group]; ok {
					group.Subcommands = append(group.Subcommands, sub)
				} else {
					group = &discord.SubcommandGroupOption{
						OptionName:  subcommand.group,
						Description: c.subGroups[subcommand.group],
						Required:    false,
						Subcommands: []*discord.SubcommandOption{sub},
					}
					options = append(options, group)
					groups[subcommand.group] = group
				}
			} else {
				options = append(options, sub)
			}
		}
	}

	return api.CreateCommandData{
		Type:        discord.ChatInputCommand,
		Name:        c.name,
		Description: c.description,
		Options:     options,
	}
}
