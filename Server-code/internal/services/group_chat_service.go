package services

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/ws"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
)

type GroupChatService struct {
	groupRepo *repository.GroupChatRepo
	userRepo  *repository.UserRepository
	hub       *ws.Hub
}

func NewGroupChatService(groupRepo *repository.GroupChatRepo, userRepo *repository.UserRepository, hub *ws.Hub) *GroupChatService {
	return &GroupChatService{groupRepo: groupRepo, userRepo: userRepo, hub: hub}
}

// pushGroupEvent 向群内指定成员推送群变动事件（chat:group_updated）
func (s *GroupChatService) pushGroupEvent(groupID string, userIDs []string, action string, extra map[string]interface{}) {
	if s.hub == nil {
		return
	}
	payload := map[string]interface{}{
		"event":    "chat:group_updated",
		"group_id": groupID,
		"action":   action,
	}
	for k, v := range extra {
		payload[k] = v
	}
	data, _ := json.Marshal(payload)
	for _, uid := range userIDs {
		s.hub.PushToUser(uid, data)
	}
}

// CreateGroup 创建群聊（创建者为群主，可同时添加初始成员）
func (s *GroupChatService) CreateGroup(userID, name string, memberIDs []string) (*models.ChatGroup, error) {
	if name == "" {
		return nil, apperrors.ErrInvalidChatContent
	}
	ownerUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	ownerMember := models.ChatGroupMember{
		UserID: ownerUUID,
		Role:   "owner",
	}
	members := []models.ChatGroupMember{ownerMember}
	seen := map[string]bool{userID: true}
	for _, id := range memberIDs {
		if id == "" || seen[id] {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			continue
		}
		seen[id] = true
		members = append(members, models.ChatGroupMember{UserID: uuid.MustParse(id), Role: "member"})
	}

	group := &models.ChatGroup{
		Name:      name,
		OwnerID:   ownerUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.groupRepo.CreateGroup(group, members); err != nil {
		return nil, err
	}

	// 通知除创建者外的成员
	notifyIDs := make([]string, 0, len(members)-1)
	for _, m := range members {
		if m.UserID.String() != userID {
			notifyIDs = append(notifyIDs, m.UserID.String())
		}
	}
	s.pushGroupEvent(group.ID.String(), notifyIDs, "created", map[string]interface{}{"name": group.Name})
	return group, nil
}

func (s *GroupChatService) MyGroups(userID string) ([]map[string]interface{}, error) {
	return s.groupRepo.ListUserGroups(userID)
}

// ensureMember 校验当前用户是群成员（返回成员记录；群不存在返回 ErrGroupNotFound，非成员返回 ErrPermissionDenied）
func (s *GroupChatService) ensureMember(groupID, userID string) (*models.ChatGroupMember, error) {
	if _, err := uuid.Parse(groupID); err != nil {
		return nil, apperrors.ErrGroupNotFound
	}
	m, err := s.groupRepo.GetMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		// 区分「群不存在」与「非群成员」
		if _, gerr := s.groupRepo.GetGroup(groupID); gerr != nil {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, apperrors.ErrPermissionDenied
	}
	return m, nil
}

func (s *GroupChatService) GroupDetail(userID, groupID string) (*models.ChatGroup, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	return s.groupRepo.GetGroup(groupID)
}

func (s *GroupChatService) Members(userID, groupID string) ([]models.ChatGroupMember, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	return s.groupRepo.ListMembers(groupID)
}

// requireOwner 校验当前用户是群主
func (s *GroupChatService) requireOwner(groupID, userID string) error {
	m, err := s.ensureMember(groupID, userID)
	if err != nil {
		return err
	}
	if m.Role != "owner" {
		return apperrors.ErrPermissionDenied
	}
	return nil
}

// AddMembers 群主添加成员（幂等）
func (s *GroupChatService) AddMembers(userID, groupID string, userIDs []string) (int64, error) {
	if err := s.requireOwner(groupID, userID); err != nil {
		return 0, err
	}
	ids := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		if id == "" || id == userID {
			continue
		}
		if u, err := uuid.Parse(id); err == nil {
			ids = append(ids, u)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	added, err := s.groupRepo.AddMembers(groupID, ids, "member")
	if err != nil {
		return 0, err
	}
	if added > 0 {
		notifyIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			notifyIDs = append(notifyIDs, id.String())
		}
		s.pushGroupEvent(groupID, notifyIDs, "member_added", nil)
	}
	return added, nil
}

// RemoveMember 群主移除成员（不能移除自己 / 其他群主）
func (s *GroupChatService) RemoveMember(userID, groupID, targetID string) error {
	if err := s.requireOwner(groupID, userID); err != nil {
		return err
	}
	if targetID == userID {
		return apperrors.ErrInvalidOperation
	}
	target, err := s.groupRepo.GetMember(groupID, targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return apperrors.ErrGroupNotFound
	}
	if err := s.groupRepo.RemoveMember(groupID, targetID); err != nil {
		return err
	}
	s.pushGroupEvent(groupID, []string{targetID}, "member_removed", nil)
	return nil
}

// RenameGroup 群主重命名群聊
func (s *GroupChatService) RenameGroup(userID, groupID, name string) error {
	if name == "" {
		return apperrors.ErrInvalidChatContent
	}
	if err := s.requireOwner(groupID, userID); err != nil {
		return err
	}
	g, err := s.groupRepo.GetGroup(groupID)
	if err != nil {
		return err
	}
	g.Name = name
	g.UpdatedAt = time.Now()
	if err := s.groupRepo.UpdateGroup(g); err != nil {
		return err
	}
	memberIDs, _ := s.groupRepo.MemberIDs(groupID)
	for _, id := range memberIDs {
		if id != userID {
			s.pushGroupEvent(groupID, []string{id}, "renamed", map[string]interface{}{"name": name})
		}
	}
	return nil
}

// Dissolve 群主解散群聊（删除群 + 成员 + 消息）
func (s *GroupChatService) Dissolve(userID, groupID string) error {
	if err := s.requireOwner(groupID, userID); err != nil {
		return err
	}
	g, err := s.groupRepo.GetGroup(groupID)
	if err != nil {
		return err
	}
	memberIDs, _ := s.groupRepo.MemberIDs(groupID)
	files, _ := s.groupRepo.ListGroupFiles(groupID)
	if err := s.groupRepo.DissolveGroup(groupID); err != nil {
		return err
	}
	// 清理群文件磁盘副本（软失败不影响业务）
	for _, f := range files {
		os.Remove(strings.TrimPrefix(f.FilePath, "/"))
	}
	notifyIDs := make([]string, 0, len(memberIDs))
	for _, id := range memberIDs {
		if id != userID {
			notifyIDs = append(notifyIDs, id)
		}
	}
	s.pushGroupEvent(groupID, notifyIDs, "dissolved", map[string]interface{}{"name": g.Name})
	return nil
}

// Quit 成员退出群聊（群主不可退出，需解散）
func (s *GroupChatService) Quit(userID, groupID string) error {
	m, err := s.ensureMember(groupID, userID)
	if err != nil {
		return err
	}
	if m.Role == "owner" {
		return apperrors.ErrInvalidOperation
	}
	if err := s.groupRepo.RemoveMember(groupID, userID); err != nil {
		return err
	}
	return nil
}

// GroupMessagePayload 群消息载荷（支持文本 / 图片 / 文件）
type GroupMessagePayload struct {
	Type     string
	Content  string
	FileName string
	FilePath string
	FileSize int64
	MimeType string
	Mentions []string
}

// FilterMentionIDs 过滤被 @ 的成员 ID：去重、保持出现顺序、仅保留真实群成员。
// 纯函数，便于单元测试。
func FilterMentionIDs(memberIDs []string, mentions []string) []string {
	if len(mentions) == 0 {
		return nil
	}
	valid := make(map[string]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		valid[id] = struct{}{}
	}
	seen := make(map[string]struct{}, len(mentions))
	out := make([]string, 0, len(mentions))
	for _, id := range mentions {
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		if _, ok := valid[id]; !ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SendGroupMessage 发送群消息，并实时推送给群内其他成员
func (s *GroupChatService) SendGroupMessage(userID, groupID string, p GroupMessagePayload) (*models.GroupMessage, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	msgType := p.Type
	if msgType == "" {
		msgType = "text"
	}
	switch msgType {
	case "text":
		if p.Content == "" {
			return nil, apperrors.ErrInvalidChatContent
		}
	case "image", "file":
		if p.FilePath == "" {
			return nil, apperrors.ErrInvalidChatContent
		}
	default:
		return nil, apperrors.ErrInvalidChatContent
	}

	senderUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, err
	}

	// 群成员列表：用于 @ 提及过滤与 WS 推送
	memberIDs, _ := s.groupRepo.MemberIDs(groupID)

	// @ 提及：去重并仅保留真实群成员，序列化为 JSON 数组存储
	mentionIDs := FilterMentionIDs(memberIDs, p.Mentions)
	mentionsJSON := ""
	if len(mentionIDs) > 0 {
		if b, merr := json.Marshal(mentionIDs); merr == nil {
			mentionsJSON = string(b)
		}
	}

	msg := &models.GroupMessage{
		GroupID:   groupUUID,
		SenderID:  senderUUID,
		Type:      msgType,
		Content:   p.Content,
		Mentions:  mentionsJSON,
		FileName:  p.FileName,
		FilePath:  p.FilePath,
		FileSize:  p.FileSize,
		MimeType:  p.MimeType,
		CreatedAt: time.Now(),
	}
	if err := s.groupRepo.CreateMessage(msg); err != nil {
		return nil, err
	}
	// HTTP 响应输出解析后的提及列表（保持数组类型，空为 []）
	if mentionIDs == nil {
		mentionIDs = []string{}
	}
	msg.MentionList = mentionIDs

	// 推送给群内除发送者外的所有成员
	if s.hub != nil {
		// 附带发送者姓名，前端弹窗/通知直接展示
		senderName := ""
		if u, uerr := s.userRepo.FindByID(userID); uerr == nil && u != nil {
			if u.Name != "" {
				senderName = u.Name
			} else {
				senderName = u.Username
			}
		}
		if mentionIDs == nil {
			mentionIDs = []string{}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"event": "chat:group_message",
			"message": map[string]interface{}{
				"id":          msg.ID,
				"group_id":    msg.GroupID,
				"sender_id":   msg.SenderID,
				"sender_name": senderName,
				"type":        msg.Type,
				"content":     msg.Content,
				"mentions":    mentionIDs,
				"file_name":   msg.FileName,
				"file_path":   msg.FilePath,
				"file_size":   msg.FileSize,
				"mime_type":   msg.MimeType,
				"created_at":  msg.CreatedAt,
			},
		})
		for _, id := range memberIDs {
			if id != userID {
				s.hub.PushToUser(id, payload)
			}
		}
	}
	return msg, nil
}

func (s *GroupChatService) ListMessages(userID, groupID string, page, pageSize int) ([]models.GroupMessage, int64, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, 0, err
	}
	msgs, total, err := s.groupRepo.ListMessages(groupID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	// 输出解析后的提及列表（mentions 字段为数组）
	for i := range msgs {
		msgs[i].MentionList = msgs[i].DecodeMentions()
	}
	return msgs, total, nil
}

func (s *GroupChatService) MarkGroupRead(userID, groupID string) error {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return err
	}
	return s.groupRepo.MarkGroupRead(groupID, userID)
}

