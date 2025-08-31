package database

import (
	"fmt"
	"log"
	"time"

	"github.com/Ghostbaby/sls-migrate/internal/config"
	"github.com/Ghostbaby/sls-migrate/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的 sql.DB 对象
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 禁用外键约束检查
	DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&models.Alert{},
		&models.AlertConfiguration{},
		&models.AlertSchedule{},
		&models.AlertTag{},
		&models.AlertQuery{},
		&models.ConditionConfiguration{},
		&models.GroupConfiguration{},
		&models.PolicyConfiguration{},
		&models.TemplateConfiguration{},
		&models.SeverityConfiguration{},
		&models.JoinConfiguration{},
		&models.SinkAlerthubConfiguration{},
		&models.SinkCmsConfiguration{},
		&models.SinkEventStoreConfiguration{},
	)
	if err != nil {
		// 重新启用外键约束检查
		DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 重新启用外键约束检查
	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("Database tables migrated successfully")
	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}
