package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	DepartmentID *uuid.UUID     `gorm:"type:uuid;index" json:"dept_id"`
	Department   *Department    `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Role         string         `gorm:"type:varchar(50);not null;default:'member'" json:"role"`
	Rank         string         `gorm:"type:varchar(50)" json:"rank"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Email        string         `gorm:"type:varchar(100)" json:"email"`
	AvatarURL    string         `gorm:"type:varchar(500)" json:"avatar"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type Department struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string      `gorm:"type:varchar(200);not null" json:"name"`
	ParentID  *uuid.UUID  `gorm:"type:uuid;index" json:"parent_id"`
	Parent    *Department `gorm:"foreignKey:ParentID" json:"-"`
	Level     int         `gorm:"default:1" json:"level"`
	LeaderID  *uuid.UUID  `gorm:"type:uuid" json:"leader_id"`
	Leader    *User       `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	SortOrder int         `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (Department) TableName() string {
	return "departments"
}

type RolePermission struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Role     string    `gorm:"type:varchar(50);not null;index" json:"role"`
	Resource string    `gorm:"type:varchar(50);not null" json:"resource"`
	Action   string    `gorm:"type:varchar(50);not null" json:"action"`
	Scope    string    `gorm:"type:varchar(50);not null;default:'self'" json:"scope"`
}

func (RolePermission) TableName() string {
	return "roles_permissions"
}