// GetAnnouncement 获取群公告（成员可查看）
func (s *GroupChatService) GetAnnouncement(userID, groupID string) (map[string]interface{}, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	g, err := s.groupRepo.GetGroup(groupID)
	if err != nil {
		return nil, apperrors.ErrGroupNotFound
	}
	return map[string]interface{}{
		"announcement":            g.Announcement,
		"announcement_updated_at": g.AnnouncementUpdatedAt,
	}, nil
}

// SetAnnouncement 群主设置/清除群公告（content 为空表示清除），并实时推送给其他成员
func (s *GroupChatService) SetAnnouncement(userID, groupID, content string) error {
	if len([]rune(content)) > 500 {
		return apperrors.ErrInvalidChatContent
	}
	if err := s.requireOwner(groupID, userID); err != nil {
		return err
	}
	if err := s.groupRepo.UpdateAnnouncement(groupID, content); err != nil {
		return err
	}
	memberIDs, _ := s.groupRepo.MemberIDs(groupID)
	var updatedAt interface{}
	if content != "" {
		updatedAt = time.Now()
	}
	for _, id := range memberIDs {
		if id == userID {
			continue
		}
		s.pushGroupEvent(groupID, []string{id}, "announcement_updated", map[string]interface{}{
			"announcement":            content,
			"announcement_updated_at": updatedAt,
		})
	}
	return nil
}

