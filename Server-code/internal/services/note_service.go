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
	Title           string               `json:"title" binding:"required"`
	Content         string               `json:"content"`
	TagIDs          []string             `json:"tags"`
	SourceType      string               `json:"source_type"`
	TemplateType    string               `json:"template_type"`
	TemplateID      string               `json:"template_id"`
	OwnerID         string               `json:"owner_id"`
	DueTime         *time.Time           `json:"due_time"`
	WorkTimeSeconds int                  `json:"work_time_seconds"`
	Assignees       []AssigneeRequest    `json:"assignees"`
	CcUserIDs       []string             `json:"cc_user_ids"`
	GroupConfig     *GroupConfigRequest  `json:"group_config"`
	CanvasConfig    *CanvasConfigRequest `json:"canvas_config"`
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
	IsPinned    *bool      `json:"is_pinned"`
	// 创建人创建后编辑参与人员/抄送人员（增删改）：非 nil 时按全量列表 diff，
	// nil 表示不调整该项
	Assignees *[]AssigneeRequest `json:"assignees"`
	CcUserIDs *[]string          `json:"cc_user_ids"`
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

// collabTemplateTypes 支持协同画布的模板类型（与种子数据保持一致）
var collabTemplateTypes = map[string]bool{
	"emergency_canvas":      true,
	"collaborative_writing": true,
	"data_analysis":         true,
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

	// 抄送人（需求20）：无论是否指派都可多选抄送，抄送人仅查看（紫色卡片 +「抄送」徽章）
	ccSeen := map[string]bool{}
	for _, cid := range req.CcUserIDs {
		id, err := uuid.Parse(cid)
		if err != nil || ccSeen[cid] || id == creatorID {
			continue
		}
		ccSeen[cid] = true
		note.Ccs = append(note.Ccs, models.NoteCc{UserID: id})
	}

	year := now.Year()
	seq, _ := s.noteRepo.GetNextSerialNumber(year)
	note.SerialNo = repository.GenerateSerialNo(year, seq)

	if err := s.noteRepo.Create(note); err != nil {
		return nil, err
	}

	// 协同房间（需求29）：画布类任务自动创建房间，否则画布/指令接口对新建任务始终404
	if collabTemplateTypes[note.TemplateType] || req.CanvasConfig != nil {
		columns := 1
		if req.CanvasConfig != nil && req.CanvasConfig.Columns > 0 {
			columns = req.CanvasConfig.Columns
		}
		room := &models.CollaborationRoom{
			NoteID:     note.ID,
			CanvasData: `{"blocks": [], "connections": []}`,
			Columns:    columns,
			IsActive:   true,
		}
		// 房间创建失败不阻断任务创建
		_ = database.DB.Create(room).Error
	}

	_ = s.recordLedger(note.ID.String(), userID, "create", "任务创建", "", "")

	// 积分：创建任务（交办给他人时额外得交办分，受指派人≠创建人防自刷）
	AwardPoints(userID, "task_create", "note", note.ID.String(), "创建任务《"+truncateNote(note.Title)+"》")
	if sourceType == "assigned" {
		for _, a := range req.Assignees {
			if a.UserID != userID {
				AwardPoints(userID, "task_create_assigned", "note", note.ID.String(),
					"创建交办任务《"+truncateNote(note.Title)+"》")
				break
			}
		}
	}

	// 通知被指派者：任务已指派给你
	if s.notifSvc != nil {
		title := "您有新任务"
		content := "「" + note.Title + "」已指派给您，请及时处理"
		for _, a := range req.Assignees {
			if a.UserID != userID {
				_ = s.notifSvc.Notify(a.UserID, userID, &note.ID, "task_assigned", title, content)
			}
		}
		// 通知抄送人：任务抄送
		for _, cid := range req.CcUserIDs {
			if cid != userID {
				_ = s.notifSvc.Notify(cid, userID, &note.ID, "task_cc", "您收到任务抄送", "「"+note.Title+"」已抄送给您，请查阅")
			}
		}
		// 需求30：别人指派/抄送的任务，工作台实时动态刷新（推送 note:updated 事件）
		// 注意：gorm Create 后会清空 note.Assignees 内存切片，故使用 req.Assignees
		for _, a := range req.Assignees {
			if a.UserID == userID {
				continue
			}
			s.notifSvc.PushNoteUpdate(a.UserID, &note.ID, "created")
		}
		for _, cid := range req.CcUserIDs {
			if cid == userID {
				continue
			}
			s.notifSvc.PushNoteUpdate(cid, &note.ID, "created")
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

// CanAccess 判断用户是否有权访问该任务（创建人 / 负责人 / 被指派人 / 抄送人，
// 部门管理员含本部门及子部门），用于详情查看等只读接口的统一越权拦截
func (s *NoteService) CanAccess(id, userID, role, deptID string) (bool, error) {
	return s.noteRepo.CheckUserAccess(id, userID, role, deptID)
}

// CanAccessIncludeDeleted 含已软删记录的访问校验（恢复已删除任务场景）
func (s *NoteService) CanAccessIncludeDeleted(id, userID, role, deptID string) (bool, error) {
	return s.noteRepo.CheckUserAccessIncludeDeleted(id, userID, role, deptID)
}

// CanManage 判断用户是否有权管理该任务（排除抄送人——抄送人仅可查看，
// 不能编辑/删除/完成/盯办），用于写操作接口的统一越权拦截
func (s *NoteService) CanManage(id, userID, role, deptID string) (bool, error) {
	return s.noteRepo.CheckParticipantAccess(id, userID, role, deptID)
}

func (s *NoteService) List(filter repository.NoteFilter, scope repository.NoteScope) ([]models.Note, int64, error) {
	return s.noteRepo.List(filter, scope)
}

// ListUserNotesForInspect 查询指定用户的工作台任务（公司领导/super_admin 专用）
func (s *NoteService) ListUserNotesForInspect(targetUserID, status string) ([]models.Note, int64, error) {
	return s.noteRepo.ListUserNotesForInspect(targetUserID, status)
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
	if req.IsPinned != nil {
		// 置顶切换：首次置顶记录 pinned_at（重复置顶不刷新时间），取消置顶清空
		if *req.IsPinned {
			if !note.IsPinned {
				now := time.Now()
				note.PinnedAt = &now
			}
		} else {
			note.PinnedAt = nil
		}
		note.IsPinned = *req.IsPinned
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

	// 创建人编辑参与人员/抄送人员（增删改）
	if req.Assignees != nil || req.CcUserIDs != nil {
		if err := s.updateParticipants(note, userID, req); err != nil {
			return nil, err
		}
	}

	_ = s.recordLedger(id, userID, "update", "任务更新", "", "")

	return s.noteRepo.FindByID(id)
}

// updateParticipants 创建人编辑任务参与人员（被指派人）与抄送人员的增删改：
// 仅创建人可操作；按全量列表与现有数据 diff，未变动人员保留签收/反馈/完成进度；
// 被移除者的工作台不再展示该任务（列表查询以 note_assignees / note_ccs 行为准），
// 并实时推送 removed 事件使对方前端立即移除任务卡片。
func (s *NoteService) updateParticipants(note *models.Note, userID string, req UpdateNoteRequest) error {
	if note.CreatorID.String() != userID {
		return apperrors.ErrPermissionDenied
	}
	if note.IsArchived {
		return apperrors.ErrNoteArchived
	}

	creatorID := note.CreatorID

	var addAssignees []models.NoteAssignee
	var removeAssigneeIDs []uuid.UUID
	var newAssigneeIDs []uuid.UUID

	// ---- 被指派人 diff（发起人角色始终保留，不参与增删） ----
	if req.Assignees != nil {
		seen := map[uuid.UUID]bool{creatorID: true}
		for _, a := range *req.Assignees {
			uid, err := uuid.Parse(a.UserID)
			if err != nil || seen[uid] {
				continue
			}
			seen[uid] = true
			newAssigneeIDs = append(newAssigneeIDs, uid)
		}

		currentAssignees := make(map[uuid.UUID]models.NoteAssignee)
		for _, a := range note.Assignees {
			if a.RoleInNote == "initiator" {
				continue
			}
			currentAssignees[a.UserID] = a
		}
		for _, uid := range newAssigneeIDs {
			if _, ok := currentAssignees[uid]; !ok {
				addAssignees = append(addAssignees, models.NoteAssignee{UserID: uid, RoleInNote: "member"})
			}
		}
		for uid := range currentAssignees {
			if !seen[uid] {
				removeAssigneeIDs = append(removeAssigneeIDs, uid)
			}
		}
	}

	// ---- 抄送人 diff（创建人不可抄送给自己） ----
	var addCcs []models.NoteCc
	var removeCcIDs []uuid.UUID
	if req.CcUserIDs != nil {
		ccSeen := map[uuid.UUID]bool{creatorID: true}
		for _, cid := range *req.CcUserIDs {
			uid, err := uuid.Parse(cid)
			if err != nil || ccSeen[uid] {
				continue
			}
			ccSeen[uid] = true
			if !currentCcSet(note, uid) {
				addCcs = append(addCcs, models.NoteCc{UserID: uid, CreatedAt: time.Now()})
			}
		}
		for _, c := range note.Ccs {
			if !ccSeen[c.UserID] {
				removeCcIDs = append(removeCcIDs, c.UserID)
			}
		}
	}

	// 原负责人被移出指派时同步转移负责人：优先指派列表首位，否则回退创建人
	var newOwnerID *uuid.UUID
	for _, uid := range removeAssigneeIDs {
		if uid == note.OwnerID {
			owner := creatorID
			if len(newAssigneeIDs) > 0 {
				owner = newAssigneeIDs[0]
			}
			newOwnerID = &owner
			break
		}
	}

	if err := s.noteRepo.UpdateParticipants(note.ID.String(), addAssignees, removeAssigneeIDs, addCcs, removeCcIDs, newOwnerID); err != nil {
		return err
	}

	_ = s.recordLedger(note.ID.String(), userID, "update",
		fmt.Sprintf("编辑参与人员（+%d -%d）/ 抄送人员（+%d -%d）",
			len(addAssignees), len(removeAssigneeIDs), len(addCcs), len(removeCcIDs)), "", "")

	// 通知与实时推送：新增者收到指派/抄送通知（任务出现在其工作台），
	// 被移除者收到移出通知（任务立即从其工作台消失）
	if s.notifSvc != nil {
		actorID := userID
		title := note.Title
		for _, a := range addAssignees {
			_ = s.notifSvc.Notify(a.UserID.String(), actorID, &note.ID, "task_assigned",
				"您有新任务", "「"+title+"」已指派给您，请及时处理")
			s.notifSvc.PushNoteUpdate(a.UserID.String(), &note.ID, "created")
		}
		for _, uid := range removeAssigneeIDs {
			_ = s.notifSvc.Notify(uid.String(), actorID, &note.ID, "task_removed",
				"您已被移出任务", "「"+title+"」已将您移出被指派人员，该任务不再对您可见")
			s.notifSvc.PushNoteUpdate(uid.String(), &note.ID, "removed")
		}
		for _, c := range addCcs {
			_ = s.notifSvc.Notify(c.UserID.String(), actorID, &note.ID, "task_cc",
				"您收到任务抄送", "「"+title+"」已抄送给您，请查阅")
			s.notifSvc.PushNoteUpdate(c.UserID.String(), &note.ID, "created")
		}
		for _, uid := range removeCcIDs {
			_ = s.notifSvc.Notify(uid.String(), actorID, &note.ID, "task_removed",
				"您已被移出任务抄送", "「"+title+"」已将您移出抄送人员")
			s.notifSvc.PushNoteUpdate(uid.String(), &note.ID, "removed")
		}
	}

	return nil
}

// currentCcSet 判断用户是否已在任务抄送人中
func currentCcSet(note *models.Note, userID uuid.UUID) bool {
	for _, c := range note.Ccs {
		if c.UserID == userID {
			return true
		}
	}
	return false
}

// FeedbackRoleInNote 用户在任务中的角色标识（时间轴徽章用）
// creator 创建人 / assigner 交办人 / assignee 被指派人 / cc 抄送人 / unknown
func FeedbackRoleInNote(note *models.Note, userID string) string {
	switch {
	case note.CreatorID.String() == userID:
		return "creator"
	case note.AssignerID != nil && note.AssignerID.String() == userID:
		return "assigner"
	}
	for i := range note.Assignees {
		if note.Assignees[i].UserID.String() == userID {
			return "assignee"
		}
	}
	for i := range note.Ccs {
		if note.Ccs[i].UserID.String() == userID {
			return "cc"
		}
	}
	return "unknown"
}

// feedbackDigest 阶段性反馈通知摘要（压平换行并截断 50 字）
func feedbackDigest(s string) string {
	r := []rune(strings.TrimSpace(strings.ReplaceAll(s, "\n", " ")))
	if len(r) > 50 {
		return string(r[:50]) + "…"
	}
	return string(r)
}

// FeedbackTimelineResp 阶段性反馈时间轴响应：正序反馈流 + 参与人统计
type FeedbackTimelineResp struct {
	Feedbacks    []models.NoteFeedback               `json:"feedbacks"`
	Participants []repository.FeedbackParticipantStat `json:"participants"`
}

// ListStageFeedbacks 任务阶段性反馈时间轴（论坛式，时间正序 + 参与人统计）
func (s *NoteService) ListStageFeedbacks(id string) (*FeedbackTimelineResp, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}
	feedbacks, err := s.noteRepo.ListFeedbacks(note.ID.String())
	if err != nil {
		return nil, err
	}
	participants, err := s.noteRepo.ListFeedbackParticipants(note.ID.String())
	if err != nil {
		return nil, err
	}
	return &FeedbackTimelineResp{Feedbacks: feedbacks, Participants: participants}, nil
}

// AddStageFeedback 追加一条阶段性反馈：权限校验 → 落库 → 同步最新反馈快照 →
// 积分（复用 feedback_submit 规则）→ 通知交办人及其他被指派人 → 台账
func (s *NoteService) AddStageFeedback(id, userID, content string) (*models.NoteFeedback, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, apperrors.ErrInvalidChatContent
	}
	if len([]rune(content)) > 2000 {
		return nil, apperrors.ErrInvalidOperation
	}

	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}
	if note.IsArchived {
		return nil, apperrors.ErrNoteArchived
	}

	// 仅创建人 / 交办人 / 被指派人可提交；抄送人与无关人员 403
	role := FeedbackRoleInNote(note, userID)
	if role != "creator" && role != "assigner" && role != "assignee" {
		return nil, apperrors.ErrPermissionDenied
	}

	uid, _ := uuid.Parse(userID)
	fb := &models.NoteFeedback{
		NoteID:  note.ID,
		UserID:  uid,
		Content: content,
	}
	// 冗余姓名：优先取关联关系中的 User.Name，时间轴渲染不依赖二次查询
	switch role {
	case "creator":
		if note.Creator != nil {
			fb.UserName = note.Creator.Name
		}
	case "assigner":
		if note.Assigner != nil {
			fb.UserName = note.Assigner.Name
		}
	default:
		for i := range note.Assignees {
			if note.Assignees[i].UserID.String() == userID && note.Assignees[i].User != nil {
				fb.UserName = note.Assignees[i].User.Name
				break
			}
		}
	}
	if err := s.noteRepo.CreateFeedback(fb); err != nil {
		return nil, err
	}

	// 受指派人提交：同步最新反馈快照（报表/积分"最新反馈"口径不变）
	if role == "assignee" {
		_ = s.noteRepo.UpdateAssigneeFeedback(note.ID.String(), userID, content)
	}

	_ = s.recordLedger(note.ID.String(), userID, "feedback", "阶段性反馈", "", "")

	// 积分：有效反馈（≥10 字），防刷由积分侧每日上限兜底
	if len([]rune(content)) >= 10 {
		AwardPoints(userID, "feedback_submit", "note", note.ID.String(), "阶段性反馈《"+truncateNote(note.Title)+"》")
	}

	// 通知交办人与其他被指派人（排除本人、去重；抄送人不打扰）
	if s.notifSvc != nil {
		recipients := map[string]bool{}
		recipients[note.CreatorID.String()] = true
		if note.AssignerID != nil {
			recipients[note.AssignerID.String()] = true
		}
		for i := range note.Assignees {
			recipients[note.Assignees[i].UserID.String()] = true
		}
		delete(recipients, userID)
		for recID := range recipients {
			_ = s.notifSvc.Notify(recID, userID, &note.ID, "task_feedback", "阶段性反馈", feedbackDigest(content))
		}
	}

	return fb, nil
}

// syncFeedbackToTimeline 把「提交完成 / 反馈填报」的内容同步写入阶段性反馈时间轴（note_feedbacks），
// 保证工作台详情的时间轴与 note_assignees 快照口径一致；冗余姓名从已 Preload 的关联用户尽力填充。
func (s *NoteService) syncFeedbackToTimeline(note *models.Note, userID, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return
	}
	fb := &models.NoteFeedback{
		NoteID:  note.ID,
		UserID:  uid,
		Content: content,
	}
	if note.Creator != nil && note.CreatorID.String() == userID {
		fb.UserName = note.Creator.Name
	}
	if fb.UserName == "" {
		for i := range note.Assignees {
			if note.Assignees[i].UserID.String() == userID && note.Assignees[i].User != nil {
				fb.UserName = note.Assignees[i].User.Name
				break
			}
		}
	}
	_ = s.noteRepo.CreateFeedback(fb)
}

