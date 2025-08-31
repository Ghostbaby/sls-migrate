package config

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// SLSConfig SLS 配置
type SLSConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Project         string `json:"project"`
	LogStore        string `json:"log_store"`
}

// LoadSLSConfig 从环境变量加载 SLS 配置
func LoadSLSConfig() *SLSConfig {
	return &SLSConfig{
		Endpoint:        getEnv("SLS_ENDPOINT", "cn-qingdao.log.aliyuncs.com"),
		AccessKeyID:     getEnv("SLS_ACCESS_KEY_ID", ""),
		AccessKeySecret: getEnv("SLS_ACCESS_KEY_SECRET", ""),
		Project:         getEnv("SLS_PROJECT", ""),
		LogStore:        getEnv("SLS_LOG_STORE", ""),
	}
}

// CreateSLSClient 创建 SLS 客户端配置
func CreateSLSClient(cfg *SLSConfig) (*openapi.Config, error) {
	// 使用配置的 SLS 凭据
	cred, err := credential.NewCredential(&credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(cfg.AccessKeyID),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
		SecurityToken:   tea.String(""), // 明确指定不使用 STS token
	})
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: cred,
		Endpoint:   tea.String(cfg.Endpoint),
		// 禁用 ECS 角色获取
		Type: tea.String("access_key"),
	}

	return config, nil
}
