package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	WeChat    WeChatConfig    `mapstructure:"wechat"`
	SMS       SMSConfig       `mapstructure:"sms"`
	WxPay     WxPayConfig     `mapstructure:"wxpay"`
	Storage   StorageConfig   `mapstructure:"storage"`
	AIImage   AIImageConfig   `mapstructure:"ai_image"`
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
	Driver               string `mapstructure:"driver"`
	DSN                  string `mapstructure:"dsn"`
	Host                 string `mapstructure:"host"`
	Port                 int    `mapstructure:"port"`
	User                 string `mapstructure:"user"`
	Password             string `mapstructure:"password"`
	Name                 string `mapstructure:"name"`
	SSLMode              string `mapstructure:"sslmode"`
	PreferSimpleProtocol bool   `mapstructure:"prefer_simple_protocol"`
	MaxOpenConns         int    `mapstructure:"max_open_conns"`
	MaxIdleConns         int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime      int    `mapstructure:"conn_max_lifetime"`
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
	AppID                string `mapstructure:"app_id"`
	AppSecret            string `mapstructure:"app_secret"`
	Token                string `mapstructure:"token"`
	EncodingAESKey       string `mapstructure:"encoding_aes_key"`
	MiniQRCodeEnvVersion string `mapstructure:"mini_qrcode_env_version"`
	MiniQRCodeCheckPath  bool   `mapstructure:"mini_qrcode_check_path"`
}

