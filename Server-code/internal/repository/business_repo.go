package repository

import (
	"labelpro-server/internal/models"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) FindAll(scope string, parentID *uuid.UUID, category string) ([]models.Tag, error) {
	var tags []models.Tag
	query := r.db.Order("sort_order ASC, name ASC")
	if scope != "" && scope != "all" {
		query = query.Where("scope = ?", scope)
	}
	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if category == "一级分类" {
		query = query.Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, name ASC")
		})
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

func (r *WorkGroupRepository) FindAll() ([]models.WorkGroup, error) {
	var groups []models.WorkGroup
	err := r.db.Preload("Members.User").Preload("Initiator").Order("created_at DESC").Find(&groups).Error
	return groups, err
}

func (r *WorkGroupRepository) FindByUserID(userID string) ([]models.WorkGroup, error) {
	var groups []models.WorkGroup
	subQuery := r.db.Model(&models.WorkGroupMember{}).Select("group_id").Where("user_id = ?", userID)
	err := r.db.Preload("Members.User").Preload("Initiator").
		Where("initiator_id = ? OR id IN (?)", userID, subQuery).
		Order("created_at DESC").
		Find(&groups).Error
	return groups, err
}

func (r *WorkGroupRepository) AddMember(groupID, userID, role, subGroup string) error {
	member := &models.WorkGroupMember{
		GroupID:      uuid.MustParse(groupID),
		UserID:       uuid.MustParse(userID),
		Role:         role,
		SubGroupName: subGroup,
	}
	return r.db.Create(member).Error
}

func (r *WorkGroupRepository) RemoveMember(groupID, userID string) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.WorkGroupMember{}).Error
}

func (r *WorkGroupRepository) UpdateStatus(groupID, status string) error {
	return r.db.Model(&models.WorkGroup{}).Where("id = ?", groupID).Update("status", status).Error
}

func (r *WorkGroupRepository) Delete(id string) error {
	r.db.Where("group_id = ?", id).Delete(&models.WorkGroupMember{})
	return r.db.Delete(&models.WorkGroup{}, "id = ?", id).Error
}

type WorkTypeStatResult struct {
	UserID     string
	WorkType   string
	GroupCount int64
}

func (r *WorkGroupRepository) GetWorkTypeStatsByUser(userID string) ([]models.WorkTypeStat, error) {
	var results []WorkTypeStatResult
	err := r.db.Table("work_group_members AS m").
		Select("m.user_id::text AS user_id, g.template_type AS work_type, COUNT(DISTINCT g.id) AS group_count").
		Joins("JOIN work_groups AS g ON g.id = m.group_id").
		Where("m.user_id = ?", userID).
		Group("m.user_id, g.template_type").
		Order("group_count DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	stats := make([]models.WorkTypeStat, len(results))
	for i, r := range results {
		stats[i] = models.WorkTypeStat{WorkType: r.WorkType, GroupCount: r.GroupCount}
	}
	return stats, nil
}

func (r *WorkGroupRepository) GetAllUsersWorkTypeStats() (map[string][]models.WorkTypeStat, error) {
	var results []WorkTypeStatResult
	err := r.db.Table("work_group_members AS m").
		Select("m.user_id::text AS user_id, g.template_type AS work_type, COUNT(DISTINCT g.id) AS group_count").
		Joins("JOIN work_groups AS g ON g.id = m.group_id").
		Group("m.user_id, g.template_type").
		Order("m.user_id, group_count DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	statsMap := make(map[string][]models.WorkTypeStat)
	for _, r := range results {
		statsMap[r.UserID] = append(statsMap[r.UserID], models.WorkTypeStat{
			WorkType:   r.WorkType,
			GroupCount: r.GroupCount,
		})
	}
	return statsMap, nil
}

func (r *WorkGroupRepository) RecommendUsersByWorkType(workType string, excludeUserID string, limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = 10
	}
	var users []models.User
	err := r.db.Unscoped().Table("work_group_members AS m").
		Select("u.*, COUNT(DISTINCT g.id) AS participation_count").
		Joins("JOIN work_groups AS g ON g.id = m.group_id").
		Joins("JOIN users AS u ON u.id = m.user_id AND u.deleted_at IS NULL").
		Where("g.template_type = ? AND u.is_active = true", workType).
		Group("u.id").
		Order("participation_count DESC").
		Limit(limit).
		Preload("Department").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	if excludeUserID != "" {
		filtered := make([]models.User, 0, len(users))
		for _, u := range users {
			if u.ID.String() != excludeUserID {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}
	if len(users) == 0 {
		err := r.db.Preload("Department").
			Where("is_active = true").
			Order("name ASC").
			Limit(limit).
			Find(&users).Error
		return users, err
	}
	return users, nil
}

type WorkGroupSearchFilter struct {
	Keyword  string
	UserID   string
	DateFrom string
	DateTo   string
	Page     int
	PageSize int
}

func (r *WorkGroupRepository) Search(f WorkGroupSearchFilter) ([]models.WorkGroup, int64, error) {
	var groups []models.WorkGroup
	var total int64

	query := r.db.Model(&models.WorkGroup{})

	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", kw, kw)
	}
	if f.UserID != "" {
		subQuery := r.db.Model(&models.WorkGroupMember{}).Select("group_id").Where("user_id = ?", f.UserID)
		query = query.Where("initiator_id = ? OR id IN (?)", f.UserID, subQuery)
	}
	if f.DateFrom != "" {
		query = query.Where("created_at >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		query = query.Where("created_at <= ?", f.DateTo+"T23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	err := query.Preload("Members.User").Preload("Initiator").
		Order("created_at DESC").Offset(offset).Limit(f.PageSize).Find(&groups).Error
	return groups, total, err
}

type PresetGroupRepository struct {
	db *gorm.DB
}

func NewPresetGroupRepository(db *gorm.DB) *PresetGroupRepository {
	return &PresetGroupRepository{db: db}
}

func (r *PresetGroupRepository) Create(preset *models.PresetGroup) error {
	return r.db.Create(preset).Error
}

func (r *PresetGroupRepository) FindAll(templateType string) ([]models.PresetGroup, error) {
	var presets []models.PresetGroup
	query := r.db.Preload("Members.User").Preload("Creator")
	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	err := query.Order("created_at DESC").Find(&presets).Error
	return presets, err
}

func (r *PresetGroupRepository) FindByID(id string) (*models.PresetGroup, error) {
	var preset models.PresetGroup
	err := r.db.Preload("Members.User").Preload("Creator").First(&preset, "id = ?", id).Error
	return &preset, err
}

func (r *PresetGroupRepository) Update(id string, preset *models.PresetGroup) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PresetGroup{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":          preset.Name,
			"description":   preset.Description,
			"template_type": preset.TemplateType,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("preset_id = ?", id).Delete(&models.PresetGroupMember{}).Error; err != nil {
			return err
		}
		for _, m := range preset.Members {
			m.PresetID = preset.ID
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PresetGroupRepository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("preset_id = ?", id).Delete(&models.PresetGroupMember{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.PresetGroup{}).Error
	})
}

func (r *PresetGroupRepository) RecommendByWorkType(workType string, limit int) ([]models.PresetGroup, error) {
	if limit <= 0 {
		limit = 5
	}
	var presets []models.PresetGroup
	query := r.db.Preload("Members.User").Preload("Creator")
	if workType != "" {
		query = query.Where("template_type = ?", workType)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&presets).Error
	return presets, err
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
