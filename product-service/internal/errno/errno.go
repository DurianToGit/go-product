package errno

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

var (
	OK            = &Error{Code: 0, Msg: "OK"}
	InvalidParams = &Error{Code: 40001, Msg: "Invalid Parameter"}
	Unauthorized  = &Error{Code: 40003, Msg: "Unauthorized"}
	ServerError   = &Error{Code: 50000, Msg: "Internal Server Error"}
)