type SMSConfig struct {
	AliyunSignName string `mapstructure:"aliyun_sign_name"`
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
	Type  string `mapstructure:"type"` // local / minio / supabase
	Local struct {
		Path    string `mapstructure:"path"`
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"local"`
	Minio struct {
		Endpoint   string `mapstructure:"endpoint"` // 如 127.0.0.1:9000
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		Bucket     string `mapstructure:"bucket"`
		UseSSL     bool   `mapstructure:"use_ssl"`
		Region     string `mapstructure:"region"`
		BaseURL    string `mapstructure:"base_url"`    // 对外访问前缀，可为 CDN 或反代域名
		PublicRead bool   `mapstructure:"public_read"` // true: 返回直链；false: 返回预签名 URL
	} `mapstructure:"minio"`
	Supabase struct {
		ProjectURL       string `mapstructure:"project_url"`
		ServiceRoleKey   string `mapstructure:"service_role_key"`
		Bucket           string `mapstructure:"bucket"`
		PublicRead       bool   `mapstructure:"public_read"`
		BaseURL          string `mapstructure:"base_url"`
		SignedURLExpires int    `mapstructure:"signed_url_expires"`
		CreateBucket     bool   `mapstructure:"create_bucket"`
		Upsert           bool   `mapstructure:"upsert"`
	} `mapstructure:"supabase"`
}

type AIImageConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ServiceURL     string `mapstructure:"service_url"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
	MaxBytes       int    `mapstructure:"max_bytes"`
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
	_ = gotenv.Load()

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
	expandEnvironmentValues(&c)
	if err := applyEnvironmentOverrides(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func expandEnvironmentValues(config *Config) {
	config.Database.DSN = os.ExpandEnv(config.Database.DSN)
	config.Database.Host = os.ExpandEnv(config.Database.Host)
	config.Database.User = os.ExpandEnv(config.Database.User)
	config.Database.Password = os.ExpandEnv(config.Database.Password)
	config.Database.Name = os.ExpandEnv(config.Database.Name)
	config.Database.SSLMode = os.ExpandEnv(config.Database.SSLMode)
	config.WeChat.AppID = os.ExpandEnv(config.WeChat.AppID)
	config.WeChat.AppSecret = os.ExpandEnv(config.WeChat.AppSecret)
	config.WeChat.Token = os.ExpandEnv(config.WeChat.Token)
	config.WeChat.EncodingAESKey = os.ExpandEnv(config.WeChat.EncodingAESKey)
	config.WeChat.MiniQRCodeEnvVersion = os.ExpandEnv(config.WeChat.MiniQRCodeEnvVersion)
	config.SMS.AliyunSignName = os.ExpandEnv(config.SMS.AliyunSignName)

	config.Storage.Local.BaseURL = os.ExpandEnv(config.Storage.Local.BaseURL)
	config.Storage.Local.Path = os.ExpandEnv(config.Storage.Local.Path)
	config.Storage.Minio.Endpoint = os.ExpandEnv(config.Storage.Minio.Endpoint)
	config.Storage.Minio.AccessKey = os.ExpandEnv(config.Storage.Minio.AccessKey)
	config.Storage.Minio.SecretKey = os.ExpandEnv(config.Storage.Minio.SecretKey)
	config.Storage.Minio.Bucket = os.ExpandEnv(config.Storage.Minio.Bucket)
	config.Storage.Minio.Region = os.ExpandEnv(config.Storage.Minio.Region)
	config.Storage.Minio.BaseURL = os.ExpandEnv(config.Storage.Minio.BaseURL)
	config.Storage.Supabase.ProjectURL = os.ExpandEnv(config.Storage.Supabase.ProjectURL)
	config.Storage.Supabase.ServiceRoleKey = os.ExpandEnv(config.Storage.Supabase.ServiceRoleKey)
	config.Storage.Supabase.Bucket = os.ExpandEnv(config.Storage.Supabase.Bucket)
	config.Storage.Supabase.BaseURL = os.ExpandEnv(config.Storage.Supabase.BaseURL)
	config.AIImage.ServiceURL = os.ExpandEnv(config.AIImage.ServiceURL)
}

func applyEnvironmentOverrides(config *Config) error {
	setStringFromEnv(&config.App.Env, "APP_ENV", "ENV")
	setStringFromEnv(&config.App.Host, "APP_HOST")
	if err := setIntFromEnv(&config.App.Port, "APP_PORT", "PORT"); err != nil {
		return err
	}
	setStringFromEnv(&config.App.Mode, "APP_MODE", "GIN_MODE")

	setStringFromEnv(&config.Database.DSN, "SUPABASE_DB_DSN", "SUPABASE_DATABASE_URL", "DATABASE_DSN", "DATABASE_URL")
	setStringFromEnv(&config.Database.Host, "DATABASE_HOST", "DB_HOST")
	if err := setIntFromEnv(&config.Database.Port, "DATABASE_PORT", "DB_PORT"); err != nil {
		return err
	}
	setStringFromEnv(&config.Database.User, "DATABASE_USER", "DB_USER")
	setStringFromEnv(&config.Database.Password, "DATABASE_PASSWORD", "DB_PASSWORD")
	setStringFromEnv(&config.Database.Name, "DATABASE_NAME", "DB_NAME")
	setStringFromEnv(&config.Database.SSLMode, "DATABASE_SSLMODE", "DB_SSLMODE")
	if err := setBoolFromEnv(&config.Database.PreferSimpleProtocol, "DATABASE_PREFER_SIMPLE_PROTOCOL", "DB_PREFER_SIMPLE_PROTOCOL"); err != nil {
		return err
	}

	setStringFromEnv(&config.Redis.Host, "REDIS_HOST")
	if err := setIntFromEnv(&config.Redis.Port, "REDIS_PORT"); err != nil {
		return err
	}
	setStringFromEnv(&config.Redis.Password, "REDIS_PASSWORD")
	if err := setIntFromEnv(&config.Redis.DB, "REDIS_DB"); err != nil {
		return err
	}

	setStringFromEnv(&config.JWT.Secret, "JWT_SECRET")
	setStringFromEnv(&config.WeChat.AppID, "WECHAT_APP_ID", "WECHAT_APPID", "WX_APP_ID", "WXAPP_APP_ID")
	setStringFromEnv(&config.WeChat.AppSecret, "WECHAT_APP_SECRET", "WX_APP_SECRET", "WXAPP_APP_SECRET")
	setStringFromEnv(&config.WeChat.Token, "WECHAT_TOKEN", "WX_TOKEN")
	setStringFromEnv(&config.WeChat.EncodingAESKey, "WECHAT_ENCODING_AES_KEY", "WX_ENCODING_AES_KEY")
	setStringFromEnv(&config.WeChat.MiniQRCodeEnvVersion, "WECHAT_MINI_QRCODE_ENV_VERSION", "WX_MINI_QRCODE_ENV_VERSION")
	if err := setBoolFromEnv(&config.WeChat.MiniQRCodeCheckPath, "WECHAT_MINI_QRCODE_CHECK_PATH", "WX_MINI_QRCODE_CHECK_PATH"); err != nil {
		return err
	}
	setStringFromEnv(&config.SMS.AliyunSignName, "ALIYUN_SMS_SIGN_NAME", "SMS_SIGN_NAME")
	setStringFromEnv(&config.Storage.Type, "STORAGE_TYPE")

	setStringFromEnv(&config.Storage.Supabase.ProjectURL, "SUPABASE_URL", "SUPABASE_PROJECT_URL")
	setStringFromEnv(&config.Storage.Supabase.ServiceRoleKey, "SUPABASE_SERVICE_ROLE_KEY", "SUPABASE_SERVICE_KEY")
	setStringFromEnv(&config.Storage.Supabase.Bucket, "SUPABASE_STORAGE_BUCKET")
	setStringFromEnv(&config.Storage.Supabase.BaseURL, "SUPABASE_STORAGE_BASE_URL")
	if err := setBoolFromEnv(&config.Storage.Supabase.PublicRead, "SUPABASE_STORAGE_PUBLIC_READ"); err != nil {
		return err
	}
	if err := setBoolFromEnv(&config.Storage.Supabase.CreateBucket, "SUPABASE_STORAGE_CREATE_BUCKET"); err != nil {
		return err
	}
	if err := setBoolFromEnv(&config.Storage.Supabase.Upsert, "SUPABASE_STORAGE_UPSERT"); err != nil {
		return err
	}
	if err := setIntFromEnv(&config.Storage.Supabase.SignedURLExpires, "SUPABASE_STORAGE_SIGNED_URL_EXPIRES"); err != nil {
		return err
	}

	if err := setBoolFromEnv(&config.AIImage.Enabled, "AI_IMAGE_ENABLED"); err != nil {
		return err
	}
	setStringFromEnv(&config.AIImage.ServiceURL, "AI_IMAGE_SERVICE_URL")
	if err := setIntFromEnv(&config.AIImage.TimeoutSeconds, "AI_IMAGE_TIMEOUT_SECONDS"); err != nil {
		return err
	}
	if err := setIntFromEnv(&config.AIImage.MaxBytes, "AI_IMAGE_MAX_BYTES"); err != nil {
		return err
	}

	return nil
}

func setStringFromEnv(target *string, keys ...string) {
	for _, key := range keys {
		value, ok := lookupNonEmptyEnv(key)
		if !ok {
			continue
		}
		*target = value
		return
	}
}

func setIntFromEnv(target *int, keys ...string) error {
	for _, key := range keys {
		value, ok := lookupNonEmptyEnv(key)
		if !ok {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse env %s as int: %w", key, err)
		}
		*target = parsed
		return nil
	}
	return nil
}

func setBoolFromEnv(target *bool, keys ...string) error {
	for _, key := range keys {
		value, ok := lookupNonEmptyEnv(key)
		if !ok {
			continue
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse env %s as bool: %w", key, err)
		}
		*target = parsed
		return nil
	}
	return nil
}

func lookupNonEmptyEnv(key string) (string, bool) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	return value, value != ""
}
