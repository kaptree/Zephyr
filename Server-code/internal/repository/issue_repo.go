package repository

import (
	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
