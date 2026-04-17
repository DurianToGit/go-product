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
	// user rate limit exceeded
	ErrUserRateLimitExceeded = &Error{Code: 40010, Msg: "User Rate Limit Exceeded"}
	// 依赖不可用
	ErrDependencyUnavailable = &Error{Code: 40011, Msg: "Dependency Unavailable"}

	ProductErrNoEnoughStock = &Error{Code: 40101, Msg: "No Enough Stock"}
	// 产品名称不能为空
	ProductErrNameEmpty = &Error{Code: 40102, Msg: "Product Name Empty"}
	// invalid stock count
	ProductErrInvalidStock = &Error{Code: 40103, Msg: "Invalid Stock Count"}
	// stock not enough or retry limit reached
	ProductErrStockNotEnough  = &Error{Code: 40104, Msg: "stock not enough or retry limit reached"}
	ProductErrStockLockFailed = &Error{Code: 40105, Msg: "stock lock failed"}
	// 秒杀库存未初始化
	ProductErrSeckillStockNotInit = &Error{Code: 40106, Msg: "秒杀库存未初始化"}
	// 秒杀库存不足
	ProductErrSeckillStockNotEnough = &Error{Code: 40107, Msg: "秒杀库存不足"}
	// 恢复库存失败
	ProductErrRestoreStockFailed              = &Error{Code: 40108, Msg: "恢复库存失败"}
	ProductErrConsumeStockDeductEventFailed   = &Error{Code: 40109, Msg: "消费库存扣减事件失败"}
	ProductErrConsumeRestockDeductEventFailed = &Error{Code: 40110, Msg: "消费恢复库存扣减事件失败"}

	UsernameAlreadyExist     = &Error{Code: 40201, Msg: "Username Already Exist"}
	UsernameNotFound         = &Error{Code: 40202, Msg: "Username Not Found"}
	UserOldPasswordIncorrect = &Error{Code: 40203, Msg: "Old Password Incorrect"}
	UserErrPasswordIncorrect = &Error{Code: 40204, Msg: "Password Incorrect"}
	UserErrNotFound          = &Error{Code: 40205, Msg: "User Not Found"}

	// 订单已存在,但是订单数据不一致
	OrderErrOrderAlreadyExist = &Error{Code: 40306, Msg: "订单已存在,但是订单信息不一致"}
	OrderErrNotEnoughStock    = &Error{Code: 40307, Msg: "库存不足"}
	OrderNotFound             = &Error{Code: 40308, Msg: "订单不存在"}
	OrderStatusInvalid        = &Error{Code: 40309, Msg: "订单状态无效"}

	ServerError = &Error{Code: 50000, Msg: "Internal Server Error"}
)
