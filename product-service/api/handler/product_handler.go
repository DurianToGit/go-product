package handler

import (
	"errors"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/response"
	"product-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductQuery struct {
	Keyword  string `form:"keyword" json:"keyword"`
	MinPrice *int64 `form:"min_price" json:"min_price"`
	MaxPrice *int64 `form:"max_price" json:"max_price"`
	Status   int    `form:"status" json:"status"`
	Page     int    `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}

func (p *ProductQuery) ToDto() *dto.ProductQuery {
	return &dto.ProductQuery{
		Keyword:  p.Keyword,
		MinPrice: p.MinPrice,
		MaxPrice: p.MaxPrice,
		Status:   p.Status,
		Page:     p.Page,
		PageSize: p.PageSize,
	}
}

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc}
}

func (h *ProductHandler) List(c *gin.Context) {
	var req ProductQuery
	if !BindAndValidateByQuery(c, &req) {
		return
	}
	q := req.ToDto()

	data, total, err := h.svc.GetProducts(c, q)
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	result := response.ResultData{
		List:     data,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	response.Success(c, result)
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
