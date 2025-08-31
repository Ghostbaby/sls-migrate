package store

import (
	"context"
	"fmt"

	"github.com/Ghostbaby/sls-migrate/internal/models"
	"github.com/Ghostbaby/sls-migrate/pkg/database"
	"gorm.io/gorm"
)

// AlertStore Alert 数据存储接口
type AlertStore interface {
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, id uint) (*models.Alert, error)
	GetByName(ctx context.Context, name string) (*models.Alert, error)
	Update(ctx context.Context, alert *models.Alert) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*models.Alert, int64, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.Alert, int64, error)
	CreateWithTransaction(ctx context.Context, alert *models.Alert) error
	UpdateWithTransaction(ctx context.Context, alert *models.Alert) error
	Count(ctx context.Context) (int64, error)
}

// alertStore Alert 数据存储实现
type alertStore struct {
	db *gorm.DB
}

// NewAlertStore 创建新的 AlertStore 实例
func NewAlertStore() AlertStore {
	return &alertStore{
		db: database.DB,
	}
}

// Create 创建 Alert
func (s *alertStore) Create(ctx context.Context, alert *models.Alert) error {
	return s.db.WithContext(ctx).Create(alert).Error
}

// GetByID 根据 ID 获取 Alert
func (s *alertStore) GetByID(ctx context.Context, id uint) (*models.Alert, error) {
	var alert models.Alert
	err := s.db.WithContext(ctx).
		Preload("Configuration").
		Preload("Configuration.ConditionConfig").
		Preload("Configuration.GroupConfig").
		Preload("Configuration.PolicyConfig").
		Preload("Configuration.TemplateConfig").
		Preload("Configuration.SeverityConfigs").
		Preload("Schedule").
		Preload("Tags").
		Preload("Queries").
		First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetByName 根据名称获取 Alert
func (s *alertStore) GetByName(ctx context.Context, name string) (*models.Alert, error) {
	var alert models.Alert
	err := s.db.WithContext(ctx).
		Preload("Configuration").
		Preload("Configuration.ConditionConfig").
		Preload("Configuration.GroupConfig").
		Preload("Configuration.PolicyConfig").
		Preload("Configuration.TemplateConfig").
		Preload("Configuration.SeverityConfigs").
		Preload("Schedule").
		Preload("Tags").
		Preload("Queries").
		Where("name = ?", name).
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// Update 更新 Alert
func (s *alertStore) Update(ctx context.Context, alert *models.Alert) error {
	return s.db.WithContext(ctx).Save(alert).Error
}

// Delete 删除 Alert
func (s *alertStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Alert{}, id).Error
}

// List 分页获取 Alert 列表
func (s *alertStore) List(ctx context.Context, offset, limit int) ([]*models.Alert, int64, error) {
	var alerts []*models.Alert
	var total int64

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&models.Alert{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := s.db.WithContext(ctx).
		Preload("Configuration").
		Preload("Schedule").
		Preload("Tags").
		Preload("Queries").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&alerts).Error

	return alerts, total, err
}

// ListByStatus 根据状态分页获取 Alert 列表
func (s *alertStore) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.Alert, int64, error) {
	var alerts []*models.Alert
	var total int64

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&models.Alert{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := s.db.WithContext(ctx).
		Preload("Configuration").
		Preload("Schedule").
		Preload("Tags").
		Preload("Queries").
		Where("status = ?", status).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&alerts).Error

	return alerts, total, err
}

