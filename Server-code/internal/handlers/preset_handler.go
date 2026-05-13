package handlers

import (
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PresetGroupHandler struct {
	presetRepo *repository.PresetGroupRepository
}

func NewPresetGroupHandler(presetRepo *repository.PresetGroupRepository) *PresetGroupHandler {
	return &PresetGroupHandler{presetRepo: presetRepo}
}

type CreatePresetReq struct {
	Name         string           `json:"name" binding:"required"`
	Description  string           `json:"description"`
	TemplateType string           `json:"template_type"`
	Members      []PresetMemberReq `json:"members"`
}

type PresetMemberReq struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	SubGroupName string `json:"sub_group_name"`
}

func (h *PresetGroupHandler) List(c *gin.Context) {
	templateType := c.Query("template_type")
	presets, err := h.presetRepo.FindAll(templateType)
	if err != nil {
		utils.InternalError(c, "查询预设组失败")
		return
	}
	utils.Success(c, presets)
}

func (h *PresetGroupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	preset, err := h.presetRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "预设组不存在")
		return
	}
	utils.Success(c, preset)
}

func (h *PresetGroupHandler) Create(c *gin.Context) {
	var req CreatePresetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写完整的预设组信息")
		return
	}

	userID := middleware.GetUserID(c)
	creatorUID, _ := uuid.Parse(userID)

	templateType := req.TemplateType
	if templateType == "" {
		templateType = "default"
	}

	preset := &models.PresetGroup{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		TemplateType: templateType,
		CreatorID:    creatorUID,
		Members:      make([]models.PresetGroupMember, 0, len(req.Members)),
	}

	for _, m := range req.Members {
		memberUID, err := uuid.Parse(m.UserID)
		if err != nil {
			continue
		}
		role := m.Role
		if role == "" {
			role = "member"
		}
		preset.Members = append(preset.Members, models.PresetGroupMember{
			UserID:       memberUID,
			Role:         role,
			SubGroupName: m.SubGroupName,
		})
	}

	if err := h.presetRepo.Create(preset); err != nil {
		utils.InternalError(c, "创建预设组失败")
		return
	}
	utils.Created(c, preset)
}

func (h *PresetGroupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.presetRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "预设组不存在")
		return
	}

	var req CreatePresetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	if req.TemplateType != "" {
		existing.TemplateType = req.TemplateType
	}

	existing.Members = make([]models.PresetGroupMember, 0, len(req.Members))
	for _, m := range req.Members {
		memberUID, err := uuid.Parse(m.UserID)
		if err != nil {
			continue
		}
		role := m.Role
		if role == "" {
			role = "member"
		}
		existing.Members = append(existing.Members, models.PresetGroupMember{
			PresetID:     existing.ID,
			UserID:       memberUID,
			Role:         role,
			SubGroupName: m.SubGroupName,
		})
	}

	if err := h.presetRepo.Update(id, existing); err != nil {
		utils.InternalError(c, "更新预设组失败")
		return
	}
	updated, _ := h.presetRepo.FindByID(id)
	utils.Success(c, updated)
}

func (h *PresetGroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.presetRepo.Delete(id); err != nil {
		utils.InternalError(c, "删除预设组失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

func (h *PresetGroupHandler) Recommend(c *gin.Context) {
	workType := c.DefaultQuery("template_type", "default")
	presets, err := h.presetRepo.RecommendByWorkType(workType, 5)
	if err != nil {
		utils.InternalError(c, "推荐预设组失败")
		return
	}
	utils.Success(c, presets)
}