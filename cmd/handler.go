package cmd

import (
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
	"reflect"
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Handler contains all commands and handles interactions for these commands. The Handler is suitable for concurrent
// use.
type Handler struct {
	commandsMu      sync.RWMutex
	commands        map[discord.CommandID]Command
	pendingCommands map[string]Command

	logger Logger
}

// NewHandler returns a pointer to a new Handler. This can be used to register & handle commands.
func NewHandler(logger Logger) *Handler {
	return &Handler{
		commands:        map[discord.CommandID]Command{},
		pendingCommands: map[string]Command{},

		logger: logger,
	}
}

// WithCommands registers one or multiple commands to the handler.
func (h *Handler) WithCommands(commands ...Command) *Handler {
	h.commandsMu.Lock()
	for _, cmd := range commands {
		h.pendingCommands[cmd.name] = cmd
	}
	h.commandsMu.Unlock()

	return h
}

// RegisterAll will globally register all currently unregistered commands. When commands are modified, this can take up
// to an hour to update in guilds. Doing this will remove all other global commands not currently pending in this
// handler.
func (h *Handler) RegisterAll(api API) error {
	return h.RegisterAllGuild(api, discord.NullGuildID)
}

// RegisterAllGuild will register all currently pending commands in a specific guild. Commands will be updated instantly
// in the guild in question. This will however remove all other guild commands not registered in this batch.
func (h *Handler) RegisterAllGuild(discordAPI API, guildId discord.GuildID) error {
	// Skip registering commands if there are no commands to register
	if len(h.pendingCommands) == 0 {
		return nil
	}
	app, err := discordAPI.CurrentApplication()
	if err != nil {
		return err
	}

	var cmds []api.CreateCommandData

	h.commandsMu.Lock()
	defer h.commandsMu.Unlock()
	for _, cmd := range h.pendingCommands {
		cmds = append(cmds, cmd.marshal())
	}

	var registeredCommands []discord.Command
	if guildId == discord.NullGuildID {
		registeredCommands, err = discordAPI.BulkOverwriteCommands(app.ID, cmds)
	} else {
		registeredCommands, err = discordAPI.BulkOverwriteGuildCommands(app.ID, guildId, cmds)
	}
	if err != nil {
		return err
	}

	for _, registeredCmd := range registeredCommands {
		cmd := h.pendingCommands[registeredCmd.Name]
		cmd.guild = guildId
		cmd.registered = true

		h.commands[registeredCmd.ID] = cmd
	}
	h.pendingCommands = map[string]Command{}
	return nil
}

// Listen registers a listen function.
func (h *Handler) Listen(api API) error {
	app, err := api.CurrentApplication()
	if err != nil {
		return err
	}
	appId := app.ID

	handler := func(event *gateway.InteractionCreateEvent) {
		commandEvent, ok := event.Data.(*discord.CommandInteraction)
		if !ok {
			// Only handle command interactions
			return
		}

		// Command fetching
		// ----------------
		// This section handles the looking for the correct command to execute, and also the right command executor.
		var executor Executor
		var options discord.CommandInteractionOptions
		{
			// Get the command with the correct id
			h.commandsMu.RLock()
			command, ok := h.commands[commandEvent.ID]
			h.commandsMu.RUnlock()
			if !ok {
				return
			}

			// Get the right executor for the command. A command can either only have a main executor, or only
			// subcommand executors. Also get the correct command options.
			options = commandEvent.Options
			if command.executor == nil {
				// Get the first parameter. This will be the subcommand name or, if applicable, the subcommand group it
				// is in.
				subOpt := options[0]
				subName := subOpt.Name
				options = subOpt.Options

				// If a subcommand group with this name exists, get the full subcommand name
				// ("subcommandGroup subcommand")
				if _, ok = command.subGroups[subName]; ok {
					subOpt2 := options[0]
					options = subOpt2.Options

					subName += " " + subOpt2.Name
				}

				subCmd, ok := command.subcommands[subName]
				if !ok {
					return
				}
				executor = subCmd.executor
			} else {
				executor = command.executor
			}
		}

		// Command parameterization
		// ------------------------
		// In this section, a new instance of the right executor will be created and all parameters will be set.
		{
			refl := reflect.New(reflect.TypeOf(executor)).Elem()
			for _, option := range options {
				field := refl.FieldByNameFunc(func(s string) bool {
					return option.Name == strings.ToLower(s)
				})
				if !field.IsValid() {
					panic(fmt.Sprintf("Field %s not valid", field.String()))
				}

				// Determine what to cast the command option to depending on the parameter type
				field.Set(reflect.ValueOf(func() interface{} {
					instance := field.Interface()

					actualType := instance
					if opt, ok := instance.(optional); ok {
						actualType = opt.get()
					}

					var r interface{}
					var err error

					switch actualType.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
						var temp int64
						temp, err = option.IntValue()
						// Determine which type of int the type actually is, and cast it
						switch actualType.(type) {
						case int:
							r = int(temp)
						case uint:
							r = uint(temp)
						case int8:
							r = int8(temp)
						case uint8:
							r = uint8(temp)
						case int16:
							r = int16(temp)
						case uint16:
							r = uint16(temp)
						case int32:
							r = int32(temp)
						case uint32:
							r = uint32(temp)
						case int64:
							r = int64(temp)
						case uint64:
							r = uint64(temp)
						}
					case float32, float64:
						var temp float64
						temp, err = option.FloatValue()
						if _, ok := actualType.(float32); ok {
							r = float32(temp)
						} else {
							r = temp
						}
					case string:
						r = option.String()
					case bool:
						r, err = option.BoolValue()

					case User, Role, Channel, Mentionable:
						var temp discord.Snowflake
						temp, err = option.SnowflakeValue()
						switch actualType.(type) {
						case User:
							r = User(temp)
						case Role:
							r = Role(temp)
						case Channel:
							r = Channel(temp)
						case Mentionable:
							r = Mentionable(temp)
						}

					default:
						panic(fmt.Sprintf("Unrecognized parameter type: %s", field.Type()))
					}

					if err != nil {
						panic(err)
					}
					if opt, ok := instance.(optional); ok {
						r = opt.set(r)
					}
					return r
				}()))
			}
			// Set the actual executor
			executor = refl.Interface().(Executor)
		}

		// Command execution
		// -----------------
		// This section executes the command and creates the cmd.Interaction, which contains extra parameters such as
		// the sender and allows for the executor to send responses back to discord.
		executor.Run(&Interaction{
			api:   api,
			appId: appId,

			interactionId:    event.ID,
			interactionToken: event.Token,

			guildId:   event.GuildID,
			channelId: event.ChannelID,
			member:    event.Member,
			user:      event.Sender(),
		})
	}
	api.AddHandler(handler)
	return nil
}
