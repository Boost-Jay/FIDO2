package impl

import (
	"fido2/internal/entity"
	"fido2/internal/repository"
	"fido2/internal/usecase"
	"sync"
)

type userUseCaseImpl struct {
	userRepo repository.UserRepository
}

var _ usecase.UserUseCase = (*userUseCaseImpl)(nil)

var (
	userUseCase usecase.UserUseCase
	userOnce    sync.Once
)

func GetUserUseCase() usecase.UserUseCase {
	userOnce.Do(func() {
		userRepo := repository.NewUserRepository()
		userUseCase = NewUserUseCase(userRepo)
	})
	return userUseCase
}

// 建構函式(Constructor)

func NewUserUseCase(userRepo repository.UserRepository) usecase.UserUseCase {
	return &userUseCaseImpl{
		userRepo: userRepo,
	}
}

func (uc *userUseCaseImpl) CreateUser(user *entity.User) error {
	return uc.userRepo.CreateUser(user)
}

func (uc *userUseCaseImpl) GetUserByID(id string) (*entity.User, error) {
	return uc.userRepo.GetUserByID(id)
}

func (uc *userUseCaseImpl) GetUserByUsername(username string) (*entity.User, error) {
	return uc.userRepo.GetUserByUsername(username)
}

func (uc *userUseCaseImpl) GetUserByChallenge(challenge string) (*entity.User, error) {
	return uc.userRepo.GetUserByChallenge(challenge)
}

func (uc *userUseCaseImpl) GetUsers() ([]*entity.User, error) {
	return uc.userRepo.GetUsers()
}

func (uc *userUseCaseImpl) UpdateUser(user *entity.User, updateData interface{}) error {
	return uc.userRepo.UpdateUser(user, updateData)
}

func (uc *userUseCaseImpl) DeleteUser(id string) error {
	return uc.userRepo.DeleteUser(id)
}