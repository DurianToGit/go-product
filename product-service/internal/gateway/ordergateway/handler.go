package ordergateway

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"product-service/internal/client/orderclient"
	"product-service/internal/errno"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/response"
)

type Handler struct {
	client orderclient.Client
}

func NewHandler(client orderclient.Client) *Handler {
	return &Handler{client: client}
}

type CreateOrderReq struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Count     int64  `json:"count" binding:"required,min=1"`
	IdemKey   string `json:"idem_key" binding:"required"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		response.ErrorWithErrno(c, errno.Unauthorized)
		return
	}

	logger.L().Info("gateway create order",
		zap.Int64("user_id", uid),
		zap.Int64("product_id", req.ProductID),
		zap.Int64("count", req.Count),
		zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())),
	)

	order, err := h.client.Create(c.Request.Context(), uid, req.ProductID, req.Count, req.IdemKey)
	if err != nil {
		logger.L().Error("gateway create order failed", zap.Error(err), zap.Int64("user_id", uid))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"id":         order.OrderId,
		"order_no":   order.OrderNo,
		"user_id":    order.UserId,
		"product_id": order.ProductId,
		"count":      order.Count,
		"amount":     order.Amount,
		"status":     order.Status,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway get order", zap.Int64("order_id", orderID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	order, err := h.client.Get(c.Request.Context(), orderID)
	if err != nil {
		logger.L().Error("gateway get order failed", zap.Error(err), zap.Int64("order_id", orderID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"id":         order.OrderId,
		"order_no":   order.OrderNo,
		"user_id":    order.UserId,
		"product_id": order.ProductId,
		"count":      order.Count,
		"amount":     order.Amount,
		"status":     order.Status,
	})
}

type CancelOrderReq struct {
	Reason string `json:"reason"`
}

func (h *Handler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	var req CancelOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
	}

	logger.L().Info("gateway cancel order", zap.Int64("order_id", orderID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	if err := h.client.Cancel(c.Request.Context(), orderID, req.Reason); err != nil {
		logger.L().Error("gateway cancel order failed", zap.Error(err), zap.Int64("order_id", orderID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) PayOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway pay order", zap.Int64("order_id", orderID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	if err := h.client.Pay(c.Request.Context(), orderID); err != nil {
		logger.L().Error("gateway pay order failed", zap.Error(err), zap.Int64("order_id", orderID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, nil)
}
