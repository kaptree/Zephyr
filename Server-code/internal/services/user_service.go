package services

import (
	"fmt"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepository
	deptRepo *repository.DepartmentRepository
}

func NewUserService(userRepo *repository.UserRepository, deptRepo *repository.DepartmentRepository) *UserService {
	return &UserService{userRepo: userRepo, deptRepo: deptRepo}
}

func (s *UserService) GetUser(id string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) ListUsers(filter repository.UserFilter) ([]models.User, int64, error) {
	return s.userRepo.List(filter)
}

func (s *UserService) GetVisibleUsers(userID string) ([]models.User, error) {
	return s.userRepo.FindVisibleUsers(userID, "")
}

func (s *UserService) UpdateUser(id string, req UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Rank != "" {
		user.Rank = req.Rank
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.AvatarURL = req.Avatar
	}
	if req.DepartmentID != "" {
		did, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			return nil, fmt.Errorf("invalid department UUID: %w", err)
		}
		user.DepartmentID = &did
		user.Department = nil
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return s.userRepo.FindByID(id)
}

func (s *UserService) DeleteUser(id string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperrors.ErrUserNotFound
	}
	return s.userRepo.Delete(id)
}

type UpdateUserRequest struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Rank         string `json:"rank"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	DepartmentID string `json:"dept_id"`
	IsActive     *bool  `json:"is_active"`
}
