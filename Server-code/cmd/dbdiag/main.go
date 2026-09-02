package main

import (
	"fmt"
	"os"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// reset 子命令：清空 schema 重建（用于测试环境重置）
	if len(os.Args) > 1 && os.Args[1] == "reset" {
		cfg, err := config.Load("config.json")
		if err != nil {
			fmt.Println("load config:", err)
			return
		}
		if err := resetSchema(cfg); err != nil {
			fmt.Println("reset failed:", err)
			os.Exit(1)
		}
		fmt.Println("SCHEMA_RESET_OK")
		return
	}
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Println("load config:", err)
		return
	}
	// 先初始化日志器（InitPostgres 成功后内部会调用 logger.Info）
	if err := logger.Init("error", "console", ".gotmp/logs", 10, 1, 1, false, false); err != nil {
		fmt.Println("init logger:", err)
		return
	}
	if err := database.InitPostgres(cfg); err != nil {
		fmt.Printf("DB connect failed: %v\n", err)
		return
	}
	var users []models.User
	database.DB.Where("username IN ?", []string{"zhang3", "zhao", "admin"}).Find(&users)
	for _, u := range users {
		match := utils.CheckPassword("Admin@123", u.PasswordHash)
		fmt.Printf("user=%s name=%s role=%s active=%v hashLen=%d hashPrefix=%.15s pwdMatch=%v\n",
			u.Username, u.Name, u.Role, u.IsActive, len(u.PasswordHash), u.PasswordHash, match)
	}

	var all []models.User
	database.DB.Select("username", "name", "role").Order("created_at").Find(&all)
	fmt.Printf("总用户数: %d\n", len(all))
	for _, u := range all {
		fmt.Printf("  %s | %s | %s\n", u.Username, u.Name, u.Role)
	}
}

func resetSchema(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := db.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
		return err
	}
	return db.Exec("CREATE SCHEMA public").Error
}
