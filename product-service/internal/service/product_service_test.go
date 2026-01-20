// product-service/internal/service/order_service_idempotency_test.go
package service_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"product-service/internal/config"
	"product-service/internal/errno"
	mysqlrepo "product-service/internal/repository/mysql"
	"product-service/internal/repository/mysql/model"
	"product-service/internal/service"
	"product-service/pkg/db"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = godotenv.Load() // 读取根目录 .env
	_ = godotenv.Load("../../.env")

	cfg := config.DBConfig{
		DBHost: os.Getenv("MYSQL_TEST_HOST"),
		DBPort: os.Getenv("MYSQL_TEST_PORT"),
		DBUser: os.Getenv("MYSQL_TEST_USER"),
		DBPass: os.Getenv("MYSQL_TEST_PASS"),
		DBName: os.Getenv("MYSQL_TEST_DB"),
	}

	require.NotEmpty(t, cfg.DBHost)
	require.NotEmpty(t, cfg.DBPort)
	require.NotEmpty(t, cfg.DBUser)
	require.NotEmpty(t, cfg.DBName)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	gdb := db.InitMySQL(dsn)

	require.NoError(t, gdb.AutoMigrate(&model.OrderModel{}))

	// 清表，保证用例隔离
	require.NoError(t, gdb.Exec("TRUNCATE TABLE orders").Error)

	return gdb
}

func TestOrderService_Create_Idempotency_Concurrent(t *testing.T) {
	gdb := newTestDB(t)

	repo := mysqlrepo.NewOrderRepository(gdb)
	productRepo := mysqlrepo.NewProductRepository(gdb)
	productService := service.NewProductService(productRepo)
	svc := service.NewOrderService(repo, productService) // 同上，按实际改

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const (
		userID    int64 = 1001
		productID int64 = 2001
		count           = 1
		idemKey         = "idem-concurrent-001"
		n               = 10
	)

	var (
		wg       sync.WaitGroup
		orderNos = make([]string, n)
		errs     = make([]error, n)
	)

	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			o, err := svc.Create(ctx, userID, productID, count, idemKey)
			errs[i] = err
			if err == nil && o != nil {
				orderNos[i] = o.OrderNo
			}
		}()
	}
	wg.Wait()

	// 1) 不应有错误（幂等应返回同一订单）
	for i := 0; i < n; i++ {
		require.NoError(t, errs[i], "goroutine %d failed", i)
		require.NotEmpty(t, orderNos[i], "goroutine %d empty orderNo", i)
	}

	// 2) 所有 orderNo 必须一致
	for i := 1; i < n; i++ {
		require.Equal(t, orderNos[0], orderNos[i], "orderNo mismatch at %d", i)
	}

	// 3) DB 里只有 1 条 user_id + idem_key
	var cnt int64
	require.NoError(t, gdb.Model(&model.OrderModel{}).
		Where("user_id = ? AND idem_key = ?", userID, idemKey).
		Count(&cnt).Error)
	require.Equal(t, int64(1), cnt)
}

func TestOrderService_Create_Idempotency_ParamMismatch(t *testing.T) {
	gdb := newTestDB(t)

	repo := mysqlrepo.NewOrderRepository(gdb)
	productRepo := mysqlrepo.NewProductRepository(gdb)
	productService := service.NewProductService(productRepo)
	svc := service.NewOrderService(repo, productService)

	ctx := context.Background()

	const (
		userID  int64 = 1002
		idemKey       = "idem-mismatch-001"
	)

	// 第一次创建
	_, err := svc.Create(ctx, userID, 2001, 1, idemKey)
	require.NoError(t, err)

	// 第二次同 idemKey 但参数不一致，应返回你新增的业务错误
	_, err = svc.Create(ctx, userID, 2002, 1, idemKey)
	require.Error(t, err)
	require.ErrorIs(t, err, errno.OrderErrOrderAlreadyExist)
}
