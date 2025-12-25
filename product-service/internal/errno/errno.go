package errno

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

var (
	OK                        = &Error{Code: 0, Msg: "OK"}
	InvalidParams             = &Error{Code: 40001, Msg: "Invalid Parameter"}
	Unauthorized              = &Error{Code: 40003, Msg: "Unauthorized"}
	ErrPasswordIncorrect      = &Error{Code: 40004, Msg: "Password Incorrect"}
	ErrDatabaseNotInitialized = &Error{Code: 40005, Msg: "Database Not Initialized"}
	ErrUserNotFound           = &Error{Code: 40006, Msg: "User Not Found"}
	OldPasswordIncorrect      = &Error{Code: 40007, Msg: "Old Password Incorrect"}
	ErrDataNotFound           = &Error{Code: 40008, Msg: "Data Not Found"}

	ProductErrNoEnoughStock = &Error{Code: 40101, Msg: "No Enough Stock"}

	ServerError = &Error{Code: 50000, Msg: "Internal Server Error"}
)
