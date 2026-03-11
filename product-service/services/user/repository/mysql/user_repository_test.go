package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"product-service/internal/errno"
	"product-service/services/user/dto"
	"product-service/services/user/repository/mysql/model"
)

func TestUserRepository_List_Pagination_Order(t *testing.T) {
	gdb := newTestDB(t)
	repo := NewUserRepository(gdb) // 你的构造函数名按实际
	ctx := context.Background()

	// seed
	require.NoError(t, gdb.Create(&model.UserModel{Username: "u1", Password: "p1"}).Error)
	require.NoError(t, gdb.Create(&model.UserModel{Username: "u2", Password: "p2"}).Error)
	require.NoError(t, gdb.Create(&model.UserModel{Username: "u3", Password: "p3"}).Error)

	// page1
	list1, total, err := repo.List(ctx, &dto.UserQuery{Page: 1, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, list1, 2)
	require.True(t, list1[0].ID > list1[1].ID) // 依赖你已补 Order("id DESC")

	// page2
	list2, total2, err := repo.List(ctx, &dto.UserQuery{Page: 2, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), total2)
	require.Len(t, list2, 1)
}

func TestUserRepository_List_Filter(t *testing.T) {
	gdb := newTestDB(t)
	repo := NewUserRepository(gdb)
	ctx := context.Background()

	require.NoError(t, gdb.Create(&model.UserModel{Username: "tom_1", Password: "p"}).Error)
	require.NoError(t, gdb.Create(&model.UserModel{Username: "tom_2", Password: "p"}).Error)
	require.NoError(t, gdb.Create(&model.UserModel{Username: "jack_1", Password: "p"}).Error)

	list, total, err := repo.List(ctx, &dto.UserQuery{
		Keyword:  "tom",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
}

func TestUserRepository_Update_Partial_And_NotFound(t *testing.T) {
	gdb := newTestDB(t)
	repo := NewUserRepository(gdb)
	ctx := context.Background()

	u := &model.UserModel{Username: "alice", Password: "oldhash"}
	require.NoError(t, gdb.Create(u).Error)

	// 只更新 username（password 不变）
	newName := "alice_new"
	require.NoError(t, repo.Update(ctx, u.ID, &dto.UserUpdate{Username: &newName}))

	var got model.UserModel
	require.NoError(t, gdb.First(&got, u.ID).Error)
	require.Equal(t, "alice_new", got.Username)
	require.Equal(t, "oldhash", got.Password)

	// 更新不存在
	err := repo.Update(ctx, 999999999, &dto.UserUpdate{Username: &newName})
	require.ErrorIs(t, err, errno.UserErrNotFound)
}
