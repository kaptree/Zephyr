package repository

import (
	"labelpro-server/internal/models"

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
