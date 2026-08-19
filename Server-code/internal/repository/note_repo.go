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
		Preload("Department").
		Preload("Assignees.User").
		Preload("Ccs.User")

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
	case "dept_admin":
		// 部门管理员：本部门（含子部门）任务 + 与自己相关（创建/负责/被指派/抄送）的任务
		// 修复 Bug2：跨部门指派的被指派人（如系统管理员派给部门管理员）能收到通知
		// 但此前列表仅按部门过滤导致不显示；此处与 CheckUserAccess 语义保持一致
		subDeptIDs, _ := r.getSubDeptIDs(scope.DepartmentID)
		query = query.Where(
			"notes.department_id IN ? OR notes.creator_id = ? OR notes.owner_id = ? OR notes.id IN (SELECT note_id FROM note_assignees WHERE user_id = ?) OR notes.id IN (SELECT note_id FROM note_ccs WHERE user_id = ?)",
			subDeptIDs, scope.UserID, scope.UserID, scope.UserID, scope.UserID,
		)
	default:
		// super_admin / group_leader / 普通用户统一按「与自己相关」过滤：
		// 创建人 / 负责人 / 被指派人 / 抄送人，防止管理员看到所有人员的任务
		query = query.Where(
			"notes.creator_id = ? OR notes.owner_id = ? OR notes.id IN (SELECT note_id FROM note_assignees WHERE user_id = ?) OR notes.id IN (SELECT note_id FROM note_ccs WHERE user_id = ?)",
			scope.UserID, scope.UserID, scope.UserID, scope.UserID,
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

// ListUserNotesForInspect 查询指定用户（创建人/负责人/被指派人/抄送人）的任务，
// 供公司领导/super_admin 在「用户工作台」板块查看对应用户的工作台内容（不受查看者自身范围限制）
func (r *NoteRepository) ListUserNotesForInspect(targetUserID, status string) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	query := r.db.Model(&models.Note{}).
		Preload("Tags").
		Preload("Creator").
		Preload("Owner").
		Preload("Assignees.User").
		Preload("Ccs.User").
		Where("(creator_id = ? OR owner_id = ? OR id IN (SELECT note_id FROM note_assignees WHERE user_id = ?) OR id IN (SELECT note_id FROM note_ccs WHERE user_id = ?))",
			targetUserID, targetUserID, targetUserID, targetUserID)

	switch status {
	case "archived":
		query = query.Where("is_archived = ?", true)
	case "completed":
		query = query.Where("color_status = ?", "green")
	default:
		query = query.Where("is_archived = ?", false)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Find(&notes).Error; err != nil {
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
		Preload("Ccs.User").
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

// ListRemindersForTarget 某人收到的盯办提醒列表（含任务标题）
func (r *NoteRepository) ListRemindersForTarget(userID string, out *[]models.Reminder) error {
	return r.db.Preload("Reminder").
		Preload("Target").
		Where("target_id = ?", userID).
		Order("created_at DESC").
		Find(out).Error
}

// AcknowledgeReminder 确认盯办提醒（is_acknowledged=true）
func (r *NoteRepository) AcknowledgeReminder(id, userID string) error {
	return r.db.Model(&models.Reminder{}).
		Where("id = ? AND target_id = ?", id, userID).
		Update("is_acknowledged", true).Error
}

func (r *NoteRepository) CreateAssignee(assignee *models.NoteAssignee) error {
	return r.db.Create(assignee).Error
}

// IsAssignee 判断用户是否为任务的被指派人
func (r *NoteRepository) IsAssignee(noteID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Count(&count).Error
	return count > 0, err
}

// UpdateAssigneeFeedback 回写被指派人的任务反馈内容
func (r *NoteRepository) UpdateAssigneeFeedback(noteID, userID string, content string) error {
	now := time.Now()
	return r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]interface{}{
			"feedback_content": content,
			"feedback_at":      now,
			"is_read":          true,
		}).Error
}

// GetAssigneeFeedback 获取某人对某任务的反馈
func (r *NoteRepository) GetAssigneeFeedback(noteID, userID string) (string, error) {
	var a models.NoteAssignee
	err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&a).Error
	if err != nil {
		return "", err
	}
	return a.FeedbackContent, nil
}

// UpdateAssigneeComplete 标记被指派人本人完成（并回写反馈内容）
func (r *NoteRepository) UpdateAssigneeComplete(noteID, userID, feedback string) error {
	now := time.Now()
	return r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]interface{}{
			"is_completed":     true,
			"completed_at":     now,
			"feedback_content": feedback,
			"feedback_at":      now,
			"is_read":          true,
		}).Error
}

// UpdateAssigneeSign 更新被指派人的任务签收状态
func (r *NoteRepository) UpdateAssigneeSign(noteID, userID string) error {
	now := time.Now()
	return r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]interface{}{
			"sign_status": "signed",
			"signed_at":   now,
		}).Error
}

// GetNoteCreatorID 获取任务发起人 ID
func (r *NoteRepository) GetNoteCreatorID(noteID string) (string, error) {
	var n models.Note
	if err := r.db.Select("creator_id").First(&n, "id = ?", noteID).Error; err != nil {
		return "", err
	}
	return n.CreatorID.String(), nil
}

// CheckUserAccess 判断用户是否有权查看任务（含抄送人）：
//   - 创建人 / 负责人 / 被指派人 / 抄送人 始终可访问
//   - dept_admin（部门管理员）可访问本部门（含子部门）的任务，与列表可见范围保持一致
func (r *NoteRepository) CheckUserAccess(noteID, userID, role, deptID string) (bool, error) {
	return r.checkAccess(noteID, userID, role, deptID, true)
}

// CheckParticipantAccess 判断用户是否有权管理任务（排除抄送人）：
// 抄送人仅可查看，不能编辑/删除/完成/盯办
func (r *NoteRepository) CheckParticipantAccess(noteID, userID, role, deptID string) (bool, error) {
	return r.checkAccess(noteID, userID, role, deptID, false)
}

func (r *NoteRepository) checkAccess(noteID, userID, role, deptID string, includeCc bool) (bool, error) {
	var note models.Note
	if err := r.db.Select("creator_id", "owner_id", "department_id").First(&note, "id = ?", noteID).Error; err != nil {
		return false, err
	}

	// 公司领导（company_leader）：权限大于部门管理员，可访问全公司任意任务
	if role == "company_leader" {
		return true, nil
	}

	if note.CreatorID.String() == userID || note.OwnerID.String() == userID {
		return true, nil
	}

	var count int64
	if err := r.db.Model(&models.NoteAssignee{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if includeCc {
		var ccCount int64
		if err := r.db.Model(&models.NoteCc{}).
			Where("note_id = ? AND user_id = ?", noteID, userID).
			Count(&ccCount).Error; err != nil {
			return false, err
		}
		if ccCount > 0 {
			return true, nil
		}
	}

	if role == "dept_admin" && deptID != "" && note.DepartmentID != nil {
		subDeptIDs, _ := r.getSubDeptIDs(deptID)
		for _, sid := range subDeptIDs {
			if note.DepartmentID.String() == sid {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsParticipant 判断用户是否为任务的参与者（创建人/负责人/被指派人，不含抄送人）
func (r *NoteRepository) IsParticipant(noteID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Note{}).
		Where("id = ? AND (creator_id = ? OR owner_id = ? OR EXISTS (SELECT 1 FROM note_assignees WHERE note_id = notes.id AND user_id = ?))",
			noteID, userID, userID, userID).
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

type SourceTypeStat struct {
	SourceType string `json:"source_type"`
	Count      int64  `json:"count"`
}

func (r *NoteRepository) SourceTypeDistribution(userID string, since time.Time) ([]SourceTypeStat, error) {
	var results []SourceTypeStat
	err := r.db.Model(&models.Note{}).
		Select("source_type, COUNT(*) as count").
		Where("creator_id = ? AND created_at >= ?", userID, since).
		Group("source_type").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

func (r *NoteRepository) CountArchivedByUser(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Note{}).
		Where("owner_id = ? AND is_archived = ?", userID, true).
		Count(&count).Error
	return count, err
}

func (r *NoteRepository) HeatmapByYear(userID string, year int) ([]NoteDayStat, error) {
	var stats []NoteDayStat
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)
	err := r.db.Model(&models.Note{}).
		Select("TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("owner_id = ?", userID).
		Where("is_archived = ?", true).
		Where("completed_at >= ? AND completed_at <= ?", startDate, endDate).
		Group("TO_CHAR(completed_at, 'YYYY-MM-DD')").Order("date ASC").
		Find(&stats).Error
	return stats, err
}

func (r *NoteRepository) HeatmapByYearAndDept(year int, deptID string) ([]NoteDayStat, error) {
	var stats []NoteDayStat
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)
	query := r.db.Model(&models.Note{}).
		Select("TO_CHAR(notes.completed_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Joins("LEFT JOIN users ON users.id = notes.owner_id").
		Where("notes.is_archived = ?", true).
		Where("notes.completed_at >= ? AND notes.completed_at <= ?", startDate, endDate)
	if deptID != "" {
		query = query.Where("users.department_id = ?", deptID)
	}
	err := query.Group("TO_CHAR(notes.completed_at, 'YYYY-MM-DD')").Order("date ASC").Find(&stats).Error
	return stats, err
}

// ==================== 团队工作成效统计 ====================

type TeamMemberStat struct {
	UserID             string  `json:"user_id"`
	UserName           string  `json:"user_name"`
	Username           string  `json:"username"`
	DeptName           string  `json:"dept_name"`
	TotalCreated       int64   `json:"total_created"`
	TotalCompleted     int64   `json:"total_completed"`
	CompletionRate     float64 `json:"completion_rate"`
	AvgCompletionHours float64 `json:"avg_completion_hours"`
	RemindReceived     int64   `json:"remind_received"`
}

type TeamStatsResult struct {
	Members        []TeamMemberStat `json:"members"`
	TotalCreated   int64            `json:"total_created"`
	TotalCompleted int64            `json:"total_completed"`
	CompletionRate float64          `json:"completion_rate"`
	MemberCount    int              `json:"member_count"`
}

// GetTeamStats 团队工作成效统计（按时间范围）
// deptID 为当前用户部门；role=super_admin 查看全部部门，其他角色查看本部门（含子部门）成员
// userIDs 非空时仅在可见成员范围内过滤指定成员（用于自定义勾选组建团队）
func (r *NoteRepository) GetTeamStats(since, now time.Time, deptID, role string, userIDs []string) (*TeamStatsResult, error) {
	var deptIDs []string
	allDepts := role == "super_admin"
	if !allDepts && deptID != "" {
		deptIDs, _ = r.getSubDeptIDs(deptID)
		if len(deptIDs) == 0 {
			deptIDs = []string{deptID}
		}
	}

	// 成员列表
	userQuery := r.db.Model(&models.User{}).
		Select("users.id AS user_id, users.name AS user_name, users.username AS username, COALESCE(departments.name, '') AS dept_name").
		Joins("LEFT JOIN departments ON departments.id = users.department_id").
		Where("users.is_active = ?", true)
	if !allDepts {
		userQuery = userQuery.Where("users.department_id IN ?", deptIDs)
	}
	var users []TeamMemberStat
	if err := userQuery.Order("dept_name ASC, user_name ASC").Scan(&users).Error; err != nil {
		return nil, err
	}

	// 自定义勾选成员：仅在可见范围内过滤指定成员
	if len(userIDs) > 0 {
		selected := make(map[string]bool, len(userIDs))
		for _, id := range userIDs {
			selected[id] = true
		}
		filtered := make([]TeamMemberStat, 0, len(users))
		for _, u := range users {
			if selected[u.UserID] {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	// 创建任务数（creator 维度）
	var createdRows []struct {
		UserID string `gorm:"column:user_id"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	r.db.Model(&models.Note{}).
		Select("creator_id AS user_id, COUNT(*) AS cnt").
		Where("created_at >= ? AND created_at <= ?", since, now).
		Group("creator_id").Scan(&createdRows)

	// 完成任务数（owner 维度）
	var completedRows []struct {
		UserID string `gorm:"column:user_id"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	r.db.Model(&models.Note{}).
		Select("owner_id AS user_id, COUNT(*) AS cnt").
		Where("completed_at >= ? AND completed_at <= ? AND is_archived = ?", since, now, true).
		Group("owner_id").Scan(&completedRows)

	// 平均完成耗时（小时）
	var avgRows []struct {
		UserID string  `gorm:"column:user_id"`
		Avg    float64 `gorm:"column:avg"`
	}
	r.db.Model(&models.Note{}).
		Select("owner_id AS user_id, AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600) AS avg").
		Where("completed_at >= ? AND completed_at <= ? AND is_archived = ? AND completed_at IS NOT NULL", since, now, true).
		Group("owner_id").Scan(&avgRows)

	// 被盯办次数
	var remindedRows []struct {
		UserID string `gorm:"column:user_id"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	r.db.Model(&models.Reminder{}).
		Select("target_id AS user_id, COUNT(*) AS cnt").
		Where("created_at >= ? AND created_at <= ?", since, now).
		Group("target_id").Scan(&remindedRows)

	createdMap := map[string]int64{}
	for _, v := range createdRows {
		createdMap[v.UserID] = v.Cnt
	}
	completedMap := map[string]int64{}
	for _, v := range completedRows {
		completedMap[v.UserID] = v.Cnt
	}
	avgMap := map[string]float64{}
	for _, v := range avgRows {
		avgMap[v.UserID] = v.Avg
	}
	remindedMap := map[string]int64{}
	for _, v := range remindedRows {
		remindedMap[v.UserID] = v.Cnt
	}

	var totalCreated, totalCompleted int64
	for i := range users {
		u := &users[i]
		u.TotalCreated = createdMap[u.UserID]
		u.TotalCompleted = completedMap[u.UserID]
		u.AvgCompletionHours = avgMap[u.UserID]
		u.RemindReceived = remindedMap[u.UserID]
		if u.TotalCreated > 0 {
			u.CompletionRate = float64(u.TotalCompleted) / float64(u.TotalCreated) * 100
		}
		totalCreated += u.TotalCreated
		totalCompleted += u.TotalCompleted
	}

	result := &TeamStatsResult{
		Members:        users,
		TotalCreated:   totalCreated,
		TotalCompleted: totalCompleted,
		MemberCount:    len(users),
	}
	if totalCreated > 0 {
		result.CompletionRate = float64(totalCompleted) / float64(totalCreated) * 100
	}
	return result, nil
}

var _ = strings.TrimSpace
