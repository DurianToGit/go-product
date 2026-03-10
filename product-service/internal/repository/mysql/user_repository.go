package mysql

import (
	"context"
	"gorm.io/gorm"
	"product-service/internal/domain"
	"product-service/internal/dto"
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
	u.Password = utils.HashPassword(u.Password)
	err = r.db.WithContext(ctx).Create(u).Error
	if err != nil {
		return nil, err
	}
	return toUserDomain(u), err
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
	if !utils.VerifyPassword(password, user.Password) {
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

func (r *UserRepository) List(ctx context.Context, query *dto.UserQuery) ([]*domain.User, int64, error) {
	var (
		list  []*model.UserModel
		total int64
	)
	page := query.Page
	pageSize := query.PageSize

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx).Model(&model.UserModel{})
	if query.Keyword != "" {
		db = db.Where("username like ?", "%"+query.Keyword+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	db.Count(&total)
	err := db.Order("id DESC").Limit(pageSize).Offset(offset).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]*domain.User, len(list))
	for i, p := range list {
		result[i] = toUserDomain(p)
	}

	return result, total, err
}

func (r *UserRepository) Update(ctx context.Context, userId int64, req *dto.UserUpdate) error {
	user := map[string]any{}
	if req.Username != nil {
		user["username"] = *req.Username
	}
	if req.Password != nil {
		user["password"] = utils.HashPassword(*req.Password)
	}
	if len(user) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", userId).Updates(user)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errno.UserErrNotFound
	}
	return nil
}

func toUserDomain(c *model.UserModel) *domain.User {
	return &domain.User{
		ID:        c.ID,
		Username:  c.Username,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
