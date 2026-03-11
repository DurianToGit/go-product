package dto

type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type UserQuery struct {
	Keyword  string
	Status   *int
	Page     int
	PageSize int
}

type UserUpdate struct {
	Username *string
	Password *string
}
