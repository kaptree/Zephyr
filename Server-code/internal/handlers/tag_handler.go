package handlers

import (
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	tagRepo *repository.TagRepository
}

func NewTagHandler(tagRepo *repository.TagRepository) *TagHandler {
	return &TagHandler{tagRepo: tagRepo}
}

func (h *TagHandler) List(c *gin.Context) {
	scope := c.DefaultQuery("scope", "all")
	tags, err := h.tagRepo.FindAll(scope)
	if err != nil {
		utils.InternalError(c, "查询标签失败")
		return
	}
	utils.Success(c, tags)
}

func (h *TagHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		SubTag   string `json:"sub_tag"`
		Color    string `json:"color"`
		Category string `json:"category"`
		Scope    string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	tag := &models.Tag{
		Name:     req.Name,
		SubTag:   req.SubTag,
		Color:    req.Color,
		Category: req.Category,
		Scope:    req.Scope,
	}
	if tag.Scope == "" {
		tag.Scope = "personal"
	}
	if tag.Color == "" {
		tag.Color = "#3B82F6"
	}

	if err := h.tagRepo.Create(tag); err != nil {
		utils.InternalError(c, "创建标签失败")
		return
	}
	utils.Created(c, tag)
}

func (h *TagHandler) Update(c *gin.Context) {
	id := c.Param("id")
	tag, err := h.tagRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "标签不存在")
		return
	}

	var req struct {
		Name     *string `json:"name"`
		SubTag   *string `json:"sub_tag"`
		Color    *string `json:"color"`
		Category *string `json:"category"`
		Scope    *string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.SubTag != nil {
		tag.SubTag = *req.SubTag
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.Category != nil {
		tag.Category = *req.Category
	}
	if req.Scope != nil {
		tag.Scope = *req.Scope
	}

	if err := h.tagRepo.Update(tag); err != nil {
		utils.InternalError(c, "更新标签失败")
		return
	}
	utils.Success(c, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	inUse, _ := h.tagRepo.IsInUse(id)
	if inUse {
		utils.BadRequest(c, "标签正在使用中，无法删除")
		return
	}
	if err := h.tagRepo.Delete(id); err != nil {
		utils.InternalError(c, "删除标签失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}
