package repository

import (
	"labelpro-server/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func (r *SystemRepository) ListAIConfigs() ([]models.AIConfig, error) {
	var configs []models.AIConfig
	err := r.db.Order("created_at DESC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	for i := range configs {
		configs[i].APIKeyMasked = "••••••••"
	}
	return configs, nil
}

func (r *SystemRepository) GetAIConfig(id string) (*models.AIConfig, error) {
	var config models.AIConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	config.APIKeyMasked = "••••••••"
	return &config, nil
}

func (r *SystemRepository) CreateAIConfig(config *models.AIConfig) error {
	return r.db.Create(config).Error
}

func (r *SystemRepository) UpdateAIConfig(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.AIConfig{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteAIConfig(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&models.AIConfig{}, "id = ?", uid).Error
}

func (r *SystemRepository) SaveConfigFileHistory(history *models.ConfigFileHistory) error {
	return r.db.Create(history).Error
}

func (r *SystemRepository) GetConfigFileHistory(fileName string, limit int) ([]models.ConfigFileHistory, error) {
	var histories []models.ConfigFileHistory
	query := r.db.Where("file_name = ?", fileName).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&histories).Error
	return histories, err
}

func (r *SystemRepository) CreateAdminLog(log *models.AdminLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) ListAdminLogs(page, pageSize int) ([]models.AdminLog, int64, error) {
	var logs []models.AdminLog
	var total int64

	if err := r.db.Model(&models.AdminLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

type OperationLogFilter struct {
	UserID   string
	UserName string
	Action   string
	Method   string
	DateFrom string
	DateTo   string
	Page     int
	PageSize int
}

func (r *SystemRepository) CreateOperationLog(log *models.OperationLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) ListOperationLogs(f OperationLogFilter) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	query := r.db.Model(&models.OperationLog{})

	if f.UserID != "" {
		query = query.Where("user_id = ?", f.UserID)
	}
	if f.UserName != "" {
		query = query.Where("user_name ILIKE ?", "%"+f.UserName+"%")
	}
	if f.Action != "" {
		query = query.Where("action = ?", f.Action)
	}
	if f.Method != "" {
		query = query.Where("method = ?", f.Method)
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
	err := query.Order("created_at DESC").Offset(offset).Limit(f.PageSize).Find(&logs).Error
	return logs, total, err
}

func (r *SystemRepository) GetDistinctActions() ([]string, error) {
	var actions []string
	err := r.db.Model(&models.OperationLog{}).Distinct("action").Pluck("action", &actions).Error
	return actions, err
}

func (r *SystemRepository) CreateWorkReport(report *models.WorkReport) error {
	return r.db.Create(report).Error
}

type WorkReportFilter struct {
	UserID   string
	Period   string
	Keyword  string
	DateFrom string
	DateTo   string
	Page     int
	PageSize int
}

func (r *SystemRepository) ListWorkReports(f WorkReportFilter) ([]models.WorkReport, int64, error) {
	var reports []models.WorkReport
	var total int64

	query := r.db.Model(&models.WorkReport{})

	if f.UserID != "" {
		query = query.Where("user_id = ?", f.UserID)
	}
	if f.Period != "" {
		query = query.Where("period = ?", f.Period)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", kw, kw)
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
	err := query.Order("created_at DESC").Offset(offset).Limit(f.PageSize).Find(&reports).Error
	return reports, total, err
}

func (r *SystemRepository) GetWorkReport(id string) (*models.WorkReport, error) {
	var report models.WorkReport
	err := r.db.Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *SystemRepository) DeleteWorkReport(id string) error {
	return r.db.Delete(&models.WorkReport{}, "id = ?", id).Error
}

func (r *SystemRepository) ListGroupReports(groupID string, page, pageSize int) ([]models.WorkReport, int64, error) {
	var reports []models.WorkReport
	var total int64
	query := r.db.Model(&models.WorkReport{}).Where("group_id = ?", groupID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

func (r *SystemRepository) FindLatestGroupReport(groupID string) (*models.WorkReport, error) {
	var report models.WorkReport
	err := r.db.Where("group_id = ?", groupID).Order("created_at DESC").First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *SystemRepository) GetReportTemplate(id string) (*models.ReportTemplate, error) {
	var tpl models.ReportTemplate
	err := r.db.Where("id = ?", id).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (r *SystemRepository) SaveReportTemplate(tpl *models.ReportTemplate) error {
	var existing models.ReportTemplate
	err := r.db.Where("id = ?", tpl.ID).First(&existing).Error
	if err != nil {
		return r.db.Create(tpl).Error
	}
	tpl.UpdatedAt = time.Now()
	return r.db.Model(&models.ReportTemplate{}).Where("id = ?", tpl.ID).Updates(map[string]interface{}{
		"name":       tpl.Name,
		"content":    tpl.Content,
		"updated_at": tpl.UpdatedAt,
	}).Error
}
