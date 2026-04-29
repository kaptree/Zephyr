package handlers

import (
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DepartmentHandler struct {
	deptRepo *repository.DepartmentRepository
}

func NewDepartmentHandler(deptRepo *repository.DepartmentRepository) *DepartmentHandler {
	return &DepartmentHandler{deptRepo: deptRepo}
}

func (h *DepartmentHandler) GetTree(c *gin.Context) {
	flat := c.Query("flat")

	if flat == "true" {
		depts, err := h.deptRepo.FindAll()
		if err != nil {
			utils.InternalError(c, "查询部门列表失败")
			return
		}
		utils.Success(c, depts)
		return
	}

	tree, err := h.deptRepo.BuildTree()
	if err != nil {
		utils.InternalError(c, "查询部门树失败")
		return
	}
	utils.Success(c, tree)
}

func (h *DepartmentHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")
	dept, err := h.deptRepo.FindByID(id)
	if err != nil || dept == nil {
		utils.NotFound(c, "部门不存在")
		return
	}

	count, _ := h.deptRepo.CountMembers(id)

	utils.Success(c, gin.H{
		"id":           dept.ID,
		"name":         dept.Name,
		"parent_id":    dept.ParentID,
		"level":        dept.Level,
		"leader":       dept.Leader,
		"member_count": count,
		"created_at":   dept.CreatedAt,
		"updated_at":   dept.UpdatedAt,
	})
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		ParentID *string `json:"parent_id"`
		Level    int     `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	model := models.Department{
		Name:  req.Name,
		Level: req.Level,
	}
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			model.ParentID = &pid
		}
	}
	if model.Level == 0 {
		model.Level = 1
	}

	if err := h.deptRepo.Create(&model); err != nil {
		utils.InternalError(c, "创建部门失败")
		return
	}

	utils.Created(c, model)
}
