package models

import (
	"time"

	"github.com/google/uuid"
)

type AIConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProviderName string    `gorm:"type:varchar(100);not null" json:"provider_name"`
	APIEndpoint  string    `gorm:"type:varchar(500);not null" json:"api_endpoint"`
	APIKey       string    `gorm:"type:varchar(500);not null" json:"-"`
	APIKeyMasked string    `gorm:"-" json:"api_key_masked"`
	ModelName    string    `gorm:"type:varchar(100)" json:"model_name"`
	Description  string    `gorm:"type:varchar(500)" json:"description"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (AIConfig) TableName() string {
	return "ai_configs"
}

type ConfigFileHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FileName       string    `gorm:"type:varchar(255);not null;index" json:"file_name"`
	FilePath       string    `gorm:"type:varchar(500);not null" json:"file_path"`
	ContentBefore  string    `gorm:"type:text" json:"content_before"`
	ContentAfter   string    `gorm:"type:text" json:"content_after"`
	ChangedBy      string    `gorm:"type:varchar(100);not null" json:"changed_by"`
	ChangedByID    string    `gorm:"type:varchar(50);not null" json:"changed_by_id"`
	ChangeSummary  string    `gorm:"type:varchar(500)" json:"change_summary"`
	CreatedAt      time.Time `json:"created_at"`
}

func (ConfigFileHistory) TableName() string {
	return "config_file_history"
}

type AdminLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID    string    `gorm:"type:varchar(50);not null;index" json:"admin_id"`
	AdminName  string    `gorm:"type:varchar(100);not null" json:"admin_name"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	Resource   string    `gorm:"type:varchar(100);not null" json:"resource"`
	ResourceID string    `gorm:"type:varchar(100)" json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IPAddress  string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent  string    `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AdminLog) TableName() string {
	return "admin_logs"
}

type OperationLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:varchar(50);not null;index" json:"user_id"`
	UserName   string    `gorm:"type:varchar(100);not null" json:"user_name"`
	Role       string    `gorm:"type:varchar(30)" json:"role"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	Method     string    `gorm:"type:varchar(10)" json:"method"`
	Path       string    `gorm:"type:varchar(200)" json:"path"`
	Resource   string    `gorm:"type:varchar(50)" json:"resource"`
	ResourceID string    `gorm:"type:varchar(100)" json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	StatusCode int       `gorm:"default:200" json:"status_code"`
	IPAddress  string    `gorm:"type:varchar(50)" json:"ip_address"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