// GroupFileMeta 群文件上传元数据（由 handler 落盘后传入）
type GroupFileMeta struct {
	FileName string
	FilePath string
	FileSize int64
	MimeType string
}

// groupFilePayload 群文件事件载荷
func groupFilePayload(f *models.GroupFile, uploaderName string) map[string]interface{} {
	return map[string]interface{}{
		"id":             f.ID,
		"group_id":       f.GroupID,
		"uploader_id":    f.UploaderID,
		"uploader_name":  uploaderName,
		"file_name":      f.FileName,
		"file_path":      f.FilePath,
		"file_size":      f.FileSize,
		"mime_type":      f.MimeType,
		"download_count": f.DownloadCount,
		"created_at":     f.CreatedAt,
	}
}

// UploadGroupFile 成员上传群文件，并实时推送给其他成员
func (s *GroupChatService) UploadGroupFile(userID, groupID string, meta GroupFileMeta) (*models.GroupFile, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	if meta.FileName == "" || meta.FilePath == "" {
		return nil, apperrors.ErrInvalidChatContent
	}
	uploaderUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, err
	}
	f := &models.GroupFile{
		GroupID:    groupUUID,
		UploaderID: uploaderUUID,
		FileName:   meta.FileName,
		FilePath:   meta.FilePath,
		FileSize:   meta.FileSize,
		MimeType:   meta.MimeType,
		CreatedAt:  time.Now(),
	}
	if err := s.groupRepo.CreateGroupFile(f); err != nil {
		return nil, err
	}

	uploaderName := ""
	if u, uerr := s.userRepo.FindByID(userID); uerr == nil && u != nil {
		f.Uploader = u
		if u.Name != "" {
			uploaderName = u.Name
		} else {
			uploaderName = u.Username
		}
	}
	if s.hub != nil {
		memberIDs, _ := s.groupRepo.MemberIDs(groupID)
		for _, id := range memberIDs {
			if id == userID {
				continue
			}
			s.pushGroupEvent(groupID, []string{id}, "file_added", map[string]interface{}{
				"file": groupFilePayload(f, uploaderName),
			})
		}
	}
	return f, nil
}

