package response

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"product-service/internal/errno"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type Meta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type ResultData struct {
	List interface{} `json:"list"`
	Mata *Meta       `json:"meta"`
}

func Success(c *gin.Context, data interface{}) {
	log.Printf("执行成功：%v", data)
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, code int, msg string) {
	log.Printf("执行失败：%d:%v", code, msg)
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

func ErrorWithErrno(c *gin.Context, errno *errno.Error) {
	log.Printf("执行失败：%v", errno.Msg)
	c.JSON(http.StatusOK, Response{
		Code: errno.Code,
		Msg:  errno.Msg,
	})
}
