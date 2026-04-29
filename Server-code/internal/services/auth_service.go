package services

import (
	"context"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *UserProfile `json:"user"`
}

type UserProfile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Rank        string   `json:"rank"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Avatar      string   `json:"avatar"`
	DeptID      string   `json:"dept_id"`
	DeptName    string   `json:"dept_name"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}

func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	if !user.IsActive {
		return nil, apperrors.ErrUserInactive
	}

	if s.cfg.Features.DemoMode {
	} else {
		if !utils.CheckPassword(password, user.PasswordHash) {
			return nil, apperrors.ErrInvalidPassword
		}
	}

	deptID := ""
	deptName := ""
	if user.Department != nil {
		deptID = user.Department.ID.String()
		deptName = user.Department.Name
	}

	tokens, err := utils.GenerateTokenPair(user.ID.String(), user.Username, user.Role, deptID)
	if err != nil {
		return nil, err
	}

	s.userRepo.UpdateLastLogin(user.ID.String())

	profile := &UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		Name:        user.Name,
		Role:        user.Role,
		Rank:        user.Rank,
		Phone:       user.Phone,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
		DeptID:      deptID,
		DeptName:    deptName,
		Permissions: s.getPermissions(user.Role),
		IsActive:    user.IsActive,
	}

	return &LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         profile,
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*utils.TokenPair, error) {
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		return nil, apperrors.ErrTokenInvalid
	}

	if database.RDB != nil {
		ctx := context.Background()
		exists, _ := database.RDB.SIsMember(ctx, "token_blacklist", claims.ID).Result()
		if exists {
			return nil, apperrors.ErrTokenRevoked
		}
	}

	return utils.GenerateTokenPair(claims.UserID, claims.Username, claims.Role, claims.DepartmentID)
}

func (s *AuthService) Logout(accessToken string) error {
	claims, err := utils.ParseToken(accessToken)
	if err != nil {
		return nil
	}

	if database.RDB != nil {
		ctx := context.Background()
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			database.RDB.SAdd(ctx, "token_blacklist", claims.ID)
			database.RDB.Expire(ctx, "token_blacklist", ttl+time.Hour)
		}
	}

	return nil
}

func (s *AuthService) IsTokenBlacklisted(tokenID string) bool {
	if database.RDB == nil {
		return false
	}
	ctx := context.Background()
	exists, _ := database.RDB.SIsMember(ctx, "token_blacklist", tokenID).Result()
	return exists
}

func (s *AuthService) GetUserProfile(userID string) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	deptID := ""
	deptName := ""
	if user.Department != nil {
		deptID = user.Department.ID.String()
		deptName = user.Department.Name
	}

	return &UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		Name:        user.Name,
		Role:        user.Role,
		Rank:        user.Rank,
		Phone:       user.Phone,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
		DeptID:      deptID,
		DeptName:    deptName,
		Permissions: s.getPermissions(user.Role),
		IsActive:    user.IsActive,
	}, nil
}

func (s *AuthService) getPermissions(role string) []string {
	perms := []string{}

	switch role {
	case "super_admin":
		perms = []string{
			"create_note_self", "create_note_assigned", "edit_others_note",
			"delete_note", "remind", "view_all_archive", "view_dept_archive",
			"view_group_archive", "manage_departments", "manage_users",
			"manage_tags", "manage_templates", "access_screen", "send_command",
		}
	case "dept_admin":
		perms = []string{
			"create_note_self", "create_note_assigned", "edit_others_note",
			"delete_note", "remind", "view_dept_archive", "view_group_archive",
			"manage_users", "manage_tags", "access_screen", "send_command",
		}
	case "group_leader":
		perms = []string{
			"create_note_self", "create_note_assigned", "edit_others_note",
			"remind", "view_group_archive", "view_dept_archive",
			"access_screen", "send_command",
		}
	case "member":
		perms = []string{
			"create_note_self", "view_dept_archive", "view_group_archive",
		}
	case "screen_role":
		perms = []string{
			"access_screen",
		}
	}

	return perms
}

func (s *AuthService) CreateUser(req CreateUserRequest) (*models.User, error) {
	existing, _ := s.userRepo.FindByUsername(req.Username)
	if existing != nil {
		return nil, apperrors.ErrDuplicateUsername
	}

	pwdErr := utils.ValidatePasswordComplexity(
		req.Password,
		s.cfg.Security.PasswordMinLength,
		s.cfg.Security.PasswordRequireUpper,
		s.cfg.Security.PasswordRequireLower,
		s.cfg.Security.PasswordRequireDigit,
		s.cfg.Security.PasswordRequireSpecial,
	)
	if pwdErr != nil {
		return nil, pwdErr
	}

	hash, err := utils.HashPasswordWithCost(req.Password, s.cfg.Security.BcryptCost)
	if err != nil {
		return nil, err
	}

	var deptID *uuid.UUID
	if req.DepartmentID != "" {
		id, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			return nil, err
		}
		deptID = &id
	}

	user := &models.User{
		Username:     req.Username,
		Name:         req.Name,
		DepartmentID: deptID,
		Role:         req.Role,
		Rank:         req.Rank,
		Phone:        req.Phone,
		Email:        req.Email,
		AvatarURL:    req.Avatar,
		PasswordHash: hash,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

type CreateUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Role         string `json:"role"`
	Rank         string `json:"rank"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	DepartmentID string `json:"dept_id"`
}
