package cmd

// Subcommand ...
type Subcommand struct {
	name, group, description string

	executor Executor
}

// Name is the name of the subcommand, and what the user will have to type to execute it. To execute it, you will need
// to run /<commandName> <subcommand> ...
func (s Subcommand) Name() string {
	return s.name
}

// Group determines what subcommand group the subcommand will be in. These are kind of like folders for subcommands, and
// to execute these nested subcommands you need to do /<commandName> <subcommandGroup> <subcommand> ...
func (s Subcommand) Group() string {
	return s.group
}

// Description is a short, descriptive message about what the subcommand does.
func (s Subcommand) Description() string {
	return s.description
}