// NotePositionInput 单个任务的画布位置
type NotePositionInput struct {
	ID   string `json:"id" binding:"required"`
	PosX int    `json:"pos_x"`
	PosY int    `json:"pos_y"`
}

// UpdateNotePositionsRequest 批量更新任务位置（一次最多 200 条）
type UpdateNotePositionsRequest struct {
	Positions []NotePositionInput `json:"positions" binding:"required,min=1,max=200,dive"`
}

// canArrangeNote 判断用户是否有权调整该任务位置（创建者/负责人/被指派人/抄送人）
func canArrangeNote(n models.Note, uid uuid.UUID) bool {
	if n.OwnerID == uid || n.CreatorID == uid {
		return true
	}
	for _, a := range n.Assignees {
		if a.UserID == uid {
			return true
		}
	}
	for _, c := range n.Ccs {
		if c.UserID == uid {
			return true
		}
	}
	return false
}

// UpdatePositions 批量更新任务在工作台画布上的位置；无权限或无效的任务会被跳过，返回实际更新数量
func (s *NoteService) UpdatePositions(userID string, req UpdateNotePositionsRequest) (int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, apperrors.ErrPermissionDenied
	}

	seen := make(map[string]bool, len(req.Positions))
	posByID := make(map[string][2]int, len(req.Positions))
	ids := make([]string, 0, len(req.Positions))
	for _, p := range req.Positions {
		if p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		ids = append(ids, p.ID)
		posByID[p.ID] = [2]int{p.PosX, p.PosY}
	}
	if len(ids) == 0 {
		return 0, nil
	}

	notes, err := s.noteRepo.FindByIDsForAccessCheck(ids)
	if err != nil {
		return 0, err
	}

	allowed := make(map[string][2]int, len(notes))
	for _, n := range notes {
		if !canArrangeNote(n, uid) {
			continue
		}
		if pos, ok := posByID[n.ID.String()]; ok {
			allowed[n.ID.String()] = pos
		}
	}
	if len(allowed) == 0 {
		return 0, nil
	}

	if err := s.noteRepo.UpdatePositions(allowed); err != nil {
		return 0, err
	}
	return len(allowed), nil
}

