package repository

import (
	"errors"

	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) FindAll() ([]models.Department, error) {
	var depts []models.Department
	err := r.db.Preload("Leader").Order("level ASC, sort_order ASC, name ASC").Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) FindByID(id string) (*models.Department, error) {
	var dept models.Department
	err := r.db.Preload("Leader").First(&dept, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepository) FindByParentID(parentID string) ([]models.Department, error) {
	var depts []models.Department
	err := r.db.Where("parent_id = ?", parentID).Order("sort_order ASC, name ASC").Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) CountMembers(deptID string) (int64, error) {
	var count int64
	uid, err := uuid.Parse(deptID)
	if err != nil {
		return 0, err
	}
	err = r.db.Model(&models.User{}).Where("department_id = ? AND is_active = true", uid).Count(&count).Error
	return count, err
}

func (r *DepartmentRepository) Create(dept *models.Department) error {
	return r.db.Create(dept).Error
}

func (r *DepartmentRepository) Update(dept *models.Department) error {
	return r.db.Save(dept).Error
}

func (r *DepartmentRepository) Delete(id string) error {
	return r.db.Delete(&models.Department{}, "id = ?", id).Error
}

func (r *DepartmentRepository) BuildTree() ([]DepartmentTreeNode, error) {
	allDepts, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	deptMap := make(map[string]*DepartmentTreeNode)
	var roots []DepartmentTreeNode

	for _, d := range allDepts {
		node := &DepartmentTreeNode{
			ID:       d.ID.String(),
			Name:     d.Name,
			ParentID: "",
			Level:    d.Level,
			Children: []DepartmentTreeNode{},
		}
		if d.Leader != nil {
			node.LeaderName = d.Leader.Name
		}
		if d.ParentID != nil {
			node.ParentID = d.ParentID.String()
		}
		count, _ := r.CountMembers(d.ID.String())
		node.MemberCount = count
		deptMap[d.ID.String()] = node
	}

	for _, d := range allDepts {
		node := deptMap[d.ID.String()]
		if d.ParentID != nil {
			parent, ok := deptMap[d.ParentID.String()]
			if ok {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}

	return roots, nil
}

func (r *DepartmentRepository) GetSubDeptIDs(deptID string) ([]string, error) {
	allDepts, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	var subIDs []string
	subIDs = append(subIDs, deptID)
	r.collectSubDepts(deptID, allDepts, &subIDs)
	return subIDs, nil
}

func (r *DepartmentRepository) collectSubDepts(parentID string, allDepts []models.Department, result *[]string) {
	for _, d := range allDepts {
		if d.ParentID != nil && d.ParentID.String() == parentID {
			*result = append(*result, d.ID.String())
			r.collectSubDepts(d.ID.String(), allDepts, result)
		}
	}
}

type DepartmentTreeNode struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	ParentID    string               `json:"parent_id"`
	Level       int                  `json:"level"`
	LeaderName  string               `json:"leader_name,omitempty"`
	MemberCount int64                `json:"member_count"`
	Children    []DepartmentTreeNode `json:"children"`
}
