package usecase

import (
	"fido2/internal/entity"
)

type UserUseCase interface {
	CreateUser(user *entity.User) error
	GetUserByID(id string) (*entity.User, error)
	GetUserByUsername(username string) (*entity.User, error)
	GetUserByChallenge(challenge string) (*entity.User, error)
	GetUsers() ([]*entity.User, error)
	UpdateUser(user *entity.User, updateData interface{}) error
	DeleteUser(id string) error
}