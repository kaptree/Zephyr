package repository

import (
	"time"

	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupChatRepo struct {
	db *gorm.DB
}

func NewGroupChatRepo(db *gorm.DB) *GroupChatRepo {
	return &GroupChatRepo{db: db}
}

// CreateGroup 事务创建群 + 初始成员
func (r *GroupChatRepo) CreateGroup(group *models.ChatGroup, members []models.ChatGroupMember) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		for i := range members {
			members[i].GroupID = group.ID
		}
		if len(members) > 0 {
			if err := tx.Create(&members).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GroupChatRepo) GetGroup(id string) (*models.ChatGroup, error) {
	var g models.ChatGroup
	if err := r.db.Preload("Owner").First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupChatRepo) UpdateGroup(group *models.ChatGroup) error {
	return r.db.Save(group).Error
}

// DissolveGroup 事务解散群：删除群 + 成员 + 消息
func (r *GroupChatRepo) DissolveGroup(groupID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&models.ChatGroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", groupID).Delete(&models.ChatGroup{}).Error
	})
}

// GetMember 查询某个成员记录（不存在返回 nil, nil）
func (r *GroupChatRepo) GetMember(groupID, userID string) (*models.ChatGroupMember, error) {
	var m models.ChatGroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ListMembers 成员列表（预加载用户信息）
func (r *GroupChatRepo) ListMembers(groupID string) ([]models.ChatGroupMember, error) {
	var members []models.ChatGroupMember
	err := r.db.Preload("User").Preload("User.Department").
		Where("group_id = ?", groupID).
		Order("CASE role WHEN 'owner' THEN 0 ELSE 1 END, joined_at ASC").
		Find(&members).Error
	return members, err
}

// MemberIDs 群成员用户 ID 列表
func (r *GroupChatRepo) MemberIDs(groupID string) ([]string, error) {
	var ids []string
	err := r.db.Model(&models.ChatGroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &ids).Error
	return ids, err
}

// AddMembers 批量添加成员（幂等：已在群内的忽略），返回实际新增数量
func (r *GroupChatRepo) AddMembers(groupID string, userIDs []uuid.UUID, role string) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	var existing []string
	if err := r.db.Model(&models.ChatGroupMember{}).
		Where("group_id = ? AND user_id IN ?", groupID, userIDs).
		Pluck("user_id", &existing).Error; err != nil {
		return 0, err
	}
	existSet := make(map[string]bool, len(existing))
	for _, id := range existing {
		existSet[id] = true
	}
	var toAdd []models.ChatGroupMember
	for _, id := range userIDs {
		if !existSet[id.String()] {
			toAdd = append(toAdd, models.ChatGroupMember{
				GroupID: uuid.MustParse(groupID),
				UserID:  id,
				Role:    role,
			})
		}
	}
	if len(toAdd) == 0 {
		return 0, nil
	}
	res := r.db.Create(&toAdd)
	return res.RowsAffected, res.Error
}

func (r *GroupChatRepo) RemoveMember(groupID, userID string) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.ChatGroupMember{}).Error
}

func (r *GroupChatRepo) CreateMessage(m *models.GroupMessage) error {
	return r.db.Create(m).Error
}

// ListMessages 分页拉取群消息（返回正序方便前端渲染）
func (r *GroupChatRepo) ListMessages(groupID string, page, pageSize int) ([]models.GroupMessage, int64, error) {
	var list []models.GroupMessage
	var total int64

	query := r.db.Model(&models.GroupMessage{}).Where("group_id = ?", groupID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("Sender").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, total, err
}

// MarkGroupRead 更新成员最后已读时间
func (r *GroupChatRepo) MarkGroupRead(groupID, userID string) error {
	now := time.Now()
	return r.db.Model(&models.ChatGroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("last_read_at", now).Error
}

// ListUserGroups 我所在的群列表：含成员数 / 最后一条消息 / 未读数，按最后消息时间倒序
func (r *GroupChatRepo) ListUserGroups(userID string) ([]map[string]interface{}, error) {
	var groups []models.ChatGroup
	if err := r.db.
		Joins("JOIN chat_group_members m ON m.group_id = chat_groups.id AND m.user_id = ?", userID).
		Order("chat_groups.created_at DESC").
		Find(&groups).Error; err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return []map[string]interface{}{}, nil
	}

	var myMemberRows []models.ChatGroupMember
	if err := r.db.Where("user_id = ?", userID).Find(&myMemberRows).Error; err != nil {
		return nil, err
	}
	lastReadMap := make(map[string]time.Time, len(myMemberRows))
	for _, m := range myMemberRows {
		if m.LastReadAt != nil {
			lastReadMap[m.GroupID.String()] = *m.LastReadAt
		}
	}

	results := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		gid := g.ID.String()
		var memberCount int64
		r.db.Model(&models.ChatGroupMember{}).Where("group_id = ?", g.ID).Count(&memberCount)

		var last models.GroupMessage
		hasLast := r.db.Where("group_id = ?", g.ID).
			Order("created_at DESC").First(&last).Error == nil

		var unread int64
		if cond := r.db.Model(&models.GroupMessage{}).
			Where("group_id = ? AND sender_id <> ?", g.ID, userID); cond != nil {
			if lr, ok := lastReadMap[gid]; ok {
				cond = cond.Where("created_at > ?", lr)
			}
			cond.Count(&unread)
		}

		item := map[string]interface{}{
			"id":           g.ID,
			"name":         g.Name,
			"owner_id":     g.OwnerID,
			"member_count": memberCount,
			"unread":       unread,
			"created_at":   g.CreatedAt,
		}
		if hasLast {
			item["last_msg"] = summarizeGroupLastMsg(&last)
			item["last_type"] = last.Type
			item["last_at"] = last.CreatedAt
		} else {
			item["last_msg"] = ""
			item["last_type"] = ""
			item["last_at"] = nil
		}
		results = append(results, item)
	}

	// 按最后消息时间倒序（无消息的群按创建时间排在后面）
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			t1, _ := results[i]["last_at"].(time.Time)
			t2, _ := results[j]["last_at"].(time.Time)
			if t2.After(t1) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}

// summarizeGroupLastMsg 群会话摘要：图片/文件消息展示占位文案
func summarizeGroupLastMsg(m *models.GroupMessage) string {
	switch m.Type {
	case "image":
		return "[图片]"
	case "file":
		if m.FileName != "" {
			return "[文件] " + m.FileName
		}
		return "[文件]"
	default:
		return m.Content
	}
}
