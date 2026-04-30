package repository

import (
	"labelpro-server/internal/models"
	"sort"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) FindAll(scope string) ([]models.Tag, error) {
	var tags []models.Tag
	query := r.db.Order("sort_order ASC, name ASC")
	if scope != "" && scope != "all" {
		query = query.Where("scope = ?", scope)
	}
	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}

	var result []models.Tag
	for i := range tags {
		var count int64
		r.db.Table("note_tags").Where("tag_id = ?", tags[i].ID).Count(&count)
		tags[i].UsageCount = count
		if count > 0 {
			result = append(result, tags[i])
		}
	}

	for i := range tags {
		if tags[i].UsageCount == 0 {
			r.db.Delete(&models.Tag{}, "id = ?", tags[i].ID)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].UsageCount != result[j].UsageCount {
			return result[i].UsageCount > result[j].UsageCount
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (r *TagRepository) FindByID(id string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepository) Update(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

func (r *TagRepository) Delete(id string) error {
	return r.db.Delete(&models.Tag{}, "id = ?", id).Error
}

func (r *TagRepository) IsInUse(id string) (bool, error) {
	var count int64
	err := r.db.Table("note_tags").Where("tag_id = ?", id).Count(&count).Error
	return count > 0, err
}

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) FindAll(templateType string) ([]models.Template, error) {
	var templates []models.Template
	query := r.db.Order("created_at DESC")
	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}
	if err := query.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepository) FindByID(id string) (*models.Template, error) {
	var tmpl models.Template
	err := r.db.First(&tmpl, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *TemplateRepository) Create(tmpl *models.Template) error {
	return r.db.Create(tmpl).Error
}

func (r *TemplateRepository) Update(tmpl *models.Template) error {
	return r.db.Save(tmpl).Error
}

func (r *TemplateRepository) Delete(id string) error {
	return r.db.Delete(&models.Template{}, "id = ?", id).Error
}

type WorkGroupRepository struct {
	db *gorm.DB
}

func NewWorkGroupRepository(db *gorm.DB) *WorkGroupRepository {
	return &WorkGroupRepository{db: db}
}

func (r *WorkGroupRepository) Create(group *models.WorkGroup) error {
	return r.db.Create(group).Error
}

func (r *WorkGroupRepository) FindByID(id string) (*models.WorkGroup, error) {
	var group models.WorkGroup
	err := r.db.Preload("Members.User").Preload("Initiator").First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *WorkGroupRepository) UpdateMember(groupID, userID, role, subGroup string) error {
	updates := map[string]interface{}{}
	if role != "" {
		updates["role"] = role
	}
	if subGroup != "" {
		updates["sub_group_name"] = subGroup
	}
	return r.db.Model(&models.WorkGroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(updates).Error
}

type CollaborationRoomRepository struct {
	db *gorm.DB
}

func NewCollaborationRoomRepository(db *gorm.DB) *CollaborationRoomRepository {
	return &CollaborationRoomRepository{db: db}
}

func (r *CollaborationRoomRepository) Create(room *models.CollaborationRoom) error {
	return r.db.Create(room).Error
}

func (r *CollaborationRoomRepository) FindByNoteID(noteID string) (*models.CollaborationRoom, error) {
	var room models.CollaborationRoom
	err := r.db.First(&room, "note_id = ?", noteID).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *CollaborationRoomRepository) UpdateCanvas(noteID, canvasData string, version int) error {
	return r.db.Model(&models.CollaborationRoom{}).
		Where("note_id = ?", noteID).
		Updates(map[string]interface{}{
			"canvas_data": canvasData,
			"version":     version,
		}).Error
}

type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) List(filter LedgerFilter) ([]models.LedgerEntry, int64, error) {
	var entries []models.LedgerEntry
	var total int64

	query := r.db.Model(&models.LedgerEntry{}).Preload("User")

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.DeptID != "" {
		query = query.Where("user_id IN (SELECT id FROM users WHERE department_id = ?)", filter.DeptID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.DateFrom != "" {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("created_at <= ?", filter.DateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

type LedgerFilter struct {
	UserID   string
	DeptID   string
	DateFrom string
	DateTo   string
	Action   string
	Page     int
	PageSize int
}

func (r *LedgerRepository) CountByAction() (map[string]int64, error) {
	type row struct {
		Action string
		Count  int64
	}
	var rows []row
	err := r.db.Model(&models.LedgerEntry{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64)
	for _, r := range rows {
		m[r.Action] = r.Count
	}
	return m, nil
}
