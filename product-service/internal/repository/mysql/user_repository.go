package mysql

import (
	"context"
	"gorm.io/gorm"
	"product-service/internal/domain"
	"product-service/internal/errno"
	"product-service/internal/repository/mysql/model"
	"product-service/pkg/utils"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
func (r *UserRepository) Register(ctx context.Context, u *model.UserModel) (*domain.User, error) {
	var user model.UserModel
	err := r.db.WithContext(ctx).Where("username = ?", u.Username).First(&user).Error
	if err == nil {
		return nil, errno.UsernameAlreadyExist
	}
	err = r.db.WithContext(ctx).Create(u).Error
	if err != nil {
		return nil, err
	}
	return toUserDomain(&user), err
}

func (r *UserRepository) Login(ctx context.Context, username string, password string) (*domain.User, error) {
	var user model.UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errno.UserErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !user.CheckPassword(password) {
		return nil, errno.UserErrPasswordIncorrect
	}
	return toUserDomain(&user), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user model.UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errno.UserErrNotFound
	}
	return toUserDomain(&user), err
}

func (r *UserRepository) GetByUserId(ctx context.Context, userId int64) (*domain.User, error) {
	var user model.UserModel
	err := r.db.WithContext(ctx).Where("id = ?", userId).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errno.UserErrNotFound
	}
	return toUserDomain(&user), err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, username string, oldPassword, newPassword string) error {
	var user model.UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return errno.UserErrNotFound
	}
	if err != nil {
		return err
	}
	if !utils.VerifyPassword(oldPassword, user.Password) {
		return errno.UserOldPasswordIncorrect
	}
	newPassword = utils.HashPassword(newPassword)
	return r.db.WithContext(ctx).Model(&user).Update("password", newPassword).Error
}
