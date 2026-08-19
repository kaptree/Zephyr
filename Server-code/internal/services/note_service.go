package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"labelpro-server/internal/database"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
)

type NoteService struct {
	noteRepo *repository.NoteRepository
	notifSvc *NotificationService
}

func NewNoteService(noteRepo *repository.NoteRepository, notifSvc *NotificationService) *NoteService {
	return &NoteService{noteRepo: noteRepo, notifSvc: notifSvc}
}

type CreateNoteRequest struct {
	Title        string               `json:"title" binding:"required"`
	Content      string               `json:"content"`
	TagIDs       []string             `json:"tags"`
	SourceType   string               `json:"source_type"`
	TemplateType string               `json:"template_type"`
	TemplateID   string               `json:"template_id"`
	OwnerID      string               `json:"owner_id"`
	DueTime      *time.Time           `json:"due_time"`
	WorkTimeSeconds int               `json:"work_time_seconds"`
	Assignees    []AssigneeRequest    `json:"assignees"`
	GroupConfig  *GroupConfigRequest  `json:"group_config"`
	CanvasConfig *CanvasConfigRequest `json:"canvas_config"`
}

type AssigneeRequest struct {
	UserID     string `json:"user_id"`
	RoleInNote string `json:"role_in_note"`
}

type GroupConfigRequest struct {
	GroupName string            `json:"group_name"`
	SubGroups []SubGroupRequest `json:"sub_groups"`
}

type SubGroupRequest struct {
	Name      string   `json:"name"`
	LeaderID  string   `json:"leader_id"`
	MemberIDs []string `json:"member_ids"`
}

type CanvasConfigRequest struct {
	Columns     int      `json:"columns"`
	ColumnUsers []string `json:"column_users"`
}

type UpdateNoteRequest struct {
	Title       *string    `json:"title"`
	Content     *string    `json:"content"`
	TagIDs      *[]string  `json:"tags"`
	DueTime     *time.Time `json:"due_time"`
	ColorStatus *string    `json:"color_status"`
	OwnerID     *string    `json:"owner_id"`
}

type CompleteNoteRequest struct {
	FeedbackContent string              `json:"feedback_content"`
	Attachments     []AttachmentRequest `json:"attachments"`
}

type AttachmentRequest struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

type RemindRequest struct {
	TargetID   string `json:"target_id" binding:"required"`
	Message    string `json:"message"`
	RemindType string `json:"remind_type"`
}

