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
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 步骤1: 删除所有关联的子表数据
		if err := s.deleteConfigurationAssociations(tx, id); err != nil {
			return fmt.Errorf("failed to delete configuration associations: %w", err)
		}

		// 步骤2: 删除 Configuration 记录
		if err := tx.Where("alert_id = ?", id).Delete(&models.AlertConfiguration{}).Error; err != nil {
			return fmt.Errorf("failed to delete alert configuration: %w", err)
		}

		// 步骤3: 删除 Schedule 记录
		if err := tx.Where("alert_id = ?", id).Delete(&models.AlertSchedule{}).Error; err != nil {
			return fmt.Errorf("failed to delete alert schedule: %w", err)
		}

		// 步骤4: 删除 Tags 记录
		if err := tx.Where("alert_id = ?", id).Delete(&models.AlertTag{}).Error; err != nil {
			return fmt.Errorf("failed to delete alert tags: %w", err)
		}

		// 步骤5: 删除 Queries 记录
		if err := tx.Where("alert_id = ?", id).Delete(&models.AlertQuery{}).Error; err != nil {
			return fmt.Errorf("failed to delete alert queries: %w", err)
		}

		// 步骤6: 最后删除主记录
		if err := tx.Delete(&models.Alert{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete alert: %w", err)
		}

		return nil
	})
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

		// 步骤2: 先创建 alert_configurations 记录
		if originalConfig != nil {
			configToCreate := models.AlertConfiguration{
				AlertID:        alert.ID,
				AutoAnnotation: originalConfig.AutoAnnotation,
				Dashboard:      originalConfig.Dashboard,
				MuteUntil:      originalConfig.MuteUntil,
				NoDataFire:     originalConfig.NoDataFire,
				NoDataSeverity: originalConfig.NoDataSeverity,
				Threshold:      originalConfig.Threshold,
				Type:           originalConfig.Type,
				Version:        originalConfig.Version,
				SendResolved:   originalConfig.SendResolved,
			}

			if err := tx.Create(&configToCreate).Error; err != nil {
				return fmt.Errorf("failed to create alert configuration: %w", err)
			}

			originalConfig.ID = configToCreate.ID
			alert.ConfigurationID = &configToCreate.ID

			// 步骤3: 创建所有配置表记录，并设置 alert_config_id
			if originalConfig.ConditionConfig != nil {
				originalConfig.ConditionConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.ConditionConfig).Error; err != nil {
					return fmt.Errorf("failed to create condition configuration: %w", err)
				}
			}

			if originalConfig.GroupConfig != nil {
				originalConfig.GroupConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.GroupConfig).Error; err != nil {
					return fmt.Errorf("failed to create group configuration: %w", err)
				}
			}

			if originalConfig.PolicyConfig != nil {
				originalConfig.PolicyConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.PolicyConfig).Error; err != nil {
					return fmt.Errorf("failed to create policy configuration: %w", err)
				}
			}

			if originalConfig.TemplateConfig != nil {
				originalConfig.TemplateConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.TemplateConfig).Error; err != nil {
					return fmt.Errorf("failed to create template configuration: %w", err)
				}
			}

			// 创建 Sink 配置
			if originalConfig.SinkAlerthubConfig != nil {
				originalConfig.SinkAlerthubConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.SinkAlerthubConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink alerthub configuration: %w", err)
				}
			}

			if originalConfig.SinkCmsConfig != nil {
				originalConfig.SinkCmsConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.SinkCmsConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink cms configuration: %w", err)
				}
			}

			if originalConfig.SinkEventStoreConfig != nil {
				originalConfig.SinkEventStoreConfig.AlertConfigID = configToCreate.ID
				if err := tx.Create(originalConfig.SinkEventStoreConfig).Error; err != nil {
					return fmt.Errorf("failed to create sink event store configuration: %w", err)
				}
			}

			// 步骤4: 创建依赖于alert_configurations的记录
			if len(originalConfig.SeverityConfigs) > 0 {
				for i := range originalConfig.SeverityConfigs {
					// 如果有 EvalCondition，先创建它
					if originalConfig.SeverityConfigs[i].EvalCondition != nil {
						// EvalCondition 需要设置 alert_config_id，它应该引用 SeverityConfig 所属的 alert_config
						originalConfig.SeverityConfigs[i].EvalCondition.AlertConfigID = configToCreate.ID
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

// deleteConfigurationAssociations 删除 Configuration 的所有关联数据
func (s *alertStore) deleteConfigurationAssociations(tx *gorm.DB, alertID uint) error {
	// 先获取 Configuration ID
	var configID uint
	if err := tx.Model(&models.AlertConfiguration{}).Where("alert_id = ?", alertID).Select("id").First(&configID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // 没有 Configuration，直接返回
		}
		return fmt.Errorf("failed to get configuration ID: %w", err)
	}

	// 删除所有关联的子表数据
	if err := tx.Where("alert_config_id = ?", configID).Delete(&models.SeverityConfiguration{}).Error; err != nil {
		return fmt.Errorf("failed to delete severity configurations: %w", err)
	}

	if err := tx.Where("alert_config_id = ?", configID).Delete(&models.JoinConfiguration{}).Error; err != nil {
		return fmt.Errorf("failed to delete join configurations: %w", err)
	}

	// 注意：这里不删除 Configuration 本身，因为主删除方法会处理
	return nil
}

// recreateConfiguration 重新创建 Configuration 及其关联数据
func (s *alertStore) recreateConfiguration(tx *gorm.DB, alert *models.Alert) error {
	if alert.Configuration == nil {
		return nil
	}

	// 创建新的 Configuration
	configToCreate := models.AlertConfiguration{
		AlertID:        alert.ID,
		AutoAnnotation: alert.Configuration.AutoAnnotation,
		Dashboard:      alert.Configuration.Dashboard,
		MuteUntil:      alert.Configuration.MuteUntil,
		NoDataFire:     alert.Configuration.NoDataFire,
		NoDataSeverity: alert.Configuration.NoDataSeverity,
		Threshold:      alert.Configuration.Threshold,
		Type:           alert.Configuration.Type,
		Version:        alert.Configuration.Version,
		SendResolved:   alert.Configuration.SendResolved,
	}

	// 创建独立的配置表记录
	if alert.Configuration.ConditionConfig != nil {
		if err := tx.Create(alert.Configuration.ConditionConfig).Error; err != nil {
			return fmt.Errorf("failed to create condition configuration: %w", err)
		}
		configToCreate.ConditionConfigID = &alert.Configuration.ConditionConfig.ID
	}

	if alert.Configuration.GroupConfig != nil {
		if err := tx.Create(alert.Configuration.GroupConfig).Error; err != nil {
			return fmt.Errorf("failed to create group configuration: %w", err)
		}
		configToCreate.GroupConfigID = &alert.Configuration.GroupConfig.ID
	}

	if alert.Configuration.PolicyConfig != nil {
		if err := tx.Create(alert.Configuration.PolicyConfig).Error; err != nil {
			return fmt.Errorf("failed to create policy configuration: %w", err)
		}
		configToCreate.PolicyConfigID = &alert.Configuration.PolicyConfig.ID
	}

	if alert.Configuration.TemplateConfig != nil {
		if err := tx.Create(alert.Configuration.TemplateConfig).Error; err != nil {
			return fmt.Errorf("failed to create template configuration: %w", err)
		}
		configToCreate.TemplateConfigID = &alert.Configuration.TemplateConfig.ID
	}

	// 创建 Sink 配置
	if alert.Configuration.SinkAlerthubConfig != nil {
		if err := tx.Create(alert.Configuration.SinkAlerthubConfig).Error; err != nil {
			return fmt.Errorf("failed to create sink alerthub configuration: %w", err)
		}
		configToCreate.SinkAlerthubConfigID = &alert.Configuration.SinkAlerthubConfig.ID
	}

	if alert.Configuration.SinkCmsConfig != nil {
		if err := tx.Create(alert.Configuration.SinkCmsConfig).Error; err != nil {
			return fmt.Errorf("failed to create sink cms configuration: %w", err)
		}
		configToCreate.SinkCmsConfigID = &alert.Configuration.SinkCmsConfig.ID
	}

	if alert.Configuration.SinkEventStoreConfig != nil {
		if err := tx.Create(alert.Configuration.SinkEventStoreConfig).Error; err != nil {
			return fmt.Errorf("failed to create sink event store configuration: %w", err)
		}
		configToCreate.SinkEventStoreConfigID = &alert.Configuration.SinkEventStoreConfig.ID
	}

	// 创建 Configuration 记录
	if err := tx.Create(&configToCreate).Error; err != nil {
		return fmt.Errorf("failed to create alert configuration: %w", err)
	}

	alert.ConfigurationID = &configToCreate.ID

	// 创建依赖于 alert_configurations 的记录
	if len(alert.Configuration.SeverityConfigs) > 0 {
		for i := range alert.Configuration.SeverityConfigs {
			// 如果有 EvalCondition，先创建它
			if alert.Configuration.SeverityConfigs[i].EvalCondition != nil {
				if err := tx.Create(alert.Configuration.SeverityConfigs[i].EvalCondition).Error; err != nil {
					return fmt.Errorf("failed to create eval condition: %w", err)
				}
				alert.Configuration.SeverityConfigs[i].EvalConditionID = &alert.Configuration.SeverityConfigs[i].EvalCondition.ID
			}

			alert.Configuration.SeverityConfigs[i].AlertConfigID = configToCreate.ID
			alert.Configuration.SeverityConfigs[i].ID = 0
		}
		if err := tx.Create(&alert.Configuration.SeverityConfigs).Error; err != nil {
			return fmt.Errorf("failed to create severity configurations: %w", err)
		}
	}

	if len(alert.Configuration.JoinConfigs) > 0 {
		for i := range alert.Configuration.JoinConfigs {
			alert.Configuration.JoinConfigs[i].AlertConfigID = configToCreate.ID
			alert.Configuration.JoinConfigs[i].ID = 0
		}
		if err := tx.Create(&alert.Configuration.JoinConfigs).Error; err != nil {
			return fmt.Errorf("failed to create join configurations: %w", err)
		}
	}

	return nil
}

// UpdateWithTransaction 在事务中更新 Alert 及其关联数据
func (s *alertStore) UpdateWithTransaction(ctx context.Context, alert *models.Alert) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 确保 Alert ID 存在
		if alert.ID == 0 {
			return fmt.Errorf("alert ID is required for update")
		}

		// 步骤1: 更新主记录
		updateData := map[string]interface{}{
			"display_name":       alert.DisplayName,
			"description":        alert.Description,
			"status":             alert.Status,
			"last_modified_time": alert.LastModifiedTime,
		}

		if err := tx.Model(&models.Alert{}).Where("id = ?", alert.ID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update alert: %w", err)
		}

		// 步骤2: 处理 Configuration 更新
		if alert.Configuration != nil {
			// 先删除旧的关联数据（但不删除主配置记录）
			if err := s.deleteConfigurationAssociations(tx, alert.ID); err != nil {
				return fmt.Errorf("failed to delete old configuration associations: %w", err)
			}

			// 更新现有的 Configuration 记录
			if err := s.updateConfiguration(tx, alert); err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}
		}

		// 步骤3: 处理 Schedule 更新
		if alert.Schedule != nil {
			// 删除旧的 Schedule
			if err := tx.Where("alert_id = ?", alert.ID).Delete(&models.AlertSchedule{}).Error; err != nil {
				return fmt.Errorf("failed to delete old schedule: %w", err)
			}

			// 创建新的 Schedule
			scheduleToCreate := models.AlertSchedule{
				AlertID:        alert.ID,
				CronExpression: alert.Schedule.CronExpression,
				Delay:          alert.Schedule.Delay,
				Interval:       alert.Schedule.Interval,
				RunImmediately: alert.Schedule.RunImmediately,
				TimeZone:       alert.Schedule.TimeZone,
				Type:           alert.Schedule.Type,
			}

			if err := tx.Create(&scheduleToCreate).Error; err != nil {
				return fmt.Errorf("failed to create new schedule: %w", err)
			}
			alert.ScheduleID = &scheduleToCreate.ID
		}

		// 步骤4: 处理 Tags 更新
		if len(alert.Tags) > 0 {
			// 删除旧的 Tags
			if err := tx.Where("alert_id = ?", alert.ID).Delete(&models.AlertTag{}).Error; err != nil {
				return fmt.Errorf("failed to delete old tags: %w", err)
			}

			// 创建新的 Tags
			tagsToCreate := make([]models.AlertTag, len(alert.Tags))
			for i, tag := range alert.Tags {
				tagsToCreate[i] = models.AlertTag{
					AlertID:  alert.ID,
					TagType:  tag.TagType,
					TagKey:   tag.TagKey,
					TagValue: tag.TagValue,
				}
			}
			if err := tx.Create(&tagsToCreate).Error; err != nil {
				return fmt.Errorf("failed to create new tags: %w", err)
			}
		}

		// 步骤5: 处理 Queries 更新
		if len(alert.Queries) > 0 {
			// 删除旧的 Queries
			if err := tx.Where("alert_id = ?", alert.ID).Delete(&models.AlertQuery{}).Error; err != nil {
				return fmt.Errorf("failed to delete old queries: %w", err)
			}

			// 创建新的 Queries
			queriesToCreate := make([]models.AlertQuery, len(alert.Queries))
			for i, query := range alert.Queries {
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
				return fmt.Errorf("failed to create new queries: %w", err)
			}
		}

		// 步骤6: 更新主记录的关联ID
		updateData = map[string]interface{}{}
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

// Count 获取 Alert 总数
func (s *alertStore) Count(ctx context.Context) (int64, error) {
	var total int64
	err := s.db.WithContext(ctx).Model(&models.Alert{}).Count(&total).Error
	return total, err
}

// updateConfiguration 更新现有的 Configuration 及其关联数据
func (s *alertStore) updateConfiguration(tx *gorm.DB, alert *models.Alert) error {
	if alert.Configuration == nil {
		return nil
	}

	// 获取现有的 Configuration ID
	var existingConfigID uint
	if err := tx.Model(&models.AlertConfiguration{}).Where("alert_id = ?", alert.ID).Select("id").First(&existingConfigID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有现有配置，则创建新的
			return s.recreateConfiguration(tx, alert)
		}
		return fmt.Errorf("failed to get existing configuration ID: %w", err)
	}

	// 更新主配置记录
	updateData := map[string]interface{}{
		"auto_annotation":  alert.Configuration.AutoAnnotation,
		"dashboard":        alert.Configuration.Dashboard,
		"mute_until":       alert.Configuration.MuteUntil,
		"no_data_fire":     alert.Configuration.NoDataFire,
		"no_data_severity": alert.Configuration.NoDataSeverity,
		"threshold":        alert.Configuration.Threshold,
		"type":             alert.Configuration.Type,
		"version":          alert.Configuration.Version,
		"send_resolved":    alert.Configuration.SendResolved,
	}

	if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", existingConfigID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update alert configuration: %w", err)
	}

	// 更新关联的配置记录 - 使用 upsert 逻辑避免重复
	if alert.Configuration.ConditionConfig != nil {
		if err := s.upsertConditionConfig(tx, existingConfigID, alert.Configuration.ConditionConfig); err != nil {
			return fmt.Errorf("failed to upsert condition configuration: %w", err)
		}
	}

	if alert.Configuration.GroupConfig != nil {
		if err := s.upsertGroupConfig(tx, existingConfigID, alert.Configuration.GroupConfig); err != nil {
			return fmt.Errorf("failed to upsert group configuration: %w", err)
		}
	}

	if alert.Configuration.PolicyConfig != nil {
		if err := s.upsertPolicyConfig(tx, existingConfigID, alert.Configuration.PolicyConfig); err != nil {
			return fmt.Errorf("failed to upsert policy configuration: %w", err)
		}
	}

	if alert.Configuration.TemplateConfig != nil {
		if err := s.upsertTemplateConfig(tx, existingConfigID, alert.Configuration.TemplateConfig); err != nil {
			return fmt.Errorf("failed to upsert template configuration: %w", err)
		}
	}

	// 更新 Sink 配置 - 使用 upsert 逻辑避免重复
	if alert.Configuration.SinkAlerthubConfig != nil {
		if err := s.upsertSinkAlerthubConfig(tx, existingConfigID, alert.Configuration.SinkAlerthubConfig); err != nil {
			return fmt.Errorf("failed to upsert sink alerthub configuration: %w", err)
		}
	}

	if alert.Configuration.SinkCmsConfig != nil {
		if err := s.upsertSinkCmsConfig(tx, existingConfigID, alert.Configuration.SinkCmsConfig); err != nil {
			return fmt.Errorf("failed to upsert sink cms configuration: %w", err)
		}
	}

	if alert.Configuration.SinkEventStoreConfig != nil {
		if err := s.upsertSinkEventStoreConfig(tx, existingConfigID, alert.Configuration.SinkEventStoreConfig); err != nil {
			return fmt.Errorf("failed to upsert sink event store configuration: %w", err)
		}
	}

	// 更新依赖于 alert_configurations 的记录
	if len(alert.Configuration.SeverityConfigs) > 0 {
		// 先删除旧的严重程度配置
		if err := tx.Where("alert_config_id = ?", existingConfigID).Delete(&models.SeverityConfiguration{}).Error; err != nil {
			return fmt.Errorf("failed to delete old severity configurations: %w", err)
		}

		// 创建新的严重程度配置
		for i := range alert.Configuration.SeverityConfigs {
			// 如果有 EvalCondition，先创建它
			if alert.Configuration.SeverityConfigs[i].EvalCondition != nil {
				if err := tx.Create(alert.Configuration.SeverityConfigs[i].EvalCondition).Error; err != nil {
					return fmt.Errorf("failed to create eval condition: %w", err)
				}
				alert.Configuration.SeverityConfigs[i].EvalConditionID = &alert.Configuration.SeverityConfigs[i].EvalCondition.ID
			}

			alert.Configuration.SeverityConfigs[i].AlertConfigID = existingConfigID
			alert.Configuration.SeverityConfigs[i].ID = 0
		}
		if err := tx.Create(&alert.Configuration.SeverityConfigs).Error; err != nil {
			return fmt.Errorf("failed to create severity configurations: %w", err)
		}
	}

	if len(alert.Configuration.JoinConfigs) > 0 {
		// 先删除旧的 Join 配置
		if err := tx.Where("alert_config_id = ?", existingConfigID).Delete(&models.JoinConfiguration{}).Error; err != nil {
			return fmt.Errorf("failed to delete old join configurations: %w", err)
		}

		// 创建新的 Join 配置
		for i := range alert.Configuration.JoinConfigs {
			alert.Configuration.JoinConfigs[i].AlertConfigID = existingConfigID
			alert.Configuration.JoinConfigs[i].ID = 0
		}
		if err := tx.Create(&alert.Configuration.JoinConfigs).Error; err != nil {
			return fmt.Errorf("failed to create join configurations: %w", err)
		}
	}

	// 设置主记录的配置ID
	alert.ConfigurationID = &existingConfigID

	return nil
}