// memberAssigneeList 返回任务的实际被指派人（排除发起者 initiator）
func memberAssigneeList(note *models.Note) []models.NoteAssignee {
	var members []models.NoteAssignee
	for _, a := range note.Assignees {
		if a.RoleInNote != "initiator" {
			members = append(members, a)
		}
	}
	return members
}

// allAssigneesCompleted 判断任务的所有被指派人是否均已本人完成
func allAssigneesCompleted(note *models.Note) bool {
	for _, a := range memberAssigneeList(note) {
		if !a.IsCompleted {
			return false
		}
	}
	return true
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

	// 需求23：指派任务 —— 被指派人各自完成本人部分并填报反馈；
	// 仅当所有被指派人完成后，发起者才可归档，是否归档由发起者决定。
	members := memberAssigneeList(note)
	if note.SourceType == "assigned" && len(members) > 0 {
		// 被指派人（非发起者）提交本人完成，不归档任务
		if isAssignee && note.CreatorID.String() != userID {
			if err := s.noteRepo.UpdateAssigneeComplete(note.ID.String(), userID, req.FeedbackContent); err != nil {
				return nil, err
			}
			// 反馈内容同步进时间轴，工作台详情可实时查看
			s.syncFeedbackToTimeline(note, userID, req.FeedbackContent)
			_ = s.recordLedger(id, userID, "complete", "本人任务完成", req.FeedbackContent, "")
			// 通知任务发起人
			if s.notifSvc != nil {
				content := req.FeedbackContent
				if content == "" {
					content = "已完成本人负责的部分"
				}
				_ = s.notifSvc.Notify(note.CreatorID.String(), userID, &note.ID, "task_completed", "任务已完成", content)
			}
			return s.noteRepo.FindByID(id)
		}
		// 发起者（或管理员）归档：需所有被指派人已完成
		if note.CreatorID.String() == userID || role == "super_admin" || role == "dept_admin" {
			if !allAssigneesCompleted(note) {
				return nil, apperrors.ErrAssigneesIncomplete
			}
			now := time.Now()
			note.ColorStatus = "green"
			note.IsArchived = true
			note.ArchiveTime = &now
			note.CompletedAt = &now
			if err := s.noteRepo.Update(note); err != nil {
				return nil, err
			}
			// 积分：发起者归档办结（完成/按时/提前/交办闭环给交办人）
			s.awardCompletionPoints(note, userID)
			_ = s.recordLedger(id, userID, "complete", "任务办结归档", "", "")
			return s.noteRepo.FindByID(id)
		}
		return nil, apperrors.ErrPermissionDenied
	}

	// 非指派任务 / 无实际被指派人的任务：完成即归档（原逻辑）
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
		// 反馈内容同步进时间轴
		s.syncFeedbackToTimeline(note, userID, req.FeedbackContent)
	}

	// 积分：任务办结归档（完成/按时/提前/交办闭环给交办人）
	s.awardCompletionPoints(note, userID)

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

