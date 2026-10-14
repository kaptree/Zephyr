package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/models"

	"gorm.io/gorm"
)

// ErrVaultEntryNotFound 密码本条目不存在（含无权访问他人条目）
var ErrVaultEntryNotFound = errors.New("密码本条目不存在")

// ===== 密码加密（AES-256-GCM，密钥 = SHA-256(配置密钥)）=====

const vaultFallbackSecret = "labelpro-vault-default-secret"

func vaultKey() []byte {
	secret := os.Getenv("VAULT_ENCRYPT_SECRET")
	if secret == "" {
		if cfg := config.GetActive(); cfg != nil {
			secret = cfg.Vault.EncryptSecret
		}
	}
	if secret == "" {
		secret = vaultFallbackSecret
	}
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// EncryptPassword 明文 → base64(nonce + GCM 密文)
func EncryptPassword(plain string) (string, error) {
	block, err := aes.NewCipher(vaultKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, sealed...)), nil
}

// DecryptPassword 密文解密；失败返回空串（旧数据/密钥变更时不 panic）
func DecryptPassword(encoded string) string {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	block, err := aes.NewCipher(vaultKey())
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	if len(raw) < gcm.NonceSize() {
		return ""
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return ""
	}
	return string(plain)
}

// ===== 我的密码本 =====

// ListMyEntries 我的密码本列表（密码已解密，本人数据）
func ListMyEntries(db *gorm.DB, userID string) ([]models.PasswordEntry, error) {
	var list []models.PasswordEntry
	if err := db.Where("user_id = ?", userID).
		Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Password = DecryptPassword(list[i].Password)
	}
	return list, nil
}

func CreateEntry(db *gorm.DB, userID, platform, website, account, password, notes string) (*models.PasswordEntry, error) {
	enc, err := EncryptPassword(password)
	if err != nil {
		return nil, err
	}
	entry := &models.PasswordEntry{
		UserID:   userID,
		Platform: platform,
		Website:  website,
		Account:  account,
		Password: enc,
		Notes:    notes,
	}
	if err := db.Create(entry).Error; err != nil {
		return nil, err
	}
	entry.Password = password
	return entry, nil
}

// UpdateEntry 更新本人条目；password 非空时更新密码，为空保持原值
func UpdateEntry(db *gorm.DB, id, userID, platform, website, account, password, notes string) (*models.PasswordEntry, error) {
	entry, err := getOwnEntry(db, id, userID)
	if err != nil {
		return nil, err
	}
	entry.Platform = platform
	entry.Website = website
	entry.Account = account
	entry.Notes = notes
	if password != "" {
		enc, err := EncryptPassword(password)
		if err != nil {
			return nil, err
		}
		entry.Password = enc
	}
	if err := db.Save(entry).Error; err != nil {
		return nil, err
	}
	if password == "" {
		entry.Password = DecryptPassword(entry.Password)
	} else {
		entry.Password = password
	}
	return entry, nil
}

// DeleteEntry 删除本人条目
func DeleteEntry(db *gorm.DB, id, userID string) error {
	res := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.PasswordEntry{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrVaultEntryNotFound
	}
	return nil
}

// SetEntryShared 共享 / 取消共享本人条目
func SetEntryShared(db *gorm.DB, id, userID string, shared bool) (*models.PasswordEntry, error) {
	entry, err := getOwnEntry(db, id, userID)
	if err != nil {
		return nil, err
	}
	if shared {
		now := time.Now()
		entry.SharedAt = &now
	} else {
		entry.SharedAt = nil
	}
	if err := db.Save(entry).Error; err != nil {
		return nil, err
	}
	entry.Password = DecryptPassword(entry.Password)
	return entry, nil
}

func getOwnEntry(db *gorm.DB, id, userID string) (*models.PasswordEntry, error) {
	var entry models.PasswordEntry
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultEntryNotFound
		}
		return nil, err
	}
	return &entry, nil
}

// ===== 密码广场 =====

// ListPlaza 所有共享条目（不含密码，附共享者姓名与按时间倒序）
func ListPlaza(db *gorm.DB) ([]models.PlazaEntry, error) {
	var list []models.PlazaEntry
	err := db.Table("password_entries p").
		Select("p.id, p.platform, p.website, p.account, p.notes, p.user_id AS owner_id, "+
			"COALESCE(u.username, '匿名用户') AS shared_by_name, p.shared_at").
		Joins("LEFT JOIN users u ON u.id::text = p.user_id").
		Where("p.shared_at IS NOT NULL").
		Order("p.shared_at DESC").
		Scan(&list).Error
	return list, err
}

// RevealPlazaPassword 查看广场条目密码（仅已共享条目）
func RevealPlazaPassword(db *gorm.DB, id string) (string, error) {
	var entry models.PasswordEntry
	if err := db.Where("id = ? AND shared_at IS NOT NULL", id).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrVaultEntryNotFound
		}
		return "", err
	}
	return DecryptPassword(entry.Password), nil
}
