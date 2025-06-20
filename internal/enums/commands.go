package enums

type DBCommand string

const (
	// DBCommandSet stores an item with the specified key and optional options.
	DBCommandSet DBCommand = "set"
	// DBCommandUpdate modifies an existing item with the specified key and value.
	DBCommandUpdate DBCommand = "update"
	// DBCommandRemove deletes an item by its key.
	DBCommandRemove DBCommand = "remove"
	// DBCommandPush adds a new item to the memory database with the specified key and value.
	DBCommandPush DBCommand = "push"
	// DBCommandPop removes and returns the last item from a slice stored at the specified key.
	DBCommandPop DBCommand = "pop"
)

var MappedCommands = map[string]DBCommand{
	"set":    DBCommandSet,
	"update": DBCommandUpdate,
	"remove": DBCommandRemove,
	"push":   DBCommandPush,
	"pop":    DBCommandPop,
}

// IsValid checks if the command is a valid DBCommand.
func (c DBCommand) IsValid() bool {
	_, exists := MappedCommands[string(c)]
	return exists
}

// String returns the string representation of the DBCommand.
func (c DBCommand) String() string {
	return string(c)
}
