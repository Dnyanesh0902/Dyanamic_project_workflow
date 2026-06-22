package service

import (
	"errors"
	"project-workflow-backend/dto"
	"project-workflow-backend/model"
	"project-workflow-backend/repository"
	"project-workflow-backend/util"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(req dto.UserCreateRequest) (*dto.UserResponse, error)
	Login(req dto.UserLoginRequest) (*dto.LoginResponse, error)
	List() ([]dto.UserResponse, error)
	Details(req dto.UserByUUIDRequest) (*dto.UserResponse, error)
	Update(req dto.UserUpdateRequest) (*dto.UserResponse, error)
	Delete(req dto.UserByUUIDRequest) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(req dto.UserCreateRequest) (*dto.UserResponse, error) {
	// Check if user already exists
	existing, _ := s.repo.GetByEmail(dto.UserByEmailRequest{Email: req.Email})
	if existing != nil {
		logrus.Warnf("Registration conflict: email %s already in use", req.Email)
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Error hashing password for registration %s: %v", req.Email, err)
		return nil, err
	}
	req.Password = string(hashedPassword)

	userUUID := util.WithoutHypenGenUUID()
	if userUUID == "" {
		userUUID = uuid.New().String()
	}

	user, err := s.repo.Create(req, userUUID)
	if err != nil {
		return nil, err
	}

	logrus.Infof("User created successfully: ID=%d, UUID=%s, Email=%s, Role=%s", user.ID, user.UUID, user.Email, user.Rolename)
	return toUserResponse(user), nil
}

func (s *userService) Login(req dto.UserLoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.GetByEmail(dto.UserByEmailRequest{Email: req.Email})
	if err != nil {
		logrus.Warnf("Login failed: email %s not found", req.Email)
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logrus.Warnf("Login failed: password mismatch for email %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := GenerateToken(user)
	if err != nil {
		logrus.Errorf("Login token generation error for %s: %v", req.Email, err)
		return nil, err
	}

	logrus.Infof("User logged in successfully: ID=%d, Email=%s", user.ID, user.Email)
	return &dto.LoginResponse{
		Token: token,
		User:  *toUserResponse(user),
	}, nil
}

func (s *userService) List() ([]dto.UserResponse, error) {
	users, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	var res []dto.UserResponse
	for _, u := range users {
		res = append(res, *toUserResponse(&u))
	}
	return res, nil
}

func (s *userService) Details(req dto.UserByUUIDRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByUUID(req)
	if err != nil {
		logrus.Warnf("User details not found: UUID %s", req.UUID)
		return nil, errors.New("user not found")
	}
	return toUserResponse(user), nil
}

func (s *userService) Update(req dto.UserUpdateRequest) (*dto.UserResponse, error) {
	_, err := s.repo.GetByUUID(dto.UserByUUIDRequest{UUID: req.UUID})
	if err != nil {
		logrus.Warnf("Update failed: user UUID %s not found", req.UUID)
		return nil, errors.New("user not found")
	}

	if req.Email != "" {
		existing, _ := s.repo.GetByEmail(dto.UserByEmailRequest{Email: req.Email})
		if existing != nil && existing.UUID != req.UUID {
			logrus.Warnf("Update conflict: email %s already in use by another user", req.Email)
			return nil, errors.New("email already in use")
		}
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("Error hashing password during update for user UUID %s: %v", req.UUID, err)
			return nil, err
		}
		req.Password = string(hashedPassword)
	}

	user, err := s.repo.Update(req)
	if err != nil {
		return nil, err
	}

	logrus.Infof("User updated successfully: UUID=%s, Email=%s", user.UUID, user.Email)
	return toUserResponse(user), nil
}

func (s *userService) Delete(req dto.UserByUUIDRequest) error {
	_, err := s.repo.GetByUUID(req)
	if err != nil {
		logrus.Warnf("Delete failed: user UUID %s not found", req.UUID)
		return errors.New("user not found")
	}

	if err := s.repo.Delete(req); err != nil {
		return err
	}

	logrus.Infof("User deleted successfully: UUID=%s", req.UUID)
	return nil
}

func toUserResponse(user *model.User) *dto.UserResponse {
	var deletedAt *time.Time
	if user.DeletedAt.Valid {
		deletedAt = &user.DeletedAt.Time
	}

	return &dto.UserResponse{
		ID:        int(user.ID),
		UUID:      user.UUID,
		Name:      user.Name,
		Email:     user.Email,
		Rolename:  user.Rolename,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
