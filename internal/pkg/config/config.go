package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	WeChat    WeChatConfig    `mapstructure:"wechat"`
	WxPay     WxPayConfig     `mapstructure:"wxpay"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

type WeChatConfig struct {
	AppID          string `mapstructure:"app_id"`
	AppSecret      string `mapstructure:"app_secret"`
	Token          string `mapstructure:"token"`
	EncodingAESKey string `mapstructure:"encoding_aes_key"`
}

// WxPayConfig 平台统一微信支付商户（用于 SaaS 租户订阅付费）
type WxPayConfig struct {
	AppID      string `mapstructure:"app_id"`
	MchID      string `mapstructure:"mch_id"`
	APIv3Key   string `mapstructure:"apiv3_key"`
	CertSerial string `mapstructure:"cert_serial"`
	NotifyURL  string `mapstructure:"notify_url"`
}

type StorageConfig struct {
	Type  string `mapstructure:"type"` // local / minio
	Local struct {
		Path    string `mapstructure:"path"`
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"local"`
	Minio struct {
		Endpoint   string `mapstructure:"endpoint"`    // 如 127.0.0.1:9000
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		Bucket     string `mapstructure:"bucket"`
		UseSSL     bool   `mapstructure:"use_ssl"`
		Region     string `mapstructure:"region"`
		BaseURL    string `mapstructure:"base_url"`    // 对外访问前缀，可为 CDN 或反代域名
		PublicRead bool   `mapstructure:"public_read"` // true: 返回直链；false: 返回预签名 URL
	} `mapstructure:"minio"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	QPS     int  `mapstructure:"qps"`
	Burst   int  `mapstructure:"burst"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
