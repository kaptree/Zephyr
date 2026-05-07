package repository

import (
	"fmt"
	"strings"
	"time"

	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

type NoteFilter struct {
	Status       string
	SourceType   string
	TagIDs       []string
	OwnerID      string
	CreatorID    string
	DepartmentID string
	ColorStatus  string
	Keyword      string
	DateFrom     string
	DateTo       string
	IsUrgent     bool
	Page         int
	PageSize     int
	SortBy       string
	SortOrder    string
}

type NoteScope struct {
	UserID       string
	Role         string
	DepartmentID string
}

func (r *NoteRepository) List(filter NoteFilter, scope NoteScope) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.db.Model(&models.Note{}).
		Preload("Tags").
		Preload("Creator").
		Preload("Owner").
		Preload("Department")

	switch filter.Status {
	case "archived":
		query = query.Where("notes.is_archived = ?", true)
	case "completed":
		query = query.Where("notes.color_status = ?", "green")
	case "active", "":
		query = query.Where("notes.is_archived = ?", false)
	}

	if filter.SourceType != "" {
		query = query.Where("notes.source_type = ?", filter.SourceType)
	}
	if filter.OwnerID != "" {
		query = query.Where("notes.owner_id = ?", filter.OwnerID)
	}
	if filter.CreatorID != "" {
		query = query.Where("notes.creator_id = ?", filter.CreatorID)
	}
	if filter.DepartmentID != "" {
		query = query.Where("notes.department_id = ?", filter.DepartmentID)
	}
	if filter.ColorStatus != "" {
		query = query.Where("notes.color_status = ?", filter.ColorStatus)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("notes.title LIKE ? OR notes.content LIKE ?", keyword, keyword)
	}
	if filter.DateFrom != "" {
		query = query.Where("notes.created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("notes.created_at <= ?", filter.DateTo)
	}
	if filter.IsUrgent {
		query = query.Where("notes.due_time IS NOT NULL AND notes.due_time <= ? AND notes.is_archived = false",
			time.Now().Add(2*time.Hour))
	}

	if len(filter.TagIDs) > 0 {
		subQuery := r.db.Table("note_tags").
			Select("note_id").
			Where("tag_id IN ?", filter.TagIDs).
			Group("note_id").
			Having("COUNT(DISTINCT tag_id) = ?", len(filter.TagIDs))
		query = query.Where("notes.id IN (?)", subQuery)
	}

	switch scope.Role {
	case "super_admin":
	case "dept_admin":
		subDeptIDs, _ := r.getSubDeptIDs(scope.DepartmentID)
		query = query.Where("notes.department_id IN ?", subDeptIDs)
	case "group_leader":
		query = query.Where(
			"notes.creator_id = ? OR notes.owner_id = ? OR notes.id IN (SELECT note_id FROM note_assignees WHERE user_id = ?)",
			scope.UserID, scope.UserID, scope.UserID,
		)
	default:
		query = query.Where(
			"notes.creator_id = ? OR notes.owner_id = ? OR notes.id IN (SELECT note_id FROM note_assignees WHERE user_id = ?)",
			scope.UserID, scope.UserID, scope.UserID,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "notes.created_at"
	sortOrder := "desc"
	if filter.SortBy != "" {
		allowedSortFields := map[string]bool{"created_at": true, "updated_at": true, "due_time": true, "title": true}
		if allowedSortFields[filter.SortBy] {
			sortBy = "notes." + filter.SortBy
		}
	}
	if filter.SortOrder == "asc" {
		sortOrder = "asc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Order(orderClause).Find(&notes).Error; err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

func (r *NoteRepository) FindByID(id string) (*models.Note, error) {
	var note models.Note
	err := r.db.
		Preload("Tags").
		Preload("Creator").
		Preload("Owner").
		Preload("Department").
		Preload("Assignees.User").
		Preload("Attachments").
		Preload("Group.Members.User").
		Preload("Reminders.Reminder").
		Preload("Reminders.Target").
		First(&note, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepository) Create(note *models.Note) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		assignees := note.Assignees
		note.Assignees = nil

		if err := tx.Create(note).Error; err != nil {
			return err
		}

		if len(note.Tags) > 0 {
			if err := tx.Model(note).Association("Tags").Replace(note.Tags); err != nil {
				return err
			}
		}

		for i := range assignees {
			assignees[i].NoteID = note.ID
		}
		if len(assignees) > 0 {
			if err := tx.Create(&assignees).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *NoteRepository) Update(note *models.Note) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(note).Error; err != nil {
			return err
		}

		if len(note.Tags) > 0 || note.Tags != nil {
			if err := tx.Model(note).Association("Tags").Replace(note.Tags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *NoteRepository) SoftDelete(id string) error {
	return r.db.Delete(&models.Note{}, "id = ?", id).Error
}

func (r *NoteRepository) HardDelete(id string) error {
	return r.db.Unscoped().Delete(&models.Note{}, "id = ?", id).Error
}

func (r *NoteRepository) Restore(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Model(&models.Note{}).Where("id = ?", uid).
		Updates(map[string]interface{}{
			"is_archived":  false,
			"archive_time": nil,
			"deleted_at":   nil,
		}).Error
}

func (r *NoteRepository) CreateLedger(entry *models.LedgerEntry) error {
	return r.db.Create(entry).Error
}

func (r *NoteRepository) CreateReminder(reminder *models.Reminder) error {
	return r.db.Create(reminder).Error
}

func (r *NoteRepository) CreateAssignee(assignee *models.NoteAssignee) error {
	return r.db.Create(assignee).Error
}

func (r *NoteRepository) CheckUserAccess(noteID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Note{}).
		Where("id = ? AND (creator_id = ? OR owner_id = ?)", noteID, userID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	err = r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *NoteRepository) getSubDeptIDs(deptID string) ([]string, error) {
	if deptID == "" {
		return []string{}, nil
	}

	var allDepts []models.Department
	if err := r.db.Find(&allDepts).Error; err != nil {
		return nil, err
	}

	var subIDs []string
	subIDs = append(subIDs, deptID)
	collectSubDepts(deptID, allDepts, &subIDs)
	return subIDs, nil
}

func collectSubDepts(parentID string, allDepts []models.Department, result *[]string) {
	for _, d := range allDepts {
		if d.ParentID != nil && d.ParentID.String() == parentID {
			*result = append(*result, d.ID.String())
			collectSubDepts(d.ID.String(), allDepts, result)
		}
	}
}

func (r *NoteRepository) GetNextSerialNumber(year int) (int, error) {
	var maxNo int
	prefix := fmt.Sprintf("资警轻燕〔%d〕", year)

	err := r.db.Model(&models.Note{}).
		Select("COALESCE(MAX(CAST(SUBSTRING(serial_no FROM '%s#\"\\d+#\"%' FOR '#') AS INTEGER)), 0)").
		Where("serial_no LIKE ?", prefix+"%").
		Pluck("COALESCE(MAX(CAST(SUBSTRING(serial_no FROM '%s#\"\\d+#\"%' FOR '#') AS INTEGER)), 0)", &maxNo).Error

	if err != nil || maxNo == 0 {
		var count int64
		r.db.Model(&models.Note{}).Where("serial_no LIKE ?", prefix+"%").Count(&count)
		maxNo = int(count)
	}

	return maxNo + 1, nil
}

func GenerateSerialNo(year, seq int) string {
	return fmt.Sprintf("资警轻燕〔%d〕%04d号", year, seq)
}

type NoteDayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (r *NoteRepository) StatsByDay(days int, archivedOnly bool) ([]NoteDayStat, error) {
	var stats []NoteDayStat
	query := r.db.Model(&models.Note{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -days))
	if archivedOnly {
		query = query.Where("is_archived = ?", true)
	}
	err := query.Group("DATE(created_at)").Order("date ASC").Find(&stats).Error
	return stats, err
}

func (r *NoteRepository) StatsByDayAndDept(days int, deptID string, archivedOnly bool) ([]NoteDayStat, error) {
	var stats []NoteDayStat
	query := r.db.Model(&models.Note{}).
		Select("DATE(notes.created_at) as date, COUNT(*) as count").
		Joins("LEFT JOIN users ON users.id = notes.owner_id").
		Where("notes.created_at >= ?", time.Now().AddDate(0, 0, -days))
	if deptID != "" {
		query = query.Where("users.department_id = ?", deptID)
	}
	if archivedOnly {
		query = query.Where("notes.is_archived = ?", true)
	}
	err := query.Group("DATE(notes.created_at)").Order("date ASC").Find(&stats).Error
	return stats, err
}

func (r *NoteRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.Note{}).Count(&count).Error
	return count, err
}

func (r *NoteRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Note{}).Where("is_archived = false").Count(&count).Error
	return count, err
}

func (r *NoteRepository) CountByDept(deptID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Note{}).
		Joins("JOIN users ON users.id = notes.owner_id").
		Where("users.department_id = ?", deptID).
		Count(&count).Error
	return count, err
}

type PersonalStats struct {
	TotalCreated       int64          `json:"total_created"`
	TotalCompleted     int64          `json:"total_completed"`
	CompletionRate     float64        `json:"completion_rate"`
	RemindReceived     int64          `json:"remind_received"`
	AvgCompletionHours float64        `json:"avg_completion_hours"`
	DailyTrend         []NoteDayStat  `json:"daily_trend"`
	TagBreakdown       []TagBreakdown `json:"tag_breakdown"`
}

type TagBreakdown struct {
	TagName string `json:"tag_name"`
	Count   int64  `json:"count"`
}

func (r *NoteRepository) GetPersonalStats(userID string, days int) (*PersonalStats, error) {
	stats := &PersonalStats{}
	since := time.Now().AddDate(0, 0, -days)

	r.db.Model(&models.Note{}).Where("creator_id = ? AND created_at >= ?", userID, since).Count(&stats.TotalCreated)

	r.db.Model(&models.Note{}).
		Where("owner_id = ? AND is_archived = ? AND completed_at >= ?", userID, true, since).
		Count(&stats.TotalCompleted)

	if stats.TotalCreated > 0 {
		stats.CompletionRate = float64(stats.TotalCompleted) / float64(stats.TotalCreated) * 100
	}

	r.db.Model(&models.Reminder{}).Where("target_id = ? AND created_at >= ?", userID, since).Count(&stats.RemindReceived)

	var dailyTrend []NoteDayStat
	r.db.Model(&models.Note{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("creator_id = ? AND created_at >= ?", userID, since).
		Group("DATE(created_at)").Order("date ASC").Find(&dailyTrend)
	stats.DailyTrend = dailyTrend
	if stats.DailyTrend == nil {
		stats.DailyTrend = []NoteDayStat{}
	}

	var tagBreakdown []TagBreakdown
	r.db.Table("note_tags").
		Select("tags.name as tag_name, COUNT(note_tags.note_id) as count").
		Joins("JOIN tags ON tags.id = note_tags.tag_id").
		Joins("JOIN notes ON notes.id = note_tags.note_id").
		Where("notes.creator_id = ? AND notes.created_at >= ?", userID, since).
		Group("tags.name").Order("count DESC").Limit(10).Find(&tagBreakdown)
	stats.TagBreakdown = tagBreakdown
	if stats.TagBreakdown == nil {
		stats.TagBreakdown = []TagBreakdown{}
	}

	rows, err := r.db.Model(&models.Note{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600)").
		Where("owner_id = ? AND is_archived = ? AND completed_at IS NOT NULL AND completed_at >= ?", userID, true, since).
		Rows()
	if err == nil && rows.Next() {
		var avgHours *float64
		rows.Scan(&avgHours)
		if avgHours != nil {
			stats.AvgCompletionHours = *avgHours
		}
		rows.Close()
	}

	return stats, nil
}

func (r *NoteRepository) ListByGroup(groupID string, userID string, page, pageSize int) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.db.Model(&models.Note{}).Where("group_id = ?", groupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Tags").Preload("Assignees").Preload("Group").Preload("Attachments").
		Preload("Reminders").Preload("Reminders.Reminder").Preload("Reminders.Target").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notes).Error
	return notes, total, err
}

func (r *NoteRepository) ListByGroupCompleted(groupID string) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Model(&models.Note{}).
		Where("group_id = ? AND color_status = ?", groupID, "green").
		Preload("Tags").
		Preload("Owner").
		Order("completed_at DESC").
		Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) ListAllByGroup(groupID string) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Model(&models.Note{}).
		Where("group_id = ?", groupID).
		Preload("Tags").
		Preload("Owner").
		Preload("Creator").
		Order("created_at ASC").
		Find(&notes).Error
	return notes, err
}

var _ = strings.TrimSpace
