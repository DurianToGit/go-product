package productgateway

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"product-service/internal/client/productclient"
	"product-service/internal/errno"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/response"
)

type Handler struct {
	client productclient.Client
}

func NewHandler(client productclient.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	logger.L().Info("gateway get product", zap.Int64("product_id", productID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	p, err := h.client.GetProduct(c.Request.Context(), productID)
	if err != nil {
		logger.L().Error("gateway get product failed", zap.Error(err), zap.Int64("product_id", productID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"id":    p.ID,
		"name":  p.Name,
		"price": p.Price,
		"stock": p.Stock,
	})
}

func (h *Handler) GetStock(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	logger.L().Info("gateway get stock", zap.Int64("product_id", productID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	stock, err := h.client.GetStock(c.Request.Context(), productID)
	if err != nil {
		logger.L().Error("gateway get stock failed", zap.Error(err), zap.Int64("product_id", productID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"product_id": productID,
		"stock":      stock,
	})
}
