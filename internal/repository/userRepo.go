package repository

import (
	"fido2/internal/entity"
	"fido2/internal/platform/db"
)

// UserRepository 定義了用戶資料操作的介面
type UserRepository interface {
	CreateUser(user *entity.User) error
	GetUserByID(id string) (*entity.User, error)
	GetUserByUsername(username string) (*entity.User, error)
	GetUserByChallenge(challenge string) (*entity.User, error)
	GetUsers() ([]*entity.User, error)
	UpdateUser(user *entity.User, updateData interface{}) error
	DeleteUser(id string) error
}

// userRepositoryImpl 實作 UserRepository 介面
type userRepositoryImpl struct{}

// NewUserRepository 建立 UserRepository 的新實例
func NewUserRepository() UserRepository {
	return &userRepositoryImpl{}
}

// CreateUser 在資料庫中建立新用戶
func (r *userRepositoryImpl) CreateUser(user *entity.User) error {
	return db.GetDB().Create(user).Error
}

// GetUserByID 透過 ID 取得用戶
func (r *userRepositoryImpl) GetUserByID(id string) (*entity.User, error) {
	var user entity.User
	if err := db.GetDB().First(&user, id).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 透過使用者名稱取得用戶
func (r *userRepositoryImpl) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	if err := db.GetDB().Where("user_name = ?", username).First(&user).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByChallenge 透過 Challenge 取得用戶
func (r *userRepositoryImpl) GetUserByChallenge(challenge string) (*entity.User, error) {
	var user entity.User
	if err := db.GetDB().Where("challenge = ?", challenge).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers 取得所有用戶
func (r *userRepositoryImpl) GetUsers() ([]*entity.User, error) {
	var users []*entity.User
	if err := db.GetDB().Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser 更新用戶資料
func (r *userRepositoryImpl) UpdateUser(user *entity.User, updateData interface{}) error {
	return db.GetDB().Model(user).Updates(updateData).Error
}

// DeleteUser 刪除用戶
func (r *userRepositoryImpl) DeleteUser(id string) error {
	return db.GetDB().Delete(&entity.User{}, id).Error
}