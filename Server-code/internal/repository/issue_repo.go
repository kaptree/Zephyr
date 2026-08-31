package repository

import (
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IssueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

type IssueFilter struct {
	Status   string // open / closed / ""
	Type     string // bug / feature / ""
	Keyword  string
	Page     int
	PageSize int
}

func (r *IssueRepository) List(f IssueFilter) ([]models.Issue, int64, error) {
	var issues []models.Issue
	var total int64

	if f.Keyword != "" && utils.IsPinyinKeyword(f.Keyword) {
		// 需求36：拼音搜索——数据库无法对中文做拼音匹配，改为全量加载后内存过滤
		var all []models.Issue
		query := r.db.Model(&models.Issue{}).Preload("User")
		if f.Status != "" {
			query = query.Where("status = ?", f.Status)
		}
		if f.Type != "" {
			query = query.Where("type = ?", f.Type)
		}
		if err := query.Order("created_at DESC").Find(&all).Error; err != nil {
			return nil, 0, err
		}
		for _, it := range all {
			if utils.MatchKeyword(f.Keyword, it.Title, it.Content) {
				issues = append(issues, it)
			}
		}
		total = int64(len(issues))
		issues = pageSlice(issues, f.Page, f.PageSize)
	} else {
		query := r.db.Model(&models.Issue{}).Preload("User")
		if f.Status != "" {
			query = query.Where("status = ?", f.Status)
		}
		if f.Type != "" {
			query = query.Where("type = ?", f.Type)
		}
		if f.Keyword != "" {
			kw := "%" + f.Keyword + "%"
			query = query.Where("title ILIKE ? OR content ILIKE ?", kw, kw)
		}

		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (f.Page - 1) * f.PageSize
		if err := query.Order("created_at DESC").
			Offset(offset).Limit(f.PageSize).Find(&issues).Error; err != nil {
			return nil, 0, err
		}
	}

	// 批量统计评论数
	if len(issues) > 0 {
		var ids []uuid.UUID
		for _, it := range issues {
			ids = append(ids, it.ID)
		}
		type cntRow struct {
			IssueID uuid.UUID
			Count   int64
		}
		var rows []cntRow
		r.db.Table("issue_comments").
			Select("issue_id, COUNT(*) AS count").
			Where("issue_id IN ?", ids).
			Group("issue_id").Scan(&rows)
		cntMap := make(map[string]int64, len(rows))
		for _, row := range rows {
			cntMap[row.IssueID.String()] = row.Count
		}
		for i := range issues {
			issues[i].CommentCount = cntMap[issues[i].ID.String()]
		}
	}

	return issues, total, nil
}

func (r *IssueRepository) FindByID(id string) (*models.Issue, error) {
	var issue models.Issue
	if err := r.db.Preload("User").First(&issue, "id = ?", id).Error; err != nil {
		return nil, err
	}
	var count int64
	r.db.Table("issue_comments").Where("issue_id = ?", issue.ID).Count(&count)
	issue.CommentCount = count
	return &issue, nil
}

// NextIssueNo 生成下一个自增编号
func (r *IssueRepository) NextIssueNo() (int, error) {
	var max int
	err := r.db.Model(&models.Issue{}).Select("COALESCE(MAX(issue_no), 0)").Scan(&max).Error
	if err != nil {
		return 1, err
	}
	return max + 1, nil
}

func (r *IssueRepository) Create(issue *models.Issue) error {
	return r.db.Create(issue).Error
}

// UpdateStatus 更新问题状态（open/closed）
func (r *IssueRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.Issue{}).Where("id = ?", id).Update("status", status).Error
}

func (r *IssueRepository) ListComments(issueID string) ([]models.IssueComment, error) {
	var comments []models.IssueComment
	err := r.db.Preload("User").
		Where("issue_id = ?", issueID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *IssueRepository) CreateComment(comment *models.IssueComment) error {
	return r.db.Create(comment).Error
}

// ---------------- 需求26：Issue 订阅 ----------------

// Subscribe 订阅 issue（幂等：已订阅则不重复插入）
func (r *IssueRepository) Subscribe(issueID, userID string) error {
	iid, _ := uuid.Parse(issueID)
	uid, _ := uuid.Parse(userID)
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.IssueSubscriber{
		IssueID: iid,
		UserID:  uid,
	}).Error
}

func (r *IssueRepository) Unsubscribe(issueID, userID string) error {
	iid, _ := uuid.Parse(issueID)
	uid, _ := uuid.Parse(userID)
	return r.db.Where("issue_id = ? AND user_id = ?", iid, uid).
		Delete(&models.IssueSubscriber{}).Error
}

func (r *IssueRepository) IsSubscribed(issueID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.IssueSubscriber{}).
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *IssueRepository) CountSubscribers(issueID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.IssueSubscriber{}).
		Where("issue_id = ?", issueID).
		Count(&count).Error
	return count, err
}

// ListSubscriberIDs 返回 issue 的全部订阅人 id（排除 excludeUserID）
func (r *IssueRepository) ListSubscriberIDs(issueID, excludeUserID string) ([]string, error) {
	var ids []string
	query := r.db.Model(&models.IssueSubscriber{}).Where("issue_id = ?", issueID)
	if excludeUserID != "" {
		query = query.Where("user_id <> ?", excludeUserID)
	}
	err := query.Pluck("user_id", &ids).Error
	return ids, err
}

// ---------------- 需求28：全局订阅（所有新 issue 通知） ----------------

// WatchAll 订阅全部 issue（幂等）
func (r *IssueRepository) WatchAll(userID string) error {
	uid, _ := uuid.Parse(userID)
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.IssueWatcher{
		UserID: uid,
	}).Error
}

func (r *IssueRepository) UnwatchAll(userID string) error {
	uid, _ := uuid.Parse(userID)
	return r.db.Where("user_id = ?", uid).Delete(&models.IssueWatcher{}).Error
}

func (r *IssueRepository) IsWatchingAll(userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.IssueWatcher{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count > 0, err
}

// ListWatcherIDs 返回全部全局订阅人 id（排除 excludeUserID）
func (r *IssueRepository) ListWatcherIDs(excludeUserID string) ([]string, error) {
	var ids []string
	query := r.db.Model(&models.IssueWatcher{})
	if excludeUserID != "" {
		query = query.Where("user_id <> ?", excludeUserID)
	}
	err := query.Pluck("user_id", &ids).Error
	return ids, err
}