// awardCompletionPoints 归档办结后的积分发放：完成 +3、按时 +2（≤due_time）、
// 提前 24h +1（≤due_time-24h）、交办闭环 +2（交办任务且交办人≠完成人）。
func (s *NoteService) awardCompletionPoints(note *models.Note, completerID string) {
	noteID := note.ID.String()
	title := truncateNote(note.Title)
	AwardPoints(completerID, "task_complete", "note", noteID, "完成任务《"+title+"》")
	if note.DueTime != nil && note.CompletedAt != nil && !note.CompletedAt.After(*note.DueTime) {
		AwardPoints(completerID, "task_complete_ontime", "note", noteID, "按时完成任务《"+title+"》")
		if !note.CompletedAt.After(note.DueTime.Add(-24 * time.Hour)) {
			AwardPoints(completerID, "task_complete_early", "note", noteID, "提前24小时完成任务《"+title+"》")
		}
	}
	if note.SourceType == "assigned" && note.CreatorID.String() != completerID {
		AwardPoints(note.CreatorID.String(), "task_assigner_closed", "note", noteID, "交办任务《"+title+"》已办结")
	}
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

	// 进行中的指派任务：被指派人提交反馈即视为本人完成
	if !note.IsArchived && note.SourceType == "assigned" && isAssignee && note.CreatorID.String() != userID {
		_ = s.noteRepo.UpdateAssigneeComplete(note.ID.String(), userID, content)
	}

	// 反馈内容同步进时间轴
	s.syncFeedbackToTimeline(note, userID, content)

	_ = s.recordLedger(id, userID, "feedback", "任务反馈填报", content, "")

	// 积分：有效反馈（≥10 字）
	if len([]rune(content)) >= 10 {
		AwardPoints(userID, "feedback_submit", "note", note.ID.String(), "反馈《"+truncateNote(note.Title)+"》")
	}

	if s.notifSvc != nil && note.CreatorID.String() != userID {
		_ = s.notifSvc.Notify(note.CreatorID.String(), userID, &note.ID, "task_feedback", "任务反馈", content)
	}

	return s.noteRepo.FindByID(id)
}

