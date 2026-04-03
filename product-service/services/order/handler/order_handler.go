package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"product-service/internal/errno"
	"product-service/pkg/logger"
	"product-service/pkg/response"
	"product-service/pkg/utils"
	"product-service/services/order/service"
	"strconv"
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

func (h *OrderHandler) Pay(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	err = h.svc.Pay(c, id)
	if err != nil {
		if errors.Is(err, errno.OrderNotFound) {
			response.ErrorWithErrno(c, errno.OrderNotFound)
			return
		}
		if errors.Is(err, errno.OrderStatusInvalid) {
			response.ErrorWithErrno(c, errno.OrderStatusInvalid)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, nil)
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), orderID); err != nil {
		if errors.Is(err, errno.OrderNotFound) {
			response.ErrorWithErrno(c, errno.OrderNotFound)
			return
		}
		if errors.Is(err, errno.OrderStatusInvalid) {
			response.ErrorWithErrno(c, errno.OrderStatusInvalid)
			return
		}
		logger.L().Error("取消订单失败", zap.Error(err))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"order_id": orderID,
		"status":   "canceled",
	})
}
