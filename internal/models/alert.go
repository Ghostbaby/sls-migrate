package models

import (
	"time"
)

// Alert 主表模型
type Alert struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	DisplayName      string    `json:"display_name" gorm:"type:varchar(255);not null"`
	Description      *string   `json:"description" gorm:"type:text"`
	Status           string    `json:"status" gorm:"type:varchar(50);default:'ENABLED'"`
	CreateTime       *int64    `json:"create_time" gorm:"type:bigint"`
	LastModifiedTime *int64    `json:"last_modified_time" gorm:"type:bigint"`
	ConfigurationID  *uint     `json:"configuration_id"`
	ScheduleID       *uint     `json:"schedule_id"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Configuration *AlertConfiguration `json:"configuration" gorm:"foreignKey:ConfigurationID"`
	Schedule      *AlertSchedule      `json:"schedule" gorm:"foreignKey:ScheduleID"`
	Tags          []AlertTag          `json:"tags" gorm:"foreignKey:AlertID"`
	Queries       []AlertQuery        `json:"queries" gorm:"foreignKey:AlertID"`
}

// TableName 指定表名
func (Alert) TableName() string {
	return "alerts"
}

// AlertConfiguration 配置表模型 - 完全匹配 SLS SDK
type AlertConfiguration struct {
	ID                    uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertID               uint      `json:"alert_id" gorm:"not null"`
	AutoAnnotation        *bool     `json:"auto_annotation" gorm:"type:boolean;default:false"`
	Dashboard             *string   `json:"dashboard" gorm:"type:varchar(255)"`
	MuteUntil             *int64    `json:"mute_until" gorm:"type:bigint"`
	NoDataFire            *bool     `json:"no_data_fire" gorm:"type:boolean;default:false"`
	NoDataSeverity        *int32    `json:"no_data_severity" gorm:"type:int"`
	Threshold             *int32    `json:"threshold" gorm:"type:int"`
	Type                  *string   `json:"type" gorm:"type:varchar(100)"`
	Version               *string   `json:"version" gorm:"type:varchar(50)"`
	SendResolved          *bool     `json:"send_resolved" gorm:"type:boolean;default:false"`
	ConditionConfigID     *uint     `json:"condition_config_id"`
	GroupConfigID         *uint     `json:"group_config_id"`
	PolicyConfigID        *uint     `json:"policy_config_id"`
	TemplateConfigID      *uint     `json:"template_config_id"`
	SinkAlerthubConfigID  *uint     `json:"sink_alerthub_config_id"`
	SinkCmsConfigID       *uint     `json:"sink_cms_config_id"`
	SinkEventStoreConfigID *uint    `json:"sink_event_store_config_id"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Alert              Alert                      `json:"alert" gorm:"foreignKey:AlertID"`
	ConditionConfig    *ConditionConfiguration    `json:"condition_config" gorm:"foreignKey:ConditionConfigID"`
	GroupConfig        *GroupConfiguration        `json:"group_config" gorm:"foreignKey:GroupConfigID"`
	PolicyConfig       *PolicyConfiguration      `json:"policy_config" gorm:"foreignKey:PolicyConfigID"`
	TemplateConfig     *TemplateConfiguration    `json:"template_config" gorm:"foreignKey:TemplateConfigID"`
	SeverityConfigs    []SeverityConfiguration   `json:"severity_configs" gorm:"foreignKey:AlertConfigID"`
	JoinConfigs        []JoinConfiguration       `json:"join_configs" gorm:"foreignKey:AlertConfigID"`
	SinkAlerthubConfig *SinkAlerthubConfiguration `json:"sink_alerthub_config" gorm:"foreignKey:SinkAlerthubConfigID"`
	SinkCmsConfig      *SinkCmsConfiguration     `json:"sink_cms_config" gorm:"foreignKey:SinkCmsConfigID"`
	SinkEventStoreConfig *SinkEventStoreConfiguration `json:"sink_event_store_config" gorm:"foreignKey:SinkEventStoreConfigID"`
}

// TableName 指定表名
func (AlertConfiguration) TableName() string {
	return "alert_configurations"
}

// AlertSchedule 调度表模型 - 完全匹配 SLS SDK
type AlertSchedule struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertID        uint      `json:"alert_id" gorm:"not null"`
	CronExpression *string   `json:"cron_expression" gorm:"type:varchar(100)"`
	Delay          *int32    `json:"delay" gorm:"type:int"`
	Interval       *string   `json:"interval" gorm:"type:varchar(50)"`
	RunImmediately *bool     `json:"run_immediately" gorm:"type:boolean;default:false"`
	TimeZone       *string   `json:"time_zone" gorm:"type:varchar(50)"`
	Type           string    `json:"type" gorm:"type:varchar(50);not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Alert Alert `json:"alert" gorm:"foreignKey:AlertID"`
}

// TableName 指定表名
func (AlertSchedule) TableName() string {
	return "alert_schedules"
}

// AlertTag 标签表模型 - 完全匹配 SLS SDK
type AlertTag struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertID   uint      `json:"alert_id" gorm:"not null"`
	TagType   string    `json:"tag_type" gorm:"type:enum('annotation','label');not null"`
	TagKey    string    `json:"tag_key" gorm:"type:varchar(255);not null"`
	TagValue  *string   `json:"tag_value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联关系
	Alert Alert `json:"alert" gorm:"foreignKey:AlertID"`
}

// TableName 指定表名
func (AlertTag) TableName() string {
	return "alert_tags"
}

