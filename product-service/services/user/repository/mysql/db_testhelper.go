package mysql

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"os"
	"product-service/internal/config"
	"product-service/pkg/db"
	"product-service/services/user/repository/mysql/model"
	"testing"
)

func newTestDB(t *testing.T) *gorm.DB {
	// 加载环境变量
	_ = godotenv.Load("../../../.env")
	// require.NoError(t, err)
	t.Helper()

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

	// 迁移（测试环境允许）
	require.NoError(t, gdb.AutoMigrate(
		&model.UserModel{},
	))

	// 清表（保证用例隔离；注意顺序：有外键时需按依赖顺序清）
	require.NoError(t, gdb.Exec("TRUNCATE TABLE users").Error)

	return gdb
}