// Sign 被指派人签收任务：将本人对应 note_assignees 行的 sign_status 置为 signed。
// 仅指派任务（source_type=assigned）且当前用户为该任务被指派人（非发起者 initiator）可签收。
func (s *NoteService) Sign(id, userID string) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil || note == nil {
		return nil, apperrors.ErrNoteNotFound
	}

	if note.SourceType != "assigned" {
		return nil, apperrors.ErrPermissionDenied
	}

	var mine *models.NoteAssignee
	for i := range note.Assignees {
		if note.Assignees[i].UserID.String() == userID {
			mine = &note.Assignees[i]
			break
		}
	}
	if mine == nil || mine.RoleInNote == "initiator" {
		return nil, apperrors.ErrPermissionDenied
	}
	if mine.SignStatus == "signed" {
		return note, nil
	}

	if err := s.noteRepo.UpdateAssigneeSign(id, userID); err != nil {
		return nil, err
	}
	_ = s.recordLedger(id, userID, "update", "任务签收", "", "")

	// 积分：签收交办任务
	AwardPoints(userID, "sign_reply", "note", note.ID.String(), "签收任务《"+truncateNote(note.Title)+"》")

	// 通知任务发起人：被指派人已签收
	if s.notifSvc != nil && note.CreatorID.String() != userID {
		_ = s.notifSvc.Notify(note.CreatorID.String(), userID, &note.ID, "task_signed", "任务已签收", "「"+note.Title+"」已被签收")
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
	// TotalCompleted 当前用户累计完成数（名下归档 + 被指派完成去重），用于荣耀时刻庆祝动画
	TotalCompleted int64                    `json:"total_completed"`
	Year           int                      `json:"year"`
	Daily          []repository.NoteDayStat `json:"daily"`
}

func (s *NoteService) GetHeatmap(userID string, year int) (*NoteHeatmap, error) {
	total, err := s.noteRepo.CountArchivedByUser(userID)
	if err != nil {
		return nil, err
	}
	completed, err := s.noteRepo.CountCompletedByUser(userID)
	if err != nil {
		return nil, err
	}
	daily, err := s.noteRepo.HeatmapByYear(userID, year)
	if err != nil {
		return nil, err
	}
	return &NoteHeatmap{
		TotalArchived:  total,
		TotalCompleted: completed,
		Year:           year,
		Daily:          daily,
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
