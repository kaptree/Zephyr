package handlers

import (
	"net/http"

	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"
	apperrors "labelpro-server/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入用户名和密码")
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case apperrors.ErrUserNotFound:
			utils.Unauthorized(c, "用户名或密码错误")
		case apperrors.ErrUserInactive:
			utils.Forbidden(c, "账号已被禁用")
		case apperrors.ErrInvalidPassword:
			utils.Unauthorized(c, "用户名或密码错误")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.Set("user_id", resp.User.ID)
	c.Set("username", resp.User.Username)
	c.Set("role", resp.User.Role)

	utils.Success(c, resp)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请提供refresh_token")
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, "令牌无效或已过期")
		return
	}

	if claims, parseErr := utils.ParseToken(req.RefreshToken); parseErr == nil {
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
	}

	utils.Success(c, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("access_token")
	if token != "" {
		_ = h.authService.Logout(token)
	}
	utils.Success(c, gin.H{"success": true})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "")
		return
	}

	profile, err := h.authService.GetUserProfile(userID)
	if err != nil {
		utils.InternalError(c, "获取用户信息失败")
		return
	}

	utils.Success(c, profile)
}

type RegisterRequest struct {
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	createReq := services.CreateUserRequest{
		Username:     req.Username,
		Name:         req.Name,
		Password:     req.Password,
		Role:         req.Role,
		Rank:         req.Rank,
		Phone:        req.Phone,
		Email:        req.Email,
		Avatar:       req.Avatar,
		DepartmentID: req.DepartmentID,
	}

	user, err := h.authService.CreateUser(createReq)
	if err != nil {
		if err == apperrors.ErrDuplicateUsername {
			utils.BadRequest(c, "用户名已存在")
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, user)
}

func (h *AuthHandler) NoRoute(c *gin.Context) {
	utils.Error(c, http.StatusNotFound, 404, "接口不存在")
}
