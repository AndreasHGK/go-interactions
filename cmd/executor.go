package cmd

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// Executor is an interface that you can implement to handle execution of a certain (sub)command.
type Executor interface {
	// Run is the function that will be called by the command handler when the command is being executed by a user.
	// A cmd.Interaction is passed along to allow for the executor to send responses to the command and access various
	// parameters, such as the command sender or the channel and guild it was sent it.
	Run(interaction *Interaction)
}

// GuildOnly is a struct that can be embedded in a command handler to make it runnable only in guilds. When the command
// is ran in the bot's direct messages, a response containing an error message will be sent, telling the user
type GuildOnly struct {
	api API
}

// makeCommandOptions determines all parameters for a given Executor. The executor must be a struct.
func makeCommandOptions(e Executor) []discord.CommandOptionValue {
	refl := reflect.TypeOf(e)
	if refl.Kind() != reflect.Struct {
		panic("command executor must be a struct")
	}

	var isLastOptional bool
	var opts []discord.CommandOptionValue

	for i := 0; i < refl.NumField(); i++ {
		field := refl.Field(i)
		if !field.IsExported() || field.Anonymous {
			continue
		}

		instance := reflect.New(field.Type).Elem().Interface()

		name := strings.ToLower(field.Name)
		desc := field.Tag.Get("description")

		if len(name) < nameMinLength || len(name) > nameMaxLength {
			panic(fmt.Sprintf("parameter name must be equal to or between %v and %v characters in length", nameMinLength, nameMaxLength))
		} else if len(desc) < descMinLength || len(desc) > descMaxLength {
			panic(fmt.Sprintf("parameter description must be equal to or between %v and %v characters in length", descMinLength, descMaxLength))
		}

		var isOptional bool
		if p, ok := instance.(optional); ok {
			isOptional = true
			instance = p.get()
		}

		// verify the order of optional/required parameters
		if isOptional {
			isLastOptional = true
		} else if isLastOptional {
			panic("non-optional command parameters must be provided before all optional parameters")
		}

		// The following anonymous function will create the discord.CommandOptionValue for the current parameter.
		opts = append(opts, func(name, desc string, isOptional bool, instance interface{}) (opt discord.CommandOptionValue) {
			switch instance.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				var min option.Int
				var max option.Int
				switch instance.(type) {
				case uint:
					min = option.NewInt(0)
				case int8:
					min = option.NewInt(math.MinInt8)
					max = option.NewInt(math.MaxInt8)
				case uint8:
					min = option.NewInt(0)
					max = option.NewInt(math.MaxUint8)
				case int16:
					min = option.NewInt(math.MinInt16)
					max = option.NewInt(math.MaxInt16)
				case uint16:
					min = option.NewInt(0)
					max = option.NewInt(math.MaxUint16)
				case int32:
					min = option.NewInt(math.MinInt32)
					max = option.NewInt(math.MaxInt32)
				case uint32:
					min = option.NewInt(0)
					max = option.NewInt(math.MaxUint32)
				case uint64:
					min = option.NewInt(0)
				}
				opt = &discord.IntegerOption{
					OptionName:  name,
					Description: desc,
					Required:    !isOptional,
					Max:         max,
					Min:         min,
				}
			case float32, float64:
				opt = &discord.NumberOption{
					OptionName:  name,
					Description: desc,
					Required:    !isOptional,
					Min:         option.NewFloat(-math.Pow(2, 53)),
					Max:         option.NewFloat(math.Pow(2, 53)),
				}
			case string:
				opt = discord.NewStringOption(name, desc, !isOptional)
			case bool:
				opt = discord.NewBooleanOption(name, desc, !isOptional)

			case User:
				opt = discord.NewUserOption(name, desc, !isOptional)
			case Role:
				opt = discord.NewRoleOption(name, desc, !isOptional)
			case Mentionable:
				opt = discord.NewMentionableOption(name, desc, !isOptional)
			case Channel:
				opt = discord.NewChannelOption(name, desc, !isOptional) // todo: channel types

			default:
				panic(fmt.Sprintf("unrecognized command parameter type: %s", reflect.TypeOf(instance).String()))
			}
			return
		}(name, desc, isOptional, instance)) // Execute the function
	}
	return opts
}
