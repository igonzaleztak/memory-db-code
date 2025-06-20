package db

var (
	ErrInvalidDataType = NewDBError("invalid data type", "data type must be string or []string")
	ErrDataNotFound    = NewDBError("item not found", "the requested data does not exist in the database")
	ErrKeyHasExpired   = NewDBError("key has expired", "the requested key has expired and is no longer available in the database")
)

type DBerror struct {
	Message    string `json:"message"`
	SysMessage string `json:"-"`
}

func NewDBError(message, sysmessage string) *DBerror {
	return &DBerror{Message: message, SysMessage: sysmessage}
}

// Error returns the error message. This method is required to implement the error interface.
// This way we can return db errors as regular errors.
func (e *DBerror) Error() string {
	return e.Message
}
