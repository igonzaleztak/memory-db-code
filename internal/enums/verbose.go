package enums

type VerboseLevel string

const (
	VerboseLevelDebug VerboseLevel = "debug"
	VerboseLevelInfo  VerboseLevel = "info"
)

var MapVerboseLevel = map[VerboseLevel]string{
	VerboseLevelDebug: "debug",
	VerboseLevelInfo:  "info",
}

// String returns the string representation of the VerboseLevel.
func (v VerboseLevel) String() string {
	return MapVerboseLevel[v]
}

// IsValid checks if the VerboseLevel is a valid value.
func (v VerboseLevel) IsValid() bool {
	_, ok := MapVerboseLevel[v]
	return ok
}