// CreateWithTransaction 在事务中创建 Alert 及其关联数据
func (s *alertStore) CreateWithTransaction(ctx context.Context, alert *models.Alert) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存关联数据的引用
		originalConfig := alert.Configuration
		originalSchedule := alert.Schedule
		originalTags := alert.Tags
		originalQueries := alert.Queries

		// 调试输出
		fmt.Printf("DEBUG: Creating alert %s\n", alert.Name)
		fmt.Printf("DEBUG: originalConfig is nil: %v\n", originalConfig == nil)
		if originalConfig != nil {
			fmt.Printf("DEBUG: originalConfig has data: Type=%v, Version=%v\n", 
				originalConfig.Type, originalConfig.Version)
		}

		// 步骤1: 创建纯净的 Alert 主记录（不包含关联数据）
		cleanAlert := models.Alert{
			Name:             alert.Name,
			DisplayName:      alert.DisplayName,
			Description:      alert.Description,
			Status:           alert.Status,
			CreateTime:       alert.CreateTime,
			LastModifiedTime: alert.LastModifiedTime,
		}

		if err := tx.Create(&cleanAlert).Error; err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}

		// 更新原始alert的ID
		alert.ID = cleanAlert.ID

		// 步骤2: 创建所有独立的配置表记录
		if originalConfig != nil {
			// 创建独立的配置表记录
			if originalConfig.ConditionConfig != nil {
				if err := tx.Create(originalConfig.ConditionConfig).Error; err != nil {
					return fmt.Errorf("failed to create condition configuration: %w", err)
				}
				originalConfig.ConditionConfigID = &originalConfig.ConditionConfig.ID
			}

			if originalConfig.GroupConfig != nil {
				if err := tx.Create(originalConfig.GroupConfig).Error; err != nil {
					return fmt.Errorf("failed to create group configuration: %w", err)
				}
				originalConfig.GroupConfigID = &originalConfig.GroupConfig.ID
			}

			if originalConfig.PolicyConfig != nil {
				if err := tx.Create(originalConfig.PolicyConfig).Error; err != nil {
					return fmt.Errorf("failed to create policy configuration: %w", err)
				}
				originalConfig.PolicyConfigID = &originalConfig.PolicyConfig.ID
			}

			if originalConfig.TemplateConfig != nil {
				if err := tx.Create(originalConfig.TemplateConfig).Error; err != nil {
					return fmt.Errorf("failed to create template configuration: %w", err)
				}
				originalConfig.TemplateConfigID = &originalConfig.TemplateConfig.ID
			}

			// 创建 Sink 配置
			if originalConfig.SinkAlerthubConfig != nil {
				if err := tx.Create(originalConfig.SinkAlerthubConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink alerthub configuration: %w", err)
				}
				originalConfig.SinkAlerthubConfigID = &originalConfig.SinkAlerthubConfig.ID
			}

			if originalConfig.SinkCmsConfig != nil {
				if err := tx.Create(originalConfig.SinkCmsConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink cms configuration: %w", err)
				}
				originalConfig.SinkCmsConfigID = &originalConfig.SinkCmsConfig.ID
			}

			if originalConfig.SinkEventStoreConfig != nil {
				if err := tx.Create(originalConfig.SinkEventStoreConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink event store configuration: %w", err)
				}
				originalConfig.SinkEventStoreConfigID = &originalConfig.SinkEventStoreConfig.ID
			}

			// 步骤3: 创建 alert_configurations 记录
			configToCreate := models.AlertConfiguration{
				AlertID:                    alert.ID,
				AutoAnnotation:             originalConfig.AutoAnnotation,
				Dashboard:                  originalConfig.Dashboard,
				MuteUntil:                  originalConfig.MuteUntil,
				NoDataFire:                 originalConfig.NoDataFire,
				NoDataSeverity:             originalConfig.NoDataSeverity,
				Threshold:                  originalConfig.Threshold,
				Type:                       originalConfig.Type,
				Version:                    originalConfig.Version,
				SendResolved:               originalConfig.SendResolved,
				ConditionConfigID:          originalConfig.ConditionConfigID,
				GroupConfigID:              originalConfig.GroupConfigID,
				PolicyConfigID:             originalConfig.PolicyConfigID,
				TemplateConfigID:           originalConfig.TemplateConfigID,
				SinkAlerthubConfigID:       originalConfig.SinkAlerthubConfigID,
				SinkCmsConfigID:            originalConfig.SinkCmsConfigID,
				SinkEventStoreConfigID:     originalConfig.SinkEventStoreConfigID,
			}

			if err := tx.Create(&configToCreate).Error; err != nil {
				return fmt.Errorf("failed to create alert configuration: %w", err)
			}

			originalConfig.ID = configToCreate.ID
			alert.ConfigurationID = &configToCreate.ID

			// 步骤4: 创建依赖于alert_configurations的记录
			if len(originalConfig.SeverityConfigs) > 0 {
				for i := range originalConfig.SeverityConfigs {
					// 如果有 EvalCondition，先创建它
					if originalConfig.SeverityConfigs[i].EvalCondition != nil {
						if err := tx.Create(originalConfig.SeverityConfigs[i].EvalCondition).Error; err != nil {
							return fmt.Errorf("failed to create eval condition: %w", err)
						}
						originalConfig.SeverityConfigs[i].EvalConditionID = &originalConfig.SeverityConfigs[i].EvalCondition.ID
					}

					originalConfig.SeverityConfigs[i].AlertConfigID = configToCreate.ID
					originalConfig.SeverityConfigs[i].ID = 0
				}
				if err := tx.Create(&originalConfig.SeverityConfigs).Error; err != nil {
					return fmt.Errorf("failed to create severity configurations: %w", err)
				}
			}

			if len(originalConfig.JoinConfigs) > 0 {
				for i := range originalConfig.JoinConfigs {
					originalConfig.JoinConfigs[i].AlertConfigID = configToCreate.ID
					originalConfig.JoinConfigs[i].ID = 0
				}
				if err := tx.Create(&originalConfig.JoinConfigs).Error; err != nil {
					return fmt.Errorf("failed to create join configurations: %w", err)
				}
			}
		}

		// 步骤5: 创建 Schedule
		if originalSchedule != nil {
			scheduleToCreate := models.AlertSchedule{
				AlertID:        alert.ID,
				CronExpression: originalSchedule.CronExpression,
				Delay:          originalSchedule.Delay,
				Interval:       originalSchedule.Interval,
				RunImmediately: originalSchedule.RunImmediately,
				TimeZone:       originalSchedule.TimeZone,
				Type:           originalSchedule.Type,
			}

			if err := tx.Create(&scheduleToCreate).Error; err != nil {
				return fmt.Errorf("failed to create alert schedule: %w", err)
			}
			alert.ScheduleID = &scheduleToCreate.ID
		}

		// 步骤6: 创建 Tags
		if len(originalTags) > 0 {
			tagsToCreate := make([]models.AlertTag, len(originalTags))
			for i, tag := range originalTags {
				tagsToCreate[i] = models.AlertTag{
					AlertID:  alert.ID,
					TagType:  tag.TagType,
					TagKey:   tag.TagKey,
					TagValue: tag.TagValue,
				}
			}
			if err := tx.Create(&tagsToCreate).Error; err != nil {
				return fmt.Errorf("failed to create alert tags: %w", err)
			}
		}

		// 步骤7: 创建 Queries
		if len(originalQueries) > 0 {
			queriesToCreate := make([]models.AlertQuery, len(originalQueries))
			for i, query := range originalQueries {
				queriesToCreate[i] = models.AlertQuery{
					AlertID:      alert.ID,
					ChartTitle:   query.ChartTitle,
					DashboardId:  query.DashboardId,
					End:          query.End,
					PowerSqlMode: query.PowerSqlMode,
					Project:      query.Project,
					Query:        query.Query,
					Region:       query.Region,
					RoleArn:      query.RoleArn,
					Start:        query.Start,
					Store:        query.Store,
					StoreType:    query.StoreType,
					TimeSpanType: query.TimeSpanType,
					Ui:           query.Ui,
				}
			}
			if err := tx.Create(&queriesToCreate).Error; err != nil {
				return fmt.Errorf("failed to create alert queries: %w", err)
			}
		}

		// 步骤8: 最后更新主记录的关联ID
		updateData := map[string]interface{}{}
		if alert.ConfigurationID != nil {
			updateData["configuration_id"] = *alert.ConfigurationID
		}
		if alert.ScheduleID != nil {
			updateData["schedule_id"] = *alert.ScheduleID
		}

		if len(updateData) > 0 {
			if err := tx.Model(&models.Alert{}).Where("id = ?", alert.ID).Updates(updateData).Error; err != nil {
				return fmt.Errorf("failed to update alert with relation IDs: %w", err)
			}
		}

		return nil
	})
}

// UpdateWithTransaction 在事务中更新 Alert 及其关联数据
func (s *alertStore) UpdateWithTransaction(ctx context.Context, alert *models.Alert) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 确保 Alert ID 存在
		if alert.ID == 0 {
			return fmt.Errorf("alert ID is required for update")
		}

		// 设置所有关联记录的外键ID
		if alert.Configuration != nil {
			alert.Configuration.AlertID = alert.ID
		}
		if alert.Schedule != nil {
			alert.Schedule.AlertID = alert.ID
		}
		for i := range alert.Tags {
			alert.Tags[i].AlertID = alert.ID
		}
		for i := range alert.Queries {
			alert.Queries[i].AlertID = alert.ID
		}

		// 使用 Select 排除自动管理的时间戳字段，避免 0000-00-00 错误
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
			Omit("created_at", "updated_at").Save(alert).Error; err != nil {
			return fmt.Errorf("failed to update alert with associations: %w", err)
		}

		return nil
	})
}

// Count 获取 Alert 总数
func (s *alertStore) Count(ctx context.Context) (int64, error) {
	var total int64
	err := s.db.WithContext(ctx).Model(&models.Alert{}).Count(&total).Error
	return total, err
}