// ListGroupFiles 群文件列表（成员可查看）
func (s *GroupChatService) ListGroupFiles(userID, groupID string) ([]models.GroupFile, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	return s.groupRepo.ListGroupFiles(groupID)
}

// DownloadGroupFile 校验成员身份并记录下载（计数 +1，推送其他成员刷新计数）
func (s *GroupChatService) DownloadGroupFile(userID, groupID, fileID string) (*models.GroupFile, error) {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return nil, err
	}
	f, err := s.groupRepo.GetGroupFile(groupID, fileID)
	if err != nil {
		return nil, apperrors.ErrGroupFileNotFound
	}
	count, err := s.groupRepo.IncrDownloadCount(groupID, fileID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		f.DownloadCount = int(count)
	}
	if s.hub != nil {
		memberIDs, _ := s.groupRepo.MemberIDs(groupID)
		for _, id := range memberIDs {
			if id == userID {
				continue
			}
			s.pushGroupEvent(groupID, []string{id}, "file_downloaded", map[string]interface{}{
				"file_id":        fileID,
				"download_count": f.DownloadCount,
			})
		}
	}
	return f, nil
}

// DeleteGroupFile 删除群文件（上传者或群主），并实时推送给其他成员
func (s *GroupChatService) DeleteGroupFile(userID, groupID, fileID string) error {
	m, err := s.ensureMember(groupID, userID)
	if err != nil {
		return err
	}
	f, err := s.groupRepo.GetGroupFile(groupID, fileID)
	if err != nil {
		return apperrors.ErrGroupFileNotFound
	}
	if m.Role != "owner" && f.UploaderID.String() != userID {
		return apperrors.ErrPermissionDenied
	}
	if err := s.groupRepo.DeleteGroupFile(groupID, fileID); err != nil {
		return err
	}
	// 清理磁盘文件（软失败不影响业务）
	os.Remove(strings.TrimPrefix(f.FilePath, "/"))
	if s.hub != nil {
		memberIDs, _ := s.groupRepo.MemberIDs(groupID)
		for _, id := range memberIDs {
			if id == userID {
				continue
			}
			s.pushGroupEvent(groupID, []string{id}, "file_removed", map[string]interface{}{
				"file_id": fileID,
			})
		}
	}
	return nil
}
