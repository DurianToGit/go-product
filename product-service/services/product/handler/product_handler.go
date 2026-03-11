package handler

import (
	"errors"
	"fmt"
	"product-service/internal/errno"
	"product-service/pkg/response"
	"product-service/pkg/utils"
	"product-service/services/product/cache"
	"product-service/services/product/domain"
	"product-service/services/product/dto"
	"product-service/services/product/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductReq struct {
	Keyword         string `form:"keyword" json:"keyword"`
	CreatorUsername string `form:"creator_username" json:"creator_username"`
	MinPrice        *int64 `form:"min_price" json:"min_price"`
	MaxPrice        *int64 `form:"max_price" json:"max_price"`
	Status          *int   `form:"status" json:"status"`
	Page            int    `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize        int    `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}

type ProductSecKillReq struct {
	Count   int64  `json:"count" binding:"required,min=1"`
	IdemKey string `json:"idem_key" binding:"required"`
}

func (p *ProductReq) ToDto() *dto.ProductQuery {
	return &dto.ProductQuery{
		Keyword:         p.Keyword,
		CreatorUsername: p.CreatorUsername,
		MinPrice:        p.MinPrice,
		MaxPrice:        p.MaxPrice,
		Status:          p.Status,
		Page:            p.Page,
		PageSize:        p.PageSize,
	}
}

type ProductHandler struct {
	svc   *service.ProductService
	cache *cache.ProductCache
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc, cache.NewProductCache()}
}

func (h *ProductHandler) List(c *gin.Context) {
	var req ProductReq
	if !utils.BindAndValidateByQuery(c, &req) {
		return
	}
	q := req.ToDto()

	data, total, err := h.svc.GetProducts(c, q)
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	result := response.ResultData{
		List: data,
		Mata: &response.Meta{
			Total:    total,
			Page:     q.Page,
			PageSize: q.PageSize,
		},
	}
	response.Success(c, result)
}

func (h *ProductHandler) Search(c *gin.Context) {
	var req ProductReq
	if !utils.BindAndValidateByQuery(c, &req) {
		return
	}
	q := req.ToDto()
	data, total, err := h.svc.ProductSearch(c, q)
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, response.ResultData{
		List: data,
		Mata: &response.Meta{
			Total:    total,
			Page:     q.Page,
			PageSize: q.PageSize,
		},
	})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req domain.Product
	if !utils.BindAndValidateByJSON(c, &req) {
		return
	}

	if err := h.svc.CreateProduct(c, &req); err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, req)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	data, ok := h.cache.Get(id)
	if !ok {
		data, err = h.svc.GetProduct(c, id)

		if errors.Is(err, errno.ErrDataNotFound) {
			response.ErrorWithErrno(c, errno.ErrDataNotFound)
			return
		}

		if err != nil {
			response.Error(c, 40001, err.Error())
			return
		}
		h.cache.Set(data)
	}
	response.Success(c, data)
}

// 秒杀
func (h *ProductHandler) DuctStockSeckill(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	var productSecKillReq ProductSecKillReq
	if !utils.BindAndValidateByJSON(c, &productSecKillReq) {
		return
	}
	userId := utils.GetUserID(c)
	val, err := h.svc.DeductStockSeckill(c, id, productSecKillReq.Count, userId, productSecKillReq.IdemKey)
	if err != nil {
		if errors.Is(err, errno.ProductErrSeckillStockNotInit) {
			response.ErrorWithErrno(c, errno.ProductErrSeckillStockNotInit)
			return
		}
		if errors.Is(err, errno.ProductErrSeckillStockNotEnough) {
			response.ErrorWithErrno(c, errno.ProductErrSeckillStockNotEnough)
			return
		}
		response.Error(c, 40000, err.Error())
		return
	}
	response.Success(c, gin.H{
		"val": val,
	})
}

// 预热商品库存
func (h *ProductHandler) PrewarmProductStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	err = h.svc.PrewarmProductStock(c, id)
	if err != nil {
		response.Error(c, 40000, fmt.Sprintf("PrewarmProductStock Failed:%v", err))
	}
	response.Success(c, nil)
}
