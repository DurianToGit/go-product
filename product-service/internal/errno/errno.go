package errno

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

var (
	OK            = &Error{Code: 0, Msg: "success"}
	InvalidParams = &Error{Code: 40001, Msg: "Invalid Parameter"}
	Unauthorized  = &Error{Code: 40003, Msg: "Unauthorized"}

	ErrDataNotFound = &Error{Code: 40008, Msg: "Data Not Found"}
	// too many requests
	ErrTooManyRequests = &Error{Code: 40009, Msg: "Too Many Requests"}

	ProductErrNoEnoughStock = &Error{Code: 40101, Msg: "No Enough Stock"}
	// 产品名称不能为空
	ProductErrNameEmpty = &Error{Code: 40102, Msg: "Product Name Empty"}
	// invalid stock count
	ProductErrInvalidStock = &Error{Code: 40103, Msg: "Invalid Stock Count"}
	// stock not enough or retry limit reached
	ProductErrStockNotEnough  = &Error{Code: 40104, Msg: "stock not enough or retry limit reached"}
	ProductErrStockLockFailed = &Error{Code: 40105, Msg: "stock lock failed"}

	UsernameAlreadyExist     = &Error{Code: 40201, Msg: "Username Already Exist"}
	UsernameNotFound         = &Error{Code: 40202, Msg: "Username Not Found"}
	UserOldPasswordIncorrect = &Error{Code: 40203, Msg: "Old Password Incorrect"}
	UserErrPasswordIncorrect = &Error{Code: 40204, Msg: "Password Incorrect"}
	UserErrNotFound          = &Error{Code: 40006, Msg: "User Not Found"}

	ServerError = &Error{Code: 50000, Msg: "Internal Server Error"}
)
