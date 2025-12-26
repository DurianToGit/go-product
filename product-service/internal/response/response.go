package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"product-service/internal/errno"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type ResultData struct {
	List     interface{}
	Total    int64
	Page     int
	PageSize int
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

func ErrorWithErrno(c *gin.Context, errno *errno.Error) {
	c.JSON(http.StatusOK, Response{
		Code: errno.Code,
		Msg:  errno.Msg,
	})
}
