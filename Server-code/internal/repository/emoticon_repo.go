package repository

import (
	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmoticonRepository struct {
	db *gorm.DB
}

func NewEmoticonRepository(db *gorm.DB) *EmoticonRepository {
	return &EmoticonRepository{db: db}
}

// List 表情列表：系统内置全部 + 指定用户上传的（他人上传的非系统表情不可见）
// category 为空时返回全部；非空时仅返回该分类
func (r *EmoticonRepository) List(userID *uuid.UUID, category string) ([]models.Emoticon, error) {
	var list []models.Emoticon
	query := r.db.Model(&models.Emoticon{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if userID != nil {
		query = query.Where("is_system = ? OR uploader_id = ?", true, *userID)
	} else {
		query = query.Where("is_system = ?", true)
	}
	if err := query.Order("created_at ASC, name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListCategories 返回可见的分类名列表（按创建时间排序）
func (r *EmoticonRepository) ListCategories(userID *uuid.UUID) ([]string, error) {
	var cats []string
	query := r.db.Model(&models.Emoticon{}).
		Select("DISTINCT category").
		Order("category ASC")
	if userID != nil {
		query = query.Where("is_system = ? OR uploader_id = ?", true, *userID)
	} else {
		query = query.Where("is_system = ?", true)
	}
	if err := query.Pluck("category", &cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *EmoticonRepository) Create(e *models.Emoticon) error {
	return r.db.Create(e).Error
}

func (r *EmoticonRepository) FindByID(id uuid.UUID) (*models.Emoticon, error) {
	var e models.Emoticon
	if err := r.db.First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmoticonRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Emoticon{}, "id = ?", id).Error
}

// ExistsByPath 判断某路径是否已存在记录（用于系统表情幂等导入）
func (r *EmoticonRepository) ExistsByPath(path string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Emoticon{}).Where("path = ?", path).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
