package service

import (
	"context"
	"fmt"

	"github.com/Ghostbaby/sls-migrate/internal/models"
	"github.com/Ghostbaby/sls-migrate/internal/store"
)

// AlertService Alert 服务接口
type AlertService interface {
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetAlertByID(ctx context.Context, id uint) (*models.Alert, error)
	GetAlertByName(ctx context.Context, name string) (*models.Alert, error)
	UpdateAlert(ctx context.Context, alert *models.Alert) error
	DeleteAlert(ctx context.Context, id uint) error
	ListAlerts(ctx context.Context, page, pageSize int) ([]*models.Alert, int64, error)
	ListAlertsByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.Alert, int64, error)
}

// alertService Alert 服务实现
type alertService struct {
	alertStore store.AlertStore
}

// NewAlertService 创建新的 AlertService 实例
func NewAlertService(alertStore store.AlertStore) AlertService {
	return &alertService{
		alertStore: alertStore,
	}
}

// CreateAlert 创建 Alert
func (s *alertService) CreateAlert(ctx context.Context, alert *models.Alert) error {
	// 验证必填字段
	if err := s.validateAlert(alert); err != nil {
		return err
	}

	// 检查名称是否已存在
	existingAlert, err := s.alertStore.GetByName(ctx, alert.Name)
	if err == nil && existingAlert != nil {
		return fmt.Errorf("alert with name '%s' already exists", alert.Name)
	}

	// 使用事务创建 Alert 及其关联数据
	return s.alertStore.CreateWithTransaction(ctx, alert)
}

// GetAlertByID 根据 ID 获取 Alert
func (s *alertService) GetAlertByID(ctx context.Context, id uint) (*models.Alert, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid alert ID")
	}

	alert, err := s.alertStore.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// GetAlertByName 根据名称获取 Alert
func (s *alertService) GetAlertByName(ctx context.Context, name string) (*models.Alert, error) {
	if name == "" {
		return nil, fmt.Errorf("alert name cannot be empty")
	}

	alert, err := s.alertStore.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// UpdateAlert 更新 Alert
func (s *alertService) UpdateAlert(ctx context.Context, alert *models.Alert) error {
	if alert.ID == 0 {
		return fmt.Errorf("invalid alert ID")
	}

	// 验证必填字段
	if err := s.validateAlert(alert); err != nil {
		return err
	}

	// 检查名称是否已被其他 Alert 使用
	if alert.Name != "" {
		existingAlert, err := s.alertStore.GetByName(ctx, alert.Name)
		if err == nil && existingAlert != nil && existingAlert.ID != alert.ID {
			return fmt.Errorf("alert with name '%s' already exists", alert.Name)
		}
	}

	// 使用事务更新 Alert 及其关联数据
	return s.alertStore.UpdateWithTransaction(ctx, alert)
}

// DeleteAlert 删除 Alert
func (s *alertService) DeleteAlert(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid alert ID")
	}

	// 检查 Alert 是否存在
	_, err := s.alertStore.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	return s.alertStore.Delete(ctx, id)
}

// ListAlerts 分页获取 Alert 列表
func (s *alertService) ListAlerts(ctx context.Context, page, pageSize int) ([]*models.Alert, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.alertStore.List(ctx, offset, pageSize)
}

// ListAlertsByStatus 根据状态分页获取 Alert 列表
func (s *alertService) ListAlertsByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.Alert, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 验证状态值
	if status != "" && status != "ENABLED" && status != "DISABLED" {
		return nil, 0, fmt.Errorf("invalid status: %s", status)
	}

	offset := (page - 1) * pageSize
	return s.alertStore.ListByStatus(ctx, status, offset, pageSize)
}

// validateAlert 验证 Alert 数据
func (s *alertService) validateAlert(alert *models.Alert) error {
	if alert.Name == "" {
		return fmt.Errorf("alert name is required")
	}

	if alert.DisplayName == "" {
		return fmt.Errorf("alert display name is required")
	}

	if alert.Status != "" && alert.Status != "ENABLED" && alert.Status != "DISABLED" {
		return fmt.Errorf("invalid status: %s", alert.Status)
	}

	return nil
}
