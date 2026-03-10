package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"product-service/internal/errno"
	"product-service/pkg/response"
	"product-service/pkg/utils"
	"product-service/services/order/service"
)

type OrderCreatedReq struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Count     int    `json:"count" binding:"required"`
	IdemKey   string `json:"idem_key" binding:"required"`
}

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{
		svc: svc,
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var order OrderCreatedReq
	if !utils.BindAndValidateByJSON(c, &order) {
		return
	}
	userId := utils.GetUserID(c)

	o, err := h.svc.Create(c, userId, order.ProductID, order.Count, order.IdemKey)
	if err != nil {
		if errors.Is(err, errno.ProductErrSeckillStockNotEnough) {
			response.ErrorWithErrno(c, errno.ProductErrSeckillStockNotEnough)
			return
		}
		if errors.Is(err, errno.ProductErrSeckillStockNotInit) {
			response.ErrorWithErrno(c, errno.ProductErrSeckillStockNotInit)
			return
		}
		if errors.Is(err, errno.OrderErrOrderAlreadyExist) {
			response.ErrorWithErrno(c, errno.OrderErrOrderAlreadyExist)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, o)
	return
}
