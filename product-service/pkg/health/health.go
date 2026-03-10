package health

import (
	"context"
	"net/http"
	"product-service/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Checker struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (c *Checker) Check(ctx context.Context) error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			return err
		}
	}

	if c.Redis != nil {
		if err := c.Redis.Ping(ctx).Err(); err != nil {
			return err
		}
	}

	return nil
}

func Handler(checker *Checker) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		err := checker.Check(ctx)

		if err != nil {
			response.Error(c, http.StatusServiceUnavailable, err.Error())
			return
		}

		response.Success(c, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	}
}