// AlertQuery 查询表模型 - 完全匹配 SLS SDK
type AlertQuery struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertID     uint      `json:"alert_id" gorm:"not null"`
	ChartTitle  *string   `json:"chart_title" gorm:"type:varchar(255)"`
	DashboardId *string   `json:"dashboard_id" gorm:"type:varchar(255)"`
	End         *string   `json:"end" gorm:"type:varchar(100)"`
	PowerSqlMode *string  `json:"power_sql_mode" gorm:"type:varchar(50)"`
	Project     *string   `json:"project" gorm:"type:varchar(255)"`
	Query       string    `json:"query" gorm:"type:text;not null"`
	Region      *string   `json:"region" gorm:"type:varchar(100)"`
	RoleArn     *string   `json:"role_arn" gorm:"type:varchar(500)"`
	Start       *string   `json:"start" gorm:"type:varchar(100)"`
	Store       *string   `json:"store" gorm:"type:varchar(255)"`
	StoreType   *string   `json:"store_type" gorm:"type:varchar(100)"`
	TimeSpanType *string  `json:"time_span_type" gorm:"type:varchar(50)"`
	Ui          *string   `json:"ui" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Alert Alert `json:"alert" gorm:"foreignKey:AlertID"`
}

// TableName 指定表名
func (AlertQuery) TableName() string {
	return "alert_queries"
}

// ConditionConfiguration 条件配置表模型 - 完全匹配 SLS SDK
type ConditionConfiguration struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Condition      *string   `json:"condition" gorm:"type:text"`
	CountCondition *string   `json:"count_condition" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (ConditionConfiguration) TableName() string {
	return "condition_configurations"
}

// GroupConfiguration 分组配置表模型 - 完全匹配 SLS SDK
type GroupConfiguration struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Fields        *string   `json:"fields" gorm:"type:text"` // 存储为逗号分隔的字符串
	Type          *string   `json:"type" gorm:"type:varchar(100)"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (GroupConfiguration) TableName() string {
	return "group_configurations"
}

// PolicyConfiguration 策略配置表模型 - 完全匹配 SLS SDK
type PolicyConfiguration struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ActionPolicyId  *string   `json:"action_policy_id" gorm:"type:varchar(255)"`
	AlertPolicyId   *string   `json:"alert_policy_id" gorm:"type:varchar(255)"`
	RepeatInterval  *string   `json:"repeat_interval" gorm:"type:varchar(100)"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (PolicyConfiguration) TableName() string {
	return "policy_configurations"
}

// TemplateConfiguration 模板配置表模型 - 完全匹配 SLS SDK
type TemplateConfiguration struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TemplateId   *string   `json:"template_id" gorm:"type:varchar(255)"`
	Lang         *string   `json:"lang" gorm:"type:varchar(10)"`
	Type         *string   `json:"type" gorm:"type:varchar(100)"`
	Version      *string   `json:"version" gorm:"type:varchar(50)"`
	Aonotations  *string   `json:"aonotations" gorm:"type:json"` // 存储为 JSON 字符串
	Tokens       *string   `json:"tokens" gorm:"type:json"`      // 存储为 JSON 字符串
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (TemplateConfiguration) TableName() string {
	return "template_configurations"
}

// SeverityConfiguration 严重程度配置表模型 - 完全匹配 SLS SDK
type SeverityConfiguration struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertConfigID  uint      `json:"alert_config_id" gorm:"not null"`
	Severity       *int32    `json:"severity" gorm:"type:int"`
	EvalConditionID *uint    `json:"eval_condition_id"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	AlertConfig    AlertConfiguration     `json:"alert_config" gorm:"foreignKey:AlertConfigID"`
	EvalCondition  *ConditionConfiguration `json:"eval_condition" gorm:"foreignKey:EvalConditionID"`
}

// TableName 指定表名
func (SeverityConfiguration) TableName() string {
	return "severity_configurations"
}

// JoinConfiguration 关联配置表模型 - 新增，匹配 SLS SDK
type JoinConfiguration struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertConfigID uint      `json:"alert_config_id" gorm:"not null"`
	JoinType      *string   `json:"join_type" gorm:"type:varchar(100)"`
	JoinConfig    *string   `json:"join_config" gorm:"type:json"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	AlertConfig AlertConfiguration `json:"alert_config" gorm:"foreignKey:AlertConfigID"`
}

// TableName 指定表名
func (JoinConfiguration) TableName() string {
	return "join_configurations"
}

// SinkAlerthubConfiguration 告警中心配置表模型 - 新增，匹配 SLS SDK
type SinkAlerthubConfiguration struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Enabled   *bool     `json:"enabled" gorm:"type:boolean;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SinkAlerthubConfiguration) TableName() string {
	return "sink_alerthub_configurations"
}

// SinkCmsConfiguration 云监控配置表模型 - 新增，匹配 SLS SDK
type SinkCmsConfiguration struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Enabled   *bool     `json:"enabled" gorm:"type:boolean;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SinkCmsConfiguration) TableName() string {
	return "sink_cms_configurations"
}

// SinkEventStoreConfiguration 事件存储配置表模型 - 新增，匹配 SLS SDK
type SinkEventStoreConfiguration struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Enabled   *bool     `json:"enabled" gorm:"type:boolean;default:false"`
	Endpoint  *string   `json:"endpoint" gorm:"type:varchar(500)"`
	EventStore *string  `json:"event_store" gorm:"type:varchar(255)"`
	Project   *string   `json:"project" gorm:"type:varchar(255)"`
	RoleArn   *string   `json:"role_arn" gorm:"type:varchar(500)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SinkEventStoreConfiguration) TableName() string {
	return "sink_event_store_configurations"
}
