package handler

import (
	"errors"
	"fmt"
	"product-service/internal/domain"
	"product-service/internal/errno"
	"product-service/internal/response"
	"product-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc}
}

func (h *ProductHandler) List(c *gin.Context) {
	var req domain.ListReq
	if !BindAndValidateByQuery(c, &req) {
		return
	}
	fmt.Println("参数page=", req.Page, ";pageSize=", req.PageSize)

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	data, err := h.svc.GetProducts(c, req.Page, req.PageSize)
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, data)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req domain.Product
	if !BindAndValidateByJSON(c, &req) {
		return
	}

	if err := h.svc.CreateProduct(c, &req); err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, req)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetProduct(c, int64(id))

	if errors.Is(err, errno.ErrDataNotFound) {
		response.ErrorWithErrno(c, errno.ErrDataNotFound)
		return
	}

	if err != nil {
		response.Error(c, 40001, err.Error())
		return
	}

	response.Success(c, data)
}

func (h *ProductHandler) DuctStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	count, _ := strconv.Atoi(c.Query("count"))
	err := h.svc.DeductStock(c, int64(id), int64(count))
	if err != nil {
		response.Error(c, 40000, "DeductStock Failed")
		return
	}
	response.Success(c, "DeductStock Success")
}

func (h *ProductHandler) DuctStockOptimistic(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	count, _ := strconv.Atoi(c.Query("count"))
	err := h.svc.DeductStockOptimistic(c, int64(id), int64(count))
	if err != nil {
		response.Error(c, 40000, "DeductStock Failed")
		return
	}
	response.Success(c, "DeductStock Success")
}
