package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ghostbaby/sls-migrate/internal/config"
	"github.com/Ghostbaby/sls-migrate/internal/models"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// SLSService SLS 服务接口
type SLSService interface {
	GetAlerts(ctx context.Context) ([]*models.Alert, error)
	GetAlertByName(ctx context.Context, name string) (*models.Alert, error)
	CreateAlert(ctx context.Context, alert *models.Alert) error
	UpdateAlert(ctx context.Context, alert *models.Alert) error
	SyncAlertsToDatabase(ctx context.Context) error
}

// slsService SLS 服务实现
type slsService struct {
	slsClient *sls20201230.Client
	project   string
	logStore  string
}

// NewSLSService 创建新的 SLSService 实例
func NewSLSService(slsConfig *config.SLSConfig) (SLSService, error) {
	client, err := config.CreateSLSClient(slsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS client: %w", err)
	}

	// 创建 SLS 客户端
	slsClient, err := sls20201230.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS client: %w", err)
	}

	return &slsService{
		slsClient: slsClient,
		project:   slsConfig.Project,
		logStore:  slsConfig.LogStore,
	}, nil
}

// GetAlerts 从阿里云 SLS 获取所有 Alert 规则
func (s *slsService) GetAlerts(ctx context.Context) ([]*models.Alert, error) {
	request := &sls20201230.ListAlertsRequest{}
	runtime := &service.RuntimeOptions{}

	response, err := s.slsClient.ListAlertsWithOptions(tea.String(s.project), request, make(map[string]*string), runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts from SLS: %w", err)
	}

	var alerts []*models.Alert
	if response.Body != nil && response.Body.Results != nil {
		for _, slsAlert := range response.Body.Results {
			alert := s.convertSLSAlertToModel(slsAlert)
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// GetAlertByName 根据名称从阿里云 SLS 获取特定 Alert 规则
func (s *slsService) GetAlertByName(ctx context.Context, name string) (*models.Alert, error) {
	// 先获取所有 alerts，然后按名称过滤
	alerts, err := s.GetAlerts(ctx)
	if err != nil {
		return nil, err
	}

	for _, alert := range alerts {
		if alert.Name == name {
			return alert, nil
		}
	}

	return nil, fmt.Errorf("alert with name '%s' not found in SLS", name)
}

// SyncAlertsToDatabase 同步阿里云 SLS 的 Alert 规则到本地数据库
func (s *slsService) SyncAlertsToDatabase(ctx context.Context) error {
	// 获取 SLS 中的所有 alerts
	slsAlerts, err := s.GetAlerts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get alerts from SLS: %w", err)
	}

	// 注意：这个方法现在主要用于获取 SLS 数据
	// 实际的数据库保存逻辑由 SyncService 处理
	// 这里返回获取到的数据，供调用方使用
	fmt.Printf("Found %d alerts in SLS\n", len(slsAlerts))
	for _, alert := range slsAlerts {
		fmt.Printf("Alert: %s (%s)\n", alert.Name, alert.DisplayName)
	}

	return nil
}

// convertSLSAlertToModel 将阿里云 SLS 的 Alert 转换为本地模型
func (s *slsService) convertSLSAlertToModel(slsAlert *sls20201230.Alert) *models.Alert {
	alert := &models.Alert{
		Name:             tea.StringValue(slsAlert.Name),
		DisplayName:      tea.StringValue(slsAlert.DisplayName),
		Description:      slsAlert.Description,
		Status:           tea.StringValue(slsAlert.Status),
		CreateTime:       slsAlert.CreateTime,
		LastModifiedTime: slsAlert.LastModifiedTime,
	}

	// 调试输出
	fmt.Printf("DEBUG: Converting SLS alert %s\n", tea.StringValue(slsAlert.Name))
	fmt.Printf("DEBUG: slsAlert.Configuration is nil: %v\n", slsAlert.Configuration == nil)

	// 转换 Configuration
	if slsAlert.Configuration != nil {
		alert.Configuration = &models.AlertConfiguration{
			AutoAnnotation: slsAlert.Configuration.AutoAnnotation,
			Dashboard:      slsAlert.Configuration.Dashboard,
			MuteUntil:      slsAlert.Configuration.MuteUntil,
			NoDataFire:     slsAlert.Configuration.NoDataFire,
			NoDataSeverity: slsAlert.Configuration.NoDataSeverity,
			Threshold:      slsAlert.Configuration.Threshold,
			Type:           slsAlert.Configuration.Type,
			Version:        slsAlert.Configuration.Version,
			SendResolved:   slsAlert.Configuration.SendResolved,
		}

		// 转换 ConditionConfiguration
		if slsAlert.Configuration.ConditionConfiguration != nil {
			conditionConfig := &models.ConditionConfiguration{
				Condition:      slsAlert.Configuration.ConditionConfiguration.Condition,
				CountCondition: slsAlert.Configuration.ConditionConfiguration.CountCondition,
			}
			alert.Configuration.ConditionConfig = conditionConfig
		}

		// 转换 GroupConfiguration
		if slsAlert.Configuration.GroupConfiguration != nil {
			// 将 Fields 数组转换为字符串
			var fieldsStr *string
			if len(slsAlert.Configuration.GroupConfiguration.Fields) > 0 {
				fields := make([]string, 0, len(slsAlert.Configuration.GroupConfiguration.Fields))
				for _, field := range slsAlert.Configuration.GroupConfiguration.Fields {
					if field != nil {
						fields = append(fields, *field)
					}
				}
				if len(fields) > 0 {
					fieldsStr = tea.String(strings.Join(fields, ","))
				}
			}

			groupConfig := &models.GroupConfiguration{
				Fields: fieldsStr,
				Type:   slsAlert.Configuration.GroupConfiguration.Type,
			}
			alert.Configuration.GroupConfig = groupConfig
		}

		// 转换 PolicyConfiguration
		if slsAlert.Configuration.PolicyConfiguration != nil {
			policyConfig := &models.PolicyConfiguration{
				AlertPolicyId:  slsAlert.Configuration.PolicyConfiguration.AlertPolicyId,
				ActionPolicyId: slsAlert.Configuration.PolicyConfiguration.ActionPolicyId,
				RepeatInterval: slsAlert.Configuration.PolicyConfiguration.RepeatInterval,
			}
			alert.Configuration.PolicyConfig = policyConfig
		}

		// 转换 TemplateConfiguration
		if slsAlert.Configuration.TemplateConfiguration != nil {
			// 处理 Aonotations 和 Tokens 的 JSON 转换
			var aonotationsJSON, tokensJSON *string

			if slsAlert.Configuration.TemplateConfiguration.Aonotations != nil {
				if aonotationsBytes, err := json.Marshal(slsAlert.Configuration.TemplateConfiguration.Aonotations); err == nil {
					aonotationsJSON = tea.String(string(aonotationsBytes))
				}
			}

			if slsAlert.Configuration.TemplateConfiguration.Tokens != nil {
				if tokensBytes, err := json.Marshal(slsAlert.Configuration.TemplateConfiguration.Tokens); err == nil {
					tokensJSON = tea.String(string(tokensBytes))
				}
			}

			templateConfig := &models.TemplateConfiguration{
				TemplateId:  slsAlert.Configuration.TemplateConfiguration.Id,
				Lang:        slsAlert.Configuration.TemplateConfiguration.Lang,
				Type:        slsAlert.Configuration.TemplateConfiguration.Type,
				Version:     slsAlert.Configuration.TemplateConfiguration.Version,
				Aonotations: aonotationsJSON,
				Tokens:      tokensJSON,
			}
			alert.Configuration.TemplateConfig = templateConfig
		}

		// 转换 SeverityConfigurations
		if slsAlert.Configuration.SeverityConfigurations != nil {
			for _, slsSeverity := range slsAlert.Configuration.SeverityConfigurations {
				severityConfig := &models.SeverityConfiguration{
					Severity: slsSeverity.Severity,
				}

				// 处理 EvalCondition
				if slsSeverity.EvalCondition != nil {
					evalCondition := &models.ConditionConfiguration{
						Condition:      slsSeverity.EvalCondition.Condition,
						CountCondition: slsSeverity.EvalCondition.CountCondition,
					}
					severityConfig.EvalCondition = evalCondition
				}

				alert.Configuration.SeverityConfigs = append(alert.Configuration.SeverityConfigs, *severityConfig)
			}
		}

		// 转换 QueryList
		if slsAlert.Configuration.QueryList != nil {
			for _, slsQuery := range slsAlert.Configuration.QueryList {
				query := &models.AlertQuery{
					ChartTitle:   slsQuery.ChartTitle,
					DashboardId:  slsQuery.DashboardId,
					End:          slsQuery.End,
					PowerSqlMode: slsQuery.PowerSqlMode,
					Project:      slsQuery.Project,
					Query:        tea.StringValue(slsQuery.Query),
					Region:       slsQuery.Region,
					RoleArn:      slsQuery.RoleArn,
					Start:        slsQuery.Start,
					Store:        slsQuery.Store,
					StoreType:    slsQuery.StoreType,
					TimeSpanType: slsQuery.TimeSpanType,
					Ui:           slsQuery.Ui,
				}
				alert.Queries = append(alert.Queries, *query)
			}
		}

		// 转换 Tags
		if slsAlert.Configuration.Tags != nil {
			for _, slsTag := range slsAlert.Configuration.Tags {
				tag := &models.AlertTag{
					TagType:  "label", // 默认为 label 类型
					TagKey:   tea.StringValue(slsTag),
					TagValue: nil, // SLS 中 Tags 是字符串数组
				}
				alert.Tags = append(alert.Tags, *tag)
			}
		}

		// 转换 Sink 配置
		if slsAlert.Configuration.SinkAlerthub != nil {
			sinkAlerthubConfig := &models.SinkAlerthubConfiguration{
				Enabled: slsAlert.Configuration.SinkAlerthub.Enabled,
			}
			alert.Configuration.SinkAlerthubConfig = sinkAlerthubConfig
		}

		if slsAlert.Configuration.SinkCms != nil {
			sinkCmsConfig := &models.SinkCmsConfiguration{
				Enabled: slsAlert.Configuration.SinkCms.Enabled,
			}
			alert.Configuration.SinkCmsConfig = sinkCmsConfig
		}

		if slsAlert.Configuration.SinkEventStore != nil {
			sinkEventStoreConfig := &models.SinkEventStoreConfiguration{
				Enabled:    slsAlert.Configuration.SinkEventStore.Enabled,
				Endpoint:   slsAlert.Configuration.SinkEventStore.Endpoint,
				EventStore: slsAlert.Configuration.SinkEventStore.EventStore,
				Project:    slsAlert.Configuration.SinkEventStore.Project,
				RoleArn:    slsAlert.Configuration.SinkEventStore.RoleArn,
			}
			alert.Configuration.SinkEventStoreConfig = sinkEventStoreConfig
		}

		// 转换 JoinConfigurations
		if slsAlert.Configuration.JoinConfigurations != nil {
			for _, slsJoinConfig := range slsAlert.Configuration.JoinConfigurations {
				// 将 Condition 和 Type 组合到 JoinConfig 字段中
				var joinConfigStr *string
				if slsJoinConfig.Condition != nil || slsJoinConfig.Type != nil {
					joinData := map[string]interface{}{
						"condition": slsJoinConfig.Condition,
						"type":      slsJoinConfig.Type,
					}
					if joinBytes, err := json.Marshal(joinData); err == nil {
						joinConfigStr = tea.String(string(joinBytes))
					}
				}

				joinConfig := &models.JoinConfiguration{
					JoinType:   slsJoinConfig.Type,
					JoinConfig: joinConfigStr,
				}
				alert.Configuration.JoinConfigs = append(alert.Configuration.JoinConfigs, *joinConfig)
			}
		}

		// 转换 Annotations
		if slsAlert.Configuration.Annotations != nil {
			for _, slsAnnotation := range slsAlert.Configuration.Annotations {
				annotation := &models.AlertTag{
					TagType:  "annotation",
					TagKey:   tea.StringValue(slsAnnotation.Key),
					TagValue: slsAnnotation.Value,
				}
				alert.Tags = append(alert.Tags, *annotation)
			}
		}
	}

	// 转换 Schedule
	if slsAlert.Schedule != nil {
		alert.Schedule = &models.AlertSchedule{
			CronExpression: slsAlert.Schedule.CronExpression,
			Delay:          slsAlert.Schedule.Delay,
			Interval:       slsAlert.Schedule.Interval,
			RunImmediately: slsAlert.Schedule.RunImmediately,
			TimeZone:       slsAlert.Schedule.TimeZone,
			Type:           tea.StringValue(slsAlert.Schedule.Type),
		}
	}

	return alert
}

// CreateAlert 在阿里云 SLS 中创建新的 Alert 规则
func (s *slsService) CreateAlert(ctx context.Context, alert *models.Alert) error {
	// 将本地模型转换为 SLS SDK 模型
	slsAlert := s.convertModelToSLSAlert(alert)

	// 创建请求
	request := &sls20201230.CreateAlertRequest{
		Name:          tea.String(alert.Name),
		DisplayName:   tea.String(alert.DisplayName),
		Description:   alert.Description,
		Configuration: slsAlert.Configuration,
		Schedule:      slsAlert.Schedule,
	}

	runtime := &service.RuntimeOptions{}

	// 调用 SLS API 创建 Alert
	_, err := s.slsClient.CreateAlertWithOptions(tea.String(s.project), request, make(map[string]*string), runtime)
	if err != nil {
		return fmt.Errorf("failed to create alert in SLS: %w", err)
	}

	return nil
}

// UpdateAlert 在阿里云 SLS 中更新现有的 Alert 规则
func (s *slsService) UpdateAlert(ctx context.Context, alert *models.Alert) error {
	// 将本地模型转换为 SLS SDK 模型
	slsAlert := s.convertModelToSLSAlert(alert)

	// 创建请求
	request := &sls20201230.UpdateAlertRequest{
		DisplayName:   tea.String(alert.DisplayName),
		Description:   alert.Description,
		Configuration: slsAlert.Configuration,
		Schedule:      slsAlert.Schedule,
	}

	runtime := &service.RuntimeOptions{}

	// 调用 SLS API 更新 Alert
	_, err := s.slsClient.UpdateAlertWithOptions(tea.String(s.project), tea.String(alert.Name), request, make(map[string]*string), runtime)
	if err != nil {
		return fmt.Errorf("failed to update alert in SLS: %w", err)
	}

	return nil
}

// convertModelToSLSAlert 将本地模型转换为 SLS SDK 模型
func (s *slsService) convertModelToSLSAlert(alert *models.Alert) *sls20201230.Alert {
	slsAlert := &sls20201230.Alert{
		Name:             tea.String(alert.Name),
		DisplayName:      tea.String(alert.DisplayName),
		Description:      alert.Description,
		Status:           tea.String(alert.Status),
		CreateTime:       alert.CreateTime,
		LastModifiedTime: alert.LastModifiedTime,
	}

	// 转换 Configuration
	if alert.Configuration != nil {
		slsConfig := &sls20201230.AlertConfiguration{
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

		// 转换 ConditionConfiguration
		if alert.Configuration.ConditionConfig != nil {
			slsConfig.ConditionConfiguration = &sls20201230.ConditionConfiguration{
				Condition:      alert.Configuration.ConditionConfig.Condition,
				CountCondition: alert.Configuration.ConditionConfig.CountCondition,
			}
		}

		// 转换 GroupConfiguration
		if alert.Configuration.GroupConfig != nil {
			var fields []*string
			if alert.Configuration.GroupConfig.Fields != nil {
				fieldList := strings.Split(*alert.Configuration.GroupConfig.Fields, ",")
				for _, field := range fieldList {
					fields = append(fields, tea.String(strings.TrimSpace(field)))
				}
			}

			slsConfig.GroupConfiguration = &sls20201230.GroupConfiguration{
				Fields: fields,
				Type:   alert.Configuration.GroupConfig.Type,
			}
		}

		// 转换 PolicyConfiguration
		if alert.Configuration.PolicyConfig != nil {
			slsConfig.PolicyConfiguration = &sls20201230.PolicyConfiguration{
				ActionPolicyId: alert.Configuration.PolicyConfig.ActionPolicyId,
				AlertPolicyId:  alert.Configuration.PolicyConfig.AlertPolicyId,
				RepeatInterval: alert.Configuration.PolicyConfig.RepeatInterval,
			}
		}

		// 转换 TemplateConfiguration
		if alert.Configuration.TemplateConfig != nil {
			var aonotations map[string]interface{}
			var tokens map[string]interface{}

			if alert.Configuration.TemplateConfig.Aonotations != nil {
				if err := json.Unmarshal([]byte(*alert.Configuration.TemplateConfig.Aonotations), &aonotations); err != nil {
					// 如果解析失败，使用空 map
					aonotations = make(map[string]interface{})
				}
			}

			if alert.Configuration.TemplateConfig.Tokens != nil {
				if err := json.Unmarshal([]byte(*alert.Configuration.TemplateConfig.Tokens), &tokens); err != nil {
					// 如果解析失败，使用空 map
					tokens = make(map[string]interface{})
				}
			}

			slsConfig.TemplateConfiguration = &sls20201230.TemplateConfiguration{
				Id:          alert.Configuration.TemplateConfig.TemplateId,
				Lang:        alert.Configuration.TemplateConfig.Lang,
				Type:        alert.Configuration.TemplateConfig.Type,
				Version:     alert.Configuration.TemplateConfig.Version,
				Aonotations: aonotations,
				Tokens:      tokens,
			}
		}

		// 转换 QueryList
		if len(alert.Queries) > 0 {
			var slsQueries []*sls20201230.AlertQuery
			for _, query := range alert.Queries {
				slsQuery := &sls20201230.AlertQuery{
					ChartTitle:   query.ChartTitle,
					DashboardId:  query.DashboardId,
					End:          query.End,
					PowerSqlMode: query.PowerSqlMode,
					Project:      query.Project,
					Query:        tea.String(query.Query),
					Region:       query.Region,
					RoleArn:      query.RoleArn,
					Start:        query.Start,
					Store:        query.Store,
					StoreType:    query.StoreType,
					TimeSpanType: query.TimeSpanType,
					Ui:           query.Ui,
				}
				slsQueries = append(slsQueries, slsQuery)
			}
			slsConfig.QueryList = slsQueries
		}

		// 转换 Tags
		if len(alert.Tags) > 0 {
			var slsTags []*string
			for _, tag := range alert.Tags {
				if tag.TagType == "label" {
					slsTags = append(slsTags, tea.String(tag.TagKey))
				}
			}
			slsConfig.Tags = slsTags
		}

		// 转换 Annotations
		if len(alert.Tags) > 0 {
			var slsAnnotations []*sls20201230.AlertTag
			for _, tag := range alert.Tags {
				if tag.TagType == "annotation" {
					slsAnnotation := &sls20201230.AlertTag{
						Key:   tea.String(tag.TagKey),
						Value: tag.TagValue,
					}
					slsAnnotations = append(slsAnnotations, slsAnnotation)
				}
			}
			slsConfig.Annotations = slsAnnotations
		}

		// 转换 Sink 配置
		if alert.Configuration.SinkAlerthubConfig != nil {
			slsConfig.SinkAlerthub = &sls20201230.SinkAlerthubConfiguration{
				Enabled: alert.Configuration.SinkAlerthubConfig.Enabled,
			}
		}

		if alert.Configuration.SinkCmsConfig != nil {
			slsConfig.SinkCms = &sls20201230.SinkCmsConfiguration{
				Enabled: alert.Configuration.SinkCmsConfig.Enabled,
			}
		}

		if alert.Configuration.SinkEventStoreConfig != nil {
			slsConfig.SinkEventStore = &sls20201230.SinkEventStoreConfiguration{
				Enabled:    alert.Configuration.SinkEventStoreConfig.Enabled,
				Endpoint:   alert.Configuration.SinkEventStoreConfig.Endpoint,
				EventStore: alert.Configuration.SinkEventStoreConfig.EventStore,
				Project:    alert.Configuration.SinkEventStoreConfig.Project,
				RoleArn:    alert.Configuration.SinkEventStoreConfig.RoleArn,
			}
		}

		slsAlert.Configuration = slsConfig
	}

	// 转换 Schedule
	if alert.Schedule != nil {
		slsAlert.Schedule = &sls20201230.Schedule{
			CronExpression: alert.Schedule.CronExpression,
			Delay:          alert.Schedule.Delay,
			Interval:       alert.Schedule.Interval,
			RunImmediately: alert.Schedule.RunImmediately,
			TimeZone:       alert.Schedule.TimeZone,
			Type:           tea.String(alert.Schedule.Type),
		}
	}

	return slsAlert
}