// upsertConditionConfig 更新或插入条件配置
func (s *alertStore) upsertConditionConfig(tx *gorm.DB, alertConfigID uint, config *models.ConditionConfiguration) error {
	// 查找现有的条件配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("condition_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing condition config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create condition configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("condition_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update condition config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"condition":       config.Condition,
			"count_condition": config.CountCondition,
		}
		if err := tx.Model(&models.ConditionConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update condition configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertGroupConfig 更新或插入分组配置
func (s *alertStore) upsertGroupConfig(tx *gorm.DB, alertConfigID uint, config *models.GroupConfiguration) error {
	// 查找现有的分组配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("group_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing group config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create group configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("group_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update group config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"fields": config.Fields,
			"type":   config.Type,
		}
		if err := tx.Model(&models.GroupConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update group configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertPolicyConfig 更新或插入策略配置
func (s *alertStore) upsertPolicyConfig(tx *gorm.DB, alertConfigID uint, config *models.PolicyConfiguration) error {
	// 查找现有的策略配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("policy_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing policy config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create policy configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("policy_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update policy config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"alert_policy_id":  config.AlertPolicyId,
			"action_policy_id": config.ActionPolicyId,
			"repeat_interval":  config.RepeatInterval,
		}
		if err := tx.Model(&models.PolicyConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update policy configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertTemplateConfig 更新或插入模板配置
func (s *alertStore) upsertTemplateConfig(tx *gorm.DB, alertConfigID uint, config *models.TemplateConfiguration) error {
	// 查找现有的模板配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("template_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing template config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create template configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("template_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update template config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"template_id": config.TemplateId,
			"lang":        config.Lang,
			"type":        config.Type,
			"version":     config.Version,
			"aonotations": config.Aonotations,
			"tokens":      config.Tokens,
		}
		if err := tx.Model(&models.TemplateConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update template configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertSinkAlerthubConfig 更新或插入 Sink Alerthub 配置
func (s *alertStore) upsertSinkAlerthubConfig(tx *gorm.DB, alertConfigID uint, config *models.SinkAlerthubConfiguration) error {
	// 查找现有的配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("sink_alerthub_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing sink alerthub config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create sink alerthub configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("sink_alerthub_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update sink alerthub config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"enabled": config.Enabled,
		}
		if err := tx.Model(&models.SinkAlerthubConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update sink alerthub configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertSinkCmsConfig 更新或插入 Sink CMS 配置
func (s *alertStore) upsertSinkCmsConfig(tx *gorm.DB, alertConfigID uint, config *models.SinkCmsConfiguration) error {
	// 查找现有的配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("sink_cms_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing sink cms config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create sink cms configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("sink_cms_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update sink cms config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"enabled": config.Enabled,
		}
		if err := tx.Model(&models.SinkCmsConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update sink cms configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}

// upsertSinkEventStoreConfig 更新或插入 Sink Event Store 配置
func (s *alertStore) upsertSinkEventStoreConfig(tx *gorm.DB, alertConfigID uint, config *models.SinkEventStoreConfiguration) error {
	// 查找现有的配置（通过主配置记录的外键引用）
	var existingConfigID *uint
	err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Select("sink_event_store_config_id").First(&existingConfigID).Error
	if err != nil {
		return fmt.Errorf("failed to get existing sink event store config ID: %w", err)
	}

	if existingConfigID == nil || *existingConfigID == 0 {
		// 不存在则创建新的
		config.ID = 0
		if err := tx.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create sink event store configuration: %w", err)
		}
		// 更新主配置记录中的引用ID
		if err := tx.Model(&models.AlertConfiguration{}).Where("id = ?", alertConfigID).Update("sink_event_store_config_id", config.ID).Error; err != nil {
			return fmt.Errorf("failed to update sink event store config reference: %w", err)
		}
	} else {
		// 存在则更新
		updateData := map[string]interface{}{
			"enabled":     config.Enabled,
			"endpoint":    config.Endpoint,
			"event_store": config.EventStore,
			"project":     config.Project,
			"role_arn":    config.RoleArn,
		}
		if err := tx.Model(&models.SinkEventStoreConfiguration{}).Where("id = ?", *existingConfigID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update sink event store configuration: %w", err)
		}
		config.ID = *existingConfigID
	}

	return nil
}