func (s *NoteService) Create(userID, role, deptID string, req CreateNoteRequest) (*models.Note, error) {
	creatorID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "self"
	}

	if role == "member" && sourceType == "assigned" {
		return nil, apperrors.ErrPermissionDenied
	}

	ownerID := creatorID
	if req.OwnerID != "" {
		oid, err := uuid.Parse(req.OwnerID)
		if err != nil {
			return nil, err
		}
		ownerID = oid
	}

	var deptUUID *uuid.UUID
	if deptID != "" {
		d, _ := uuid.Parse(deptID)
		deptUUID = &d
	}

	if req.TemplateID != "" {
		var tmpl models.Template
		if err := database.DB.First(&tmpl, "id = ?", req.TemplateID).Error; err == nil {
			var templateFields []map[string]interface{}
			if json.Unmarshal([]byte(tmpl.Fields), &templateFields) == nil {
				var fieldLines []string
				for _, f := range templateFields {
					if name, ok := f["name"].(string); ok {
						fieldLines = append(fieldLines, "【"+name+"】")
					}
				}
				if len(fieldLines) > 0 {
					templatePrefix := "📋 模板：" + tmpl.Name + "\n" + strings.Join(fieldLines, "\n") + "\n\n"
					req.Content = templatePrefix + req.Content
				}
			}
		}
	}

	now := time.Now()
	initialColorStatus := "yellow"
	if sourceType == "assigned" {
		initialColorStatus = "red"
	}
	if sourceType == "collaboration" {
		initialColorStatus = "blue"
	}

	// 工作时间选项：未显式指定截止时间时，按工作时间自动计算截止时间
	dueTime := req.DueTime
	if req.WorkTimeSeconds > 0 && dueTime == nil {
		t := now.Add(time.Duration(req.WorkTimeSeconds) * time.Second)
		dueTime = &t
	}

	note := &models.Note{
		Title:           req.Title,
		Content:         req.Content,
		ColorStatus:     initialColorStatus,
		SourceType:      sourceType,
		TemplateType:    req.TemplateType,
		CreatorID:       creatorID,
		OwnerID:         ownerID,
		DepartmentID:    deptUUID,
		DueTime:         dueTime,
		WorkTimeSeconds: req.WorkTimeSeconds,
	}

	if note.TemplateType == "" {
		note.TemplateType = "default"
	}

	if len(req.TagIDs) > 0 {
		for _, tid := range req.TagIDs {
			id, err := uuid.Parse(tid)
			if err != nil {
				continue
			}
			note.Tags = append(note.Tags, models.Tag{ID: id})
		}
	}

	note.Assignees = append(note.Assignees, models.NoteAssignee{
		UserID:     creatorID,
		RoleInNote: "initiator",
	})

	for _, a := range req.Assignees {
		uid, err := uuid.Parse(a.UserID)
		if err != nil {
			continue
		}
		roleInNote := a.RoleInNote
		if roleInNote == "" {
			roleInNote = "member"
		}
		note.Assignees = append(note.Assignees, models.NoteAssignee{
			UserID:     uid,
			RoleInNote: roleInNote,
		})
	}

	year := now.Year()
	seq, _ := s.noteRepo.GetNextSerialNumber(year)
	note.SerialNo = repository.GenerateSerialNo(year, seq)

	if err := s.noteRepo.Create(note); err != nil {
		return nil, err
	}

	_ = s.recordLedger(note.ID.String(), userID, "create", "任务创建", "", "")

	// 通知被指派者：任务已指派给你
	if s.notifSvc != nil {
		title := "您有新任务"
		content := "「" + note.Title + "」已指派给您，请及时处理"
		for _, a := range req.Assignees {
			if a.UserID != userID {
				_ = s.notifSvc.Notify(a.UserID, userID, &note.ID, "task_assigned", title, content)
			}
		}
	}

	return s.noteRepo.FindByID(note.ID.String())
}

func (s *NoteService) GetByID(id string) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, apperrors.ErrNoteNotFound
	}
	return note, nil
}

// CanAccess 判断用户是否有权访问该任务（创建人 / 负责人 / 被指派人，
// 部门管理员含本部门及子部门），用于详情查看、编辑、删除、导出、恢复等接口的统一越权拦截
func (s *NoteService) CanAccess(id, userID, role, deptID string) (bool, error) {
	return s.noteRepo.CheckUserAccess(id, userID, role, deptID)
}

func (s *NoteService) List(filter repository.NoteFilter, scope repository.NoteScope) ([]models.Note, int64, error) {
	return s.noteRepo.List(filter, scope)
}

func (s *NoteService) Update(id, userID string, req UpdateNoteRequest) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}

	if req.Title != nil {
		note.Title = *req.Title
	}
	if req.Content != nil {
		note.Content = *req.Content
	}
	if req.DueTime != nil {
		note.DueTime = req.DueTime
	}
	if req.ColorStatus != nil {
		note.ColorStatus = *req.ColorStatus
	}
	if req.OwnerID != nil {
		oid, err := uuid.Parse(*req.OwnerID)
		if err == nil {
			note.OwnerID = oid
		}
	}

	if req.TagIDs != nil {
		note.Tags = nil
		for _, tid := range *req.TagIDs {
			id, err := uuid.Parse(tid)
			if err != nil {
				continue
			}
			note.Tags = append(note.Tags, models.Tag{ID: id})
		}
	}

	if err := s.noteRepo.Update(note); err != nil {
		return nil, err
	}

	_ = s.recordLedger(id, userID, "update", "任务更新", "", "")

	return s.noteRepo.FindByID(id)
}

