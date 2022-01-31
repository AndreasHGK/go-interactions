package cmd

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

// todo: allow these to be used
type (
	// User is a parameter that allows a user to be input.
	User discord.UserID
	// Role is a parameter that allows the user to provide a certain role in a guild as argument.
	Role discord.RoleID
	// Channel is a parameter that allows the user to provide a certain channel in a guild as argument. This can be
	// all types of channels.
	Channel discord.ChannelID
	// Mentionable is a parameter that allows the user to provide anything that is mentionable as argument. This
	// includes users and roles.
	Mentionable discord.Snowflake
)

// autocompletedParam represents a parameter that can dynamically provide autocompletion hints while the user is typing
// out the parameter.
type autocompletedParam interface {
	Autocomplete(partialString string) []string
	// todo
}

// Optional is a wrapper for any parameter type in order to make it an optional parameter. This allows for the parameter
// to indicate whether it was provided by the user or not.
type Optional[V any] struct {
	val      V
	provided bool
}

// Get returns the underlying value of the parameter. Note that if it was not provided, the nil value will be returned.
func (o Optional[V]) Get() V {
	return o.val
}

// GetOrFallback is a convenience method that returns the underlying value of the parameter if it exists, and if it does
// not exist the provided fallback value will be returned.
func (o Optional[V]) GetOrFallback(fallback V) V {
	if o.provided {
		return o.val
	}
	return fallback
}

// Provided indicates whether the parameter was provided by the user.
func (o Optional[V]) Provided() bool {
	return o.provided
}

// optional is an interface to allow for working with different types of optional parameters.
type optional interface {
	get() any
	set(val any) any
	_optional()
}

func (o Optional[V]) get() any {
	return o.val
}

func (o Optional[V]) set(val any) any {
	o.val = val.(V)
	o.provided = true
	return o
}

func (o Optional[V]) _optional() {}
