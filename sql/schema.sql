-- Alert 数据库 Schema
-- 创建时间: 2024-12-19
-- 描述: SLS Alert 迁移系统的数据库结构

-- 创建数据库
CREATE DATABASE IF NOT EXISTS sls_migrate CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE sls_migrate;

-- 1. 主表: alerts
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    name VARCHAR(255) NOT NULL COMMENT 'Alert名称，必填字段',
    display_name VARCHAR(255) NOT NULL COMMENT '显示名称，必填字段',
    description TEXT COMMENT '描述信息',
    status VARCHAR(50) DEFAULT 'ENABLED' COMMENT '状态: ENABLED/DISABLED',
    create_time BIGINT COMMENT '创建时间戳',
    last_modified_time BIGINT COMMENT '最后修改时间戳',
    configuration_id BIGINT UNSIGNED COMMENT '配置ID，关联alert_configurations表',
    schedule_id BIGINT UNSIGNED COMMENT '调度ID，关联alert_schedules表',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    UNIQUE KEY uk_name (name),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time),
    INDEX idx_last_modified_time (last_modified_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alert主表';

-- 2. 配置表: alert_configurations
CREATE TABLE IF NOT EXISTS alert_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert ID',
    auto_annotation BOOLEAN DEFAULT FALSE COMMENT '是否自动注解',
    dashboard VARCHAR(255) COMMENT '仪表板名称',
    mute_until BIGINT COMMENT '静音截止时间戳',
    no_data_fire BOOLEAN DEFAULT FALSE COMMENT '无数据时是否触发',
    no_data_severity INT COMMENT '无数据时的严重程度',
    threshold INT COMMENT '阈值',
    `type` VARCHAR(100) COMMENT '类型',
    version VARCHAR(50) COMMENT '版本',
    send_resolved BOOLEAN DEFAULT FALSE COMMENT '是否发送已解决的通知',
    condition_config_id BIGINT UNSIGNED COMMENT '条件配置ID',
    group_config_id BIGINT UNSIGNED COMMENT '分组配置ID',
    policy_config_id BIGINT UNSIGNED COMMENT '策略配置ID',
    template_config_id BIGINT UNSIGNED COMMENT '模板配置ID',
    sink_alerthub_config_id BIGINT UNSIGNED COMMENT 'Sink Alerthub配置ID',
    sink_cms_config_id BIGINT UNSIGNED COMMENT 'Sink CMS配置ID',
    sink_event_store_config_id BIGINT UNSIGNED COMMENT 'Sink Event Store配置ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_type (`type`),
    INDEX idx_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alert配置表';

-- 3. 调度表: alert_schedules
CREATE TABLE IF NOT EXISTS alert_schedules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert ID',
    cron_expression VARCHAR(100) COMMENT 'Cron表达式',
    delay INT COMMENT '延迟时间(秒)',
    `interval` VARCHAR(50) COMMENT '间隔时间',
    run_immediately BOOLEAN DEFAULT FALSE COMMENT '是否立即运行',
    time_zone VARCHAR(50) COMMENT '时区',
    `type` VARCHAR(50) NOT NULL COMMENT '调度类型，必填',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_type (`type`),
    INDEX idx_run_immediately (run_immediately)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alert调度表';

-- 4. 标签表: alert_tags
CREATE TABLE IF NOT EXISTS alert_tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert ID',
    `tag_type` ENUM('annotation', 'label') NOT NULL COMMENT '标签类型: annotation/label',
    `tag_key` VARCHAR(255) NOT NULL COMMENT '标签键',
    tag_value TEXT COMMENT '标签值',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    UNIQUE KEY uk_alert_tag (alert_id, `tag_type`, `tag_key`),
    INDEX idx_alert_id (alert_id),
    INDEX idx_tag_type (`tag_type`),
    INDEX idx_tag_key (`tag_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alert标签表';

-- 5. 查询表: alert_queries
CREATE TABLE IF NOT EXISTS alert_queries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert ID',
    chart_title VARCHAR(255) COMMENT '图表标题',
    dashboard_id VARCHAR(255) COMMENT '仪表板ID',
    `end` VARCHAR(100) COMMENT '结束时间',
    power_sql_mode VARCHAR(50) COMMENT 'Power SQL模式',
    project VARCHAR(255) COMMENT '项目名称',
    query TEXT NOT NULL COMMENT '查询语句',
    region VARCHAR(100) COMMENT '区域',
    role_arn VARCHAR(500) COMMENT '角色ARN',
    start VARCHAR(100) COMMENT '开始时间',
    store VARCHAR(255) COMMENT '存储',
    store_type VARCHAR(100) COMMENT '存储类型',
    time_span_type VARCHAR(50) COMMENT '时间跨度类型',
    ui VARCHAR(255) COMMENT 'UI配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_project (project)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alert查询表';

-- 6. 条件配置表: condition_configurations
CREATE TABLE IF NOT EXISTS condition_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `condition` TEXT COMMENT '条件表达式',
    count_condition TEXT COMMENT '计数条件表达式',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='条件配置表';

-- 7. 分组配置表: group_configurations
CREATE TABLE IF NOT EXISTS group_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    fields TEXT COMMENT '分组字段，逗号分隔',
    `type` VARCHAR(100) COMMENT '分组类型',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分组配置表';

-- 8. 策略配置表: policy_configurations
CREATE TABLE IF NOT EXISTS policy_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_policy_id VARCHAR(255) COMMENT 'Alert策略ID',
    action_policy_id VARCHAR(255) COMMENT 'Action策略ID',
    repeat_interval VARCHAR(100) COMMENT '重复间隔',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='策略配置表';

-- 9. 模板配置表: template_configurations
CREATE TABLE IF NOT EXISTS template_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    template_id VARCHAR(255) COMMENT '模板ID',
    lang VARCHAR(50) COMMENT '语言',
    `type` VARCHAR(100) COMMENT '模板类型',
    version VARCHAR(50) COMMENT '版本',
    aonotations TEXT COMMENT '注解，JSON格式',
    tokens TEXT COMMENT '令牌，JSON格式',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模板配置表';

-- 10. 严重程度配置表: severity_configurations
CREATE TABLE IF NOT EXISTS severity_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_config_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert配置ID',
    severity INT NOT NULL COMMENT '严重程度级别',
    eval_condition_id BIGINT UNSIGNED COMMENT '评估条件ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    FOREIGN KEY (alert_config_id) REFERENCES alert_configurations(id) ON DELETE CASCADE,
    INDEX idx_alert_config_id (alert_config_id),
    INDEX idx_severity (severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='严重程度配置表';

-- 11. Sink Alerthub配置表
CREATE TABLE IF NOT EXISTS sink_alerthub_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sink Alerthub配置表';

-- 12. Sink CMS配置表
CREATE TABLE IF NOT EXISTS sink_cms_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sink CMS配置表';

-- 13. Sink Event Store配置表
CREATE TABLE IF NOT EXISTS sink_event_store_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sink Event Store配置表';

-- 14. Join配置表
CREATE TABLE IF NOT EXISTS join_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    alert_config_id BIGINT UNSIGNED NOT NULL COMMENT '关联的Alert配置ID',
    join_type VARCHAR(100) COMMENT 'Join类型',
    join_config TEXT COMMENT 'Join配置，JSON格式',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    FOREIGN KEY (alert_config_id) REFERENCES alert_configurations(id) ON DELETE CASCADE,
    INDEX idx_alert_config_id (alert_config_id),
    INDEX idx_join_type (join_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Join配置表';

-- 添加外键约束
ALTER TABLE alert_configurations 
ADD CONSTRAINT fk_condition_config FOREIGN KEY (condition_config_id) REFERENCES condition_configurations(id) ON DELETE SET NULL,
ADD CONSTRAINT fk_group_config FOREIGN KEY (group_config_id) REFERENCES group_configurations(id) ON DELETE SET NULL,
ADD CONSTRAINT fk_policy_config FOREIGN KEY (policy_config_id) REFERENCES policy_configurations(id) ON DELETE SET NULL,
ADD CONSTRAINT fk_template_config FOREIGN KEY (template_config_id) REFERENCES template_configurations(id) ON DELETE SET NULL;

-- 创建索引优化查询性能
CREATE INDEX idx_alerts_composite ON alerts(status, create_time);
CREATE INDEX idx_config_composite ON alert_configurations(alert_id, `type`);
CREATE INDEX idx_schedule_composite ON alert_schedules(alert_id, `type`);
CREATE INDEX idx_tags_composite ON alert_tags(alert_id, `tag_type`);

-- 添加主表的外键约束
ALTER TABLE alerts 
ADD CONSTRAINT fk_alerts_configuration FOREIGN KEY (configuration_id) REFERENCES alert_configurations(id) ON DELETE SET NULL,
ADD CONSTRAINT fk_alerts_schedule FOREIGN KEY (schedule_id) REFERENCES alert_schedules(id) ON DELETE SET NULL;
