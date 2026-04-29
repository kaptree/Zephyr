package handlers

import (
	"strconv"

	"labelpro-server/internal/repository"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repository.UserFilter{
		DeptID:   c.Query("dept_id"),
		Role:     c.Query("role"),
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	}

	users, total, err := h.userService.ListUsers(filter)
	if err != nil {
		utils.InternalError(c, "查询用户列表失败")
		return
	}

	utils.Paginated(c, users, total, page, pageSize)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetUser(id)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	utils.Success(c, user)
}

func (h *UserHandler) GetVisibleUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	users, err := h.userService.GetVisibleUsers(userID)
	if err != nil {
		utils.InternalError(c, "查询可见用户失败")
		return
	}
	utils.Success(c, users)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	user, err := h.userService.UpdateUser(id, req)
	if err != nil {
		utils.InternalError(c, "更新用户失败")
		return
	}

	utils.Success(c, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.userService.DeleteUser(id); err != nil {
		utils.InternalError(c, "删除用户失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}
