package models

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string     `gorm:"type:varchar(50);not null" json:"name"`
	SubTag     string     `gorm:"type:varchar(100)" json:"sub_tag"`
	Color      string     `gorm:"type:varchar(7);default:'#3B82F6'" json:"color"`
	Category   string     `gorm:"type:varchar(30)" json:"category"`
	Scope      string     `gorm:"type:varchar(20);default:'personal'" json:"scope"`
	CreatorID  *uuid.UUID `gorm:"type:uuid" json:"creator_id"`
	Creator    *User      `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	SortOrder  int        `gorm:"default:0" json:"sort_order"`
	UsageCount int64      `gorm:"-" json:"usage_count"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type Template struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Type      string     `gorm:"type:varchar(30);default:'default'" json:"type"`
	Fields    string     `gorm:"type:jsonb;default:'[]'" json:"fields"`
	Layout    string     `gorm:"type:varchar(10);default:'1'" json:"layout"`
	IsSystem  bool       `gorm:"default:false" json:"is_system"`
	CreatorID *uuid.UUID `gorm:"type:uuid" json:"creator_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Template) TableName() string {
	return "templates"
}

type WorkGroup struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	NoteID       *uuid.UUID `gorm:"type:uuid" json:"note_id"`
	InitiatorID  uuid.UUID  `gorm:"type:uuid;not null" json:"initiator_id"`
	Initiator    *User      `gorm:"foreignKey:InitiatorID" json:"initiator,omitempty"`
	TemplateType string     `gorm:"type:varchar(30);default:'default'" json:"template_type"`
	Status       string     `gorm:"type:varchar(20);default:'active'" json:"status"`
	DueTime      *time.Time `json:"due_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Members []WorkGroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

func (WorkGroup) TableName() string {
	return "work_groups"
}

type WorkGroupMember struct {
	GroupID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role         string    `gorm:"type:varchar(20);default:'member'" json:"role"`
	SubGroupName string    `gorm:"type:varchar(100)" json:"sub_group_name"`
}

func (WorkGroupMember) TableName() string {
	return "work_group_members"
}

type CollaborationRoom struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	NoteID         uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"note_id"`
	CanvasData     string     `gorm:"type:jsonb;default:'{}'" json:"canvas_data"`
	Columns        int        `gorm:"default:1" json:"columns"`
	Version        int        `gorm:"default:0" json:"version"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (CollaborationRoom) TableName() string {
	return "collaboration_rooms"
}

type Reminder struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	NoteID         uuid.UUID `gorm:"type:uuid;not null;index" json:"note_id"`
	ReminderID     uuid.UUID `gorm:"type:uuid;not null" json:"reminder_id"`
	Reminder       *User     `gorm:"foreignKey:ReminderID" json:"reminder,omitempty"`
	TargetID       uuid.UUID `gorm:"type:uuid;not null" json:"target_id"`
	Target         *User     `gorm:"foreignKey:TargetID" json:"target,omitempty"`
	Message        string    `gorm:"type:text" json:"message"`
	RemindType     string    `gorm:"type:varchar(20);default:'normal'" json:"remind_type"`
	IsAcknowledged bool      `gorm:"default:false" json:"is_acknowledged"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Reminder) TableName() string {
	return "reminders"
}

type PresetGroup struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	TemplateType string     `gorm:"type:varchar(30);default:'default'" json:"template_type"`
	CreatorID    uuid.UUID  `gorm:"type:uuid;not null" json:"creator_id"`
	Creator      *User      `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Members      []PresetGroupMember `gorm:"foreignKey:PresetID" json:"members,omitempty"`
}

func (PresetGroup) TableName() string {
	return "preset_groups"
}

type PresetGroupMember struct {
	PresetID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"preset_id"`
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role         string    `gorm:"type:varchar(20);default:'member'" json:"role"`
	SubGroupName string    `gorm:"type:varchar(100)" json:"sub_group_name"`
}

func (PresetGroupMember) TableName() string {
	return "preset_group_members"
}

type LedgerEntry struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	NoteID       uuid.UUID `gorm:"type:uuid;index" json:"note_id"`
	UserID       uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action       string    `gorm:"type:varchar(30);not null" json:"action"`
	ActionDetail string    `gorm:"type:text" json:"action_detail"`
	IPAddress    string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent    string    `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}

func (LedgerEntry) TableName() string {
	return "ledger_entries"
}
