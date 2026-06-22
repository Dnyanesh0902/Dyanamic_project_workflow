package repository

import (
	"project-workflow-backend/dto"
	"project-workflow-backend/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(req dto.UserCreateRequest, uuid string) (*model.User, error)
	GetByEmail(req dto.UserByEmailRequest) (*model.User, error)
	GetByUUID(req dto.UserByUUIDRequest) (*model.User, error)
	List() ([]model.User, error)
	Update(req dto.UserUpdateRequest) (*model.User, error)
	Delete(req dto.UserByUUIDRequest) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(req dto.UserCreateRequest, uuid string) (*model.User, error) {
	user := &model.User{
		UUID:     uuid,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Rolename: req.Rolename,
	}
	if err := r.db.Create(user).Error; err != nil {
		logrus.Errorf("DB Error creating user with email %s: %v", req.Email, err)
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(req dto.UserByEmailRequest) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("DB Error getting user by email %s: %v", req.Email, err)
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUUID(req dto.UserByUUIDRequest) (*model.User, error) {
	var user model.User
	if err := r.db.Where("uuid = ?", req.UUID).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("DB Error getting user by UUID %s: %v", req.UUID, err)
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List() ([]model.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		logrus.Errorf("DB Error listing users: %v", err)
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(req dto.UserUpdateRequest) (*model.User, error) {
	var user model.User
	if err := r.db.Where("uuid = ?", req.UUID).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("DB Error finding user UUID %s for update: %v", req.UUID, err)
		}
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		user.Password = req.Password
	}
	if req.Rolename != "" {
		user.Rolename = req.Rolename
	}

	if err := r.db.Save(&user).Error; err != nil {
		logrus.Errorf("DB Error updating user UUID %s: %v", req.UUID, err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Delete(req dto.UserByUUIDRequest) error {
	if err := r.db.Where("uuid = ?", req.UUID).Delete(&model.User{}).Error; err != nil {
		logrus.Errorf("DB Error deleting user UUID %s: %v", req.UUID, err)
		return err
	}
	return nil
}
