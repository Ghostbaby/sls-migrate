package service

import (
	"context"
	"fmt"
	"log"

	"github.com/Ghostbaby/sls-migrate/internal/models"
	"github.com/Ghostbaby/sls-migrate/internal/store"
)

// SyncService 同步服务接口
type SyncService interface {
	SyncSLSToDatabase(ctx context.Context) error
	SyncDatabaseToSLS(ctx context.Context) error
	GetSyncStatus(ctx context.Context) (*SyncStatus, error)
}

// SyncStatus 同步状态
type SyncStatus struct {
	LastSyncTime  string `json:"last_sync_time"`
	SLSAlertCount int    `json:"sls_alert_count"`
	DBAlertCount  int    `json:"db_alert_count"`
	SyncedCount   int    `json:"synced_count"`
	FailedCount   int    `json:"failed_count"`
	Status        string `json:"status"`
	LastError     string `json:"last_error,omitempty"`
}

// syncService 同步服务实现
type syncService struct {
	slsService   SLSService
	alertStore   store.AlertStore
	alertService AlertService
}

// NewSyncService 创建新的 SyncService 实例
func NewSyncService(slsService SLSService, alertStore store.AlertStore, alertService AlertService) SyncService {
	return &syncService{
		slsService:   slsService,
		alertStore:   alertStore,
		alertService: alertService,
	}
}

// SyncSLSToDatabase 从阿里云 SLS 同步 Alert 规则到本地数据库
func (s *syncService) SyncSLSToDatabase(ctx context.Context) error {
	log.Println("Starting SLS to Database sync...")

	// 获取 SLS 中的所有 alerts
	slsAlerts, err := s.slsService.GetAlerts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get alerts from SLS: %w", err)
	}

	log.Printf("Found %d alerts in SLS", len(slsAlerts))

	var syncedCount, failedCount, updatedCount, createdCount int
	var lastError string

	for _, slsAlert := range slsAlerts {
		// 检查是否已存在
		existingAlert, err := s.alertStore.GetByName(ctx, slsAlert.Name)
		if err == nil && existingAlert != nil {
			// 检查是否需要更新（比较关键字段）
			if s.needsUpdate(existingAlert, slsAlert) {
				// 更新现有记录
				slsAlert.ID = existingAlert.ID
				if err := s.alertService.UpdateAlert(ctx, slsAlert); err != nil {
					log.Printf("Failed to update alert %s: %v", slsAlert.Name, err)
					failedCount++
					lastError = err.Error()
					continue
				}
				log.Printf("Updated alert: %s", slsAlert.Name)
				updatedCount++
			} else {
				log.Printf("Alert %s is up to date, skipping", slsAlert.Name)
			}
		} else {
			// 创建新记录
			if err := s.alertService.CreateAlert(ctx, slsAlert); err != nil {
				log.Printf("Failed to create alert %s: %v", slsAlert.Name, err)
				failedCount++
				lastError = err.Error()
				continue
			}
			log.Printf("Created alert: %s", slsAlert.Name)
			createdCount++
		}
		syncedCount++
	}

	log.Printf("Sync completed. Total: %d, Created: %d, Updated: %d, Skipped: %d, Failed: %d",
		syncedCount, createdCount, updatedCount, syncedCount-createdCount-updatedCount, failedCount)

	if failedCount > 0 {
		return fmt.Errorf("sync completed with %d failures. Last error: %s", failedCount, lastError)
	}

	return nil
}

// SyncDatabaseToSLS 从本地数据库同步 Alert 规则到阿里云 SLS
func (s *syncService) SyncDatabaseToSLS(ctx context.Context) error {
	log.Println("Starting Database to SLS sync...")

	// 获取数据库中的所有 alerts
	dbAlerts, _, err := s.alertStore.List(ctx, 0, 1000) // 获取所有记录
	if err != nil {
		return fmt.Errorf("failed to get alerts from database: %w", err)
	}

	log.Printf("Found %d alerts in database", len(dbAlerts))

	var syncedCount, failedCount int
	var lastError string

	for _, dbAlert := range dbAlerts {
		// 检查 SLS 中是否已存在
		existingSLSAlert, err := s.slsService.GetAlertByName(ctx, dbAlert.Name)
		if err == nil && existingSLSAlert != nil {
			// 更新现有的 SLS Alert
			if err := s.slsService.UpdateAlert(ctx, dbAlert); err != nil {
				log.Printf("Failed to update alert %s in SLS: %v", dbAlert.Name, err)
				failedCount++
				lastError = err.Error()
				continue
			}
			log.Printf("Updated alert in SLS: %s", dbAlert.Name)
		} else {
			// 创建新的 SLS Alert
			if err := s.slsService.CreateAlert(ctx, dbAlert); err != nil {
				log.Printf("Failed to create alert %s in SLS: %v", dbAlert.Name, err)
				failedCount++
				lastError = err.Error()
				continue
			}
			log.Printf("Created alert in SLS: %s", dbAlert.Name)
		}
		syncedCount++
	}

	log.Printf("Database to SLS sync completed. Synced: %d, Failed: %d", syncedCount, failedCount)

	if failedCount > 0 {
		return fmt.Errorf("sync completed with %d failures. Last error: %s", failedCount, lastError)
	}

	return nil
}

// GetSyncStatus 获取同步状态
func (s *syncService) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	// 获取 SLS 中的 alert 数量
	slsAlerts, err := s.slsService.GetAlerts(ctx)
	slsCount := 0
	if err == nil {
		slsCount = len(slsAlerts)
	}

	// 获取数据库中的 alert 数量
	dbCount, err := s.alertStore.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count database alerts: %w", err)
	}

	status := &SyncStatus{
		SLSAlertCount: slsCount,
		DBAlertCount:  int(dbCount),
		Status:        "unknown",
	}

	if err != nil {
		status.Status = "sls_connection_failed"
		status.LastError = err.Error()
	} else {
		status.Status = "healthy"
	}

	return status, nil
}

// needsUpdate 检查是否需要更新 Alert
func (s *syncService) needsUpdate(existing, new *models.Alert) bool {
	// 比较关键字段，决定是否需要更新
	if existing.LastModifiedTime == nil || new.LastModifiedTime == nil {
		return true // 如果时间戳缺失，保守地选择更新
	}

	// 比较最后修改时间
	if *existing.LastModifiedTime != *new.LastModifiedTime {
		return true
	}

	// 比较其他关键字段
	if existing.DisplayName != new.DisplayName {
		return true
	}

	if existing.Status != new.Status {
		return true
	}

	// 如果有描述字段，也进行比较
	if existing.Description == nil && new.Description != nil {
		return true
	}
	if existing.Description != nil && new.Description == nil {
		return true
	}
	if existing.Description != nil && new.Description != nil && *existing.Description != *new.Description {
		return true
	}

	return false
}
