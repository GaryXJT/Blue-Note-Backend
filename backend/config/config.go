package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	MongoDB struct {
		URI      string
		Database string
	}
	Redis struct {
		URI      string
		Password string
	}
	JWT struct {
		Secret string
		Expire int // token过期时间（小时）
	}
	Server struct {
		Port int
	}
	ObjectStorage struct {
		Endpoint        string
		AccessKey       string
		SecretKey       string
		InternalEndpoint string
		ExternalEndpoint string
		BucketName      string
		UseSSL          bool
	}
	DefaultAvatar string `mapstructure:"default_avatar"`
}

var GlobalConfig Config

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("mongodb.uri", "mongodb://root:spdlw7qd@blue-note-db-mongodb.ns-h49hpg7e.svc:27017")
	viper.SetDefault("mongodb.database", "bluenote")
	viper.SetDefault("redis.uri", "redis://default:dfwfdwgx@blue-note-redis-db-redis.ns-h49hpg7e.svc:6379")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", 168)
	viper.SetDefault("server.port", 8080)
	
	// 对象存储默认配置
	viper.SetDefault("objectstorage.endpoint", "object-storage.objectstorage-system.svc.cluster.local")
	viper.SetDefault("objectstorage.internalendpoint", "object-storage.objectstorage-system.svc.cluster.local")
	viper.SetDefault("objectstorage.externalendpoint", "static-host-h49hpg7e-blue-note.sealoshzh.site")
	viper.SetDefault("objectstorage.accesskey", "h49hpg7e")
	viper.SetDefault("objectstorage.secretkey", "p4j8pq9ctnbhshpj")
	viper.SetDefault("objectstorage.bucketname", "h49hpg7e-blue-note")
	viper.SetDefault("objectstorage.usessl", false)
	viper.SetDefault("default_avatar", "default-avatar.jpg") // 使用正确的文件扩展名

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

func GetConfig() *Config {
	return &GlobalConfig
} 