func (s *NoteService) Complete(id, userID, role string, req CompleteNoteRequest) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}

	isGroupNote := note.GroupID != nil
	isAssignee, _ := s.noteRepo.IsAssignee(id, userID)
	if !isGroupNote && note.SourceType == "assigned" && note.OwnerID.String() != userID &&
		!isAssignee && role != "super_admin" && role != "dept_admin" {
		return nil, apperrors.ErrPermissionDenied
	}

	now := time.Now()
	note.ColorStatus = "green"
	note.IsArchived = true
	note.ArchiveTime = &now
	note.CompletedAt = &now

	if err := s.noteRepo.Update(note); err != nil {
		return nil, err
	}

	// 回写被指派人的反馈内容到 note_assignees
	if req.FeedbackContent != "" {
		_ = s.noteRepo.UpdateAssigneeFeedback(note.ID.String(), userID, req.FeedbackContent)
	}

	_ = s.recordLedger(id, userID, "complete", "任务办结归档", req.FeedbackContent, "")

	// 通知任务发起人（被指派者完成任务时）
	if s.notifSvc != nil && note.CreatorID.String() != userID {
		notifType := "task_completed"
		title := "任务已完成"
		content := req.FeedbackContent
		if content == "" {
			content = "您指派的任务已完成归档"
		}
		_ = s.notifSvc.Notify(note.CreatorID.String(), userID, &note.ID, notifType, title, content)
	}

	return s.noteRepo.FindByID(id)
}

// Feedback 被指派人对已完成的任务补充反馈填报，并通知任务发起人
func (s *NoteService) Feedback(id, userID, content string) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}
	if content == "" {
		return nil, apperrors.ErrInvalidChatContent
	}

	// 仅任务发起人 / 被指派人可提交反馈
	isAssignee, _ := s.noteRepo.IsAssignee(id, userID)
	if note.CreatorID.String() != userID && !isAssignee {
		return nil, apperrors.ErrPermissionDenied
	}

	if err := s.noteRepo.UpdateAssigneeFeedback(note.ID.String(), userID, content); err != nil {
		return nil, err
	}

	_ = s.recordLedger(id, userID, "feedback", "任务反馈填报", content, "")

	if s.notifSvc != nil && note.CreatorID.String() != userID {
		_ = s.notifSvc.Notify(note.CreatorID.String(), userID, &note.ID, "task_feedback", "任务反馈", content)
	}

	return s.noteRepo.FindByID(id)
}

func (s *NoteService) Remind(id, userID string, req RemindRequest) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}

	reminderID, _ := uuid.Parse(userID)
	targetID, _ := uuid.Parse(req.TargetID)

	remindType := req.RemindType
	if remindType == "" {
		remindType = "normal"
	}

	now := time.Now()
	note.ColorStatus = "red"
	note.RemindCount++
	note.LastRemindAt = &now

	reminder := &models.Reminder{
		NoteID:     note.ID,
		ReminderID: reminderID,
		TargetID:   targetID,
		Message:    req.Message,
		RemindType: remindType,
	}

	if err := s.noteRepo.Update(note); err != nil {
		return nil, err
	}

	_ = s.noteRepo.CreateReminder(reminder)
	_ = s.recordLedger(id, userID, "remind", "盯办提醒", req.Message, "")

	// 通知被盯办人
	if s.notifSvc != nil && req.TargetID != userID {
		title := "任务催办提醒"
		content := req.Message
		if content == "" {
			content = "「" + note.Title + "」有新的催办提醒，请尽快处理"
		}
		_ = s.notifSvc.Notify(req.TargetID, userID, &note.ID, "task_remind", title, content)
	}

	return s.noteRepo.FindByID(id)
}

func (s *NoteService) Delete(id string, hardDelete bool) error {
	if hardDelete {
		return s.noteRepo.HardDelete(id)
	}
	return s.noteRepo.SoftDelete(id)
}

func (s *NoteService) Restore(id, userID string) (*models.Note, error) {
	if err := s.noteRepo.Restore(id); err != nil {
		return nil, err
	}
	_ = s.recordLedger(id, userID, "update", "任务恢复", "", "")
	return s.noteRepo.FindByID(id)
}

func (s *NoteService) recordLedger(noteID, userID, action, detail, ip, ua string) error {
	nid, _ := uuid.Parse(noteID)
	uid, _ := uuid.Parse(userID)

	entry := &models.LedgerEntry{
		NoteID:       nid,
		UserID:       uid,
		Action:       action,
		ActionDetail: detail,
		IPAddress:    ip,
		UserAgent:    ua,
	}
	return s.noteRepo.CreateLedger(entry)
}

type NoteStats struct {
	TotalNotes  int64                    `json:"total_notes"`
	ActiveNotes int64                    `json:"active_notes"`
	Trend       []repository.NoteDayStat `json:"trend"`
}

type NoteHeatmap struct {
	TotalArchived int64                    `json:"total_archived"`
	Year          int                      `json:"year"`
	Daily         []repository.NoteDayStat `json:"daily"`
}

func (s *NoteService) GetHeatmap(userID string, year int) (*NoteHeatmap, error) {
	total, err := s.noteRepo.CountArchivedByUser(userID)
	if err != nil {
		return nil, err
	}
	daily, err := s.noteRepo.HeatmapByYear(userID, year)
	if err != nil {
		return nil, err
	}
	return &NoteHeatmap{
		TotalArchived: total,
		Year:          year,
		Daily:         daily,
	}, nil
}

func (s *NoteService) GetStats(days int, deptID string, status string) (*NoteStats, error) {
	total, err := s.noteRepo.CountAll()
	if err != nil {
		return nil, err
	}
	active, err := s.noteRepo.CountActive()
	if err != nil {
		return nil, err
	}
	archivedOnly := status == "archived"
	var trend []repository.NoteDayStat
	if deptID != "" {
		trend, err = s.noteRepo.StatsByDayAndDept(days, deptID, archivedOnly)
	} else {
		trend, err = s.noteRepo.StatsByDay(days, archivedOnly)
	}
	if err != nil {
		return nil, err
	}
	return &NoteStats{
		TotalNotes:  total,
		ActiveNotes: active,
		Trend:       trend,
	}, nil
}

// ==================== 到期提醒调度 ====================

// dueRemindThreshold 到期提醒阈值：距截止时间不足该时长时发送提醒（与配置注释「到期前2小时」一致）
const dueRemindThreshold = 2 * time.Hour

// CheckDueReminders 扫描即将到期（剩余不足阈值）且尚未提醒过的未完成任务，向相关人发送到期提醒通知
func (s *NoteService) CheckDueReminders() error {
	var notes []models.Note
	now := time.Now()
	deadline := now.Add(dueRemindThreshold)

	if err := database.DB.
		Preload("Assignees.User").
		Where("due_time IS NOT NULL AND completed_at IS NULL AND is_archived = ? AND due_remind_at IS NULL AND due_time <= ?", false, deadline).
		Find(&notes).Error; err != nil {
		return err
	}

	for i := range notes {
		n := &notes[i]
		if n.DueTime == nil {
			continue
		}

		// 收集通知对象：任务负责人 + 所有被指派人（去重）
		targets := make(map[string]bool)
		if n.OwnerID != uuid.Nil {
			targets[n.OwnerID.String()] = true
		}
		if n.CreatorID != uuid.Nil {
			targets[n.CreatorID.String()] = true
		}
		for _, a := range n.Assignees {
			if a.UserID != uuid.Nil {
				targets[a.UserID.String()] = true
			}
		}

		dueFmt := n.DueTime.Format("2006-01-02 15:04")
		title := "【任务到期提醒】" + n.Title
		content := fmt.Sprintf("任务「%s」将于 %s 截止，剩余时间不足 2 小时，请尽快处理。", n.Title, dueFmt)
		noteID := n.ID
		for uid := range targets {
			_ = s.notifSvc.Notify(uid, "", &noteID, "task_due", title, content)
		}

		// 标记已提醒，避免重复发送
		markedAt := time.Now()
		_ = database.DB.Model(n).Update("due_remind_at", markedAt)
	}
	return nil
}

// StartDueRemindScheduler 启动到期提醒调度器（按配置间隔扫描，默认10分钟）
func (s *NoteService) StartDueRemindScheduler(intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}
	interval := time.Duration(intervalMinutes) * time.Minute
	go func() {
		// 启动后先执行一次，再按间隔循环
		for {
			if err := s.CheckDueReminders(); err != nil {
				// 静默失败，下一周期重试
				_ = err
			}
			time.Sleep(interval)
		}
	}()
}
