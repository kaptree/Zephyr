package services

import (
	"strings"
	"sync"
	"time"

	"labelpro-server/internal/models"

	"gorm.io/gorm"
)

// FilePolicyService 聊天文件传输策略服务
// 提供内存缓存 + 数据库持久化，管理员修改策略后调用 Update 即热加载（无需重启）
type FilePolicyService struct {
	mu       sync.RWMutex
	db       *gorm.DB
	allow    map[string]bool   // 白名单（空 = 不限制）
	blocked  map[string]bool   // 黑名单
	allowRaw string
	blockRaw string
	updated  time.Time
}

var (
	filePolicyOnce sync.Once
	filePolicySvc  *FilePolicyService
)

// 默认黑名单：可执行 / 脚本 / 网页等潜在危险文件
const defaultBlockedExtensions = ".exe,.bat,.cmd,.com,.scr,.msi,.ps1,.vbs,.js,.jar,.dll,.sys,.sh,.bin,.apk,.msc,.reg,.hta,.cpl,.pif,.wsf,.wsh,.html,.htm,.svg"

// GetFilePolicyService 获取全局单例（未初始化时返回 nil）
func GetFilePolicyService() *FilePolicyService {
	return filePolicySvc
}

// InitFilePolicyService 初始化策略服务：加载数据库策略，无记录时写入默认策略
func InitFilePolicyService(db *gorm.DB) *FilePolicyService {
	filePolicyOnce.Do(func() {
		filePolicySvc = &FilePolicyService{db: db}
		if err := filePolicySvc.ensureSeed(); err == nil {
			_ = filePolicySvc.Reload()
		}
	})
	return filePolicySvc
}

// ensureSeed 保证表中存在默认策略记录
func (s *FilePolicyService) ensureSeed() error {
	var count int64
	if err := s.db.Model(&models.ChatFilePolicy{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return s.db.Create(&models.ChatFilePolicy{
			ID:                1,
			AllowExtensions:   "",
			BlockedExtensions: defaultBlockedExtensions,
			UpdatedAt:         time.Now(),
		}).Error
	}
	return nil
}

// Reload 从数据库重新加载策略（热加载入口）
func (s *FilePolicyService) Reload() error {
	var p models.ChatFilePolicy
	if err := s.db.First(&p).Error; err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allow = parseExtensions(p.AllowExtensions)
	s.blocked = parseExtensions(p.BlockedExtensions)
	s.allowRaw = p.AllowExtensions
	s.blockRaw = p.BlockedExtensions
	s.updated = p.UpdatedAt
	return nil
}

// Update 保存策略到数据库并刷新内存缓存（立即生效）
func (s *FilePolicyService) Update(allow, blocked string) error {
	allow = normalizeExtensions(allow)
	blocked = normalizeExtensions(blocked)

	var p models.ChatFilePolicy
	if err := s.db.First(&p).Error; err != nil {
		if err := s.db.Create(&models.ChatFilePolicy{
			ID:                1,
			AllowExtensions:   allow,
			BlockedExtensions: blocked,
			UpdatedAt:         time.Now(),
		}).Error; err != nil {
			return err
		}
	} else {
		if err := s.db.Model(&p).Updates(map[string]interface{}{
			"allow_extensions":   allow,
			"blocked_extensions": blocked,
			"updated_at":         time.Now(),
		}).Error; err != nil {
			return err
		}
	}
	return s.Reload()
}

// AllowExtensions 白名单原始字符串
func (s *FilePolicyService) AllowExtensions() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.allowRaw
}

// BlockedExtensions 黑名单原始字符串
func (s *FilePolicyService) BlockedExtensions() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockRaw
}

// UpdatedAt 策略最后更新时间
func (s *FilePolicyService) UpdatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.updated
}

// IsAllowed 判断扩展名是否允许传输（ext 形如 ".png"）
func (s *FilePolicyService) IsAllowed(ext string) bool {
	ext = strings.ToLower(strings.TrimSpace(ext))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.blocked[ext] {
		return false
	}
	if len(s.allow) > 0 {
		return s.allow[ext]
	}
	return true
}

// parseExtensions 解析逗号分隔扩展名列表为集合（小写、去空白）
func parseExtensions(s string) map[string]bool {
	set := make(map[string]bool)
	for _, part := range strings.Split(s, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			if !strings.HasPrefix(part, ".") {
				part = "." + part
			}
			set[part] = true
		}
	}
	return set
}

// normalizeExtensions 规范化输入：统一为小写、补点、逗号分隔
func normalizeExtensions(s string) string {
	var parts []string
	for _, part := range strings.Split(s, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			if !strings.HasPrefix(part, ".") {
				part = "." + part
			}
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ",")
}
