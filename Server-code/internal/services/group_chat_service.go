package services

import (
	"encoding/json"
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
	if err := s.groupRepo.DissolveGroup(groupID); err != nil {
		return err
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
	msg := &models.GroupMessage{
		GroupID:   groupUUID,
		SenderID:  senderUUID,
		Type:      msgType,
		Content:   p.Content,
		FileName:  p.FileName,
		FilePath:  p.FilePath,
		FileSize:  p.FileSize,
		MimeType:  p.MimeType,
		CreatedAt: time.Now(),
	}
	if err := s.groupRepo.CreateMessage(msg); err != nil {
		return nil, err
	}

	// 推送给群内除发送者外的所有成员
	if s.hub != nil {
		memberIDs, _ := s.groupRepo.MemberIDs(groupID)
		// 附带发送者姓名，前端弹窗/通知直接展示
		senderName := ""
		if u, uerr := s.userRepo.FindByID(userID); uerr == nil && u != nil {
			if u.Name != "" {
				senderName = u.Name
			} else {
				senderName = u.Username
			}
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
	return s.groupRepo.ListMessages(groupID, page, pageSize)
}

func (s *GroupChatService) MarkGroupRead(userID, groupID string) error {
	if _, err := s.ensureMember(groupID, userID); err != nil {
		return err
	}
	return s.groupRepo.MarkGroupRead(groupID, userID)
}
