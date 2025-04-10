package config

import (
	"fmt"
	"os"

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
	Environment   string // 添加环境变量
}

var GlobalConfig Config

func Init() error {
	// 检测环境
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production" // 默认为生产环境
	}
	
	fmt.Printf("环境变量 APP_ENV = %s\n", env)
	
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	if env == "local" || env == "development" {
		// 本地开发环境配置
		viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
		viper.SetDefault("mongodb.database", "bluenote_dev")
		viper.SetDefault("redis.uri", "redis://localhost:6379")
		
		// 本地环境使用新的对象存储配置
		viper.SetDefault("objectstorage.endpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.internalendpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.externalendpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.accesskey", "duc9pmlo")
		viper.SetDefault("objectstorage.secretkey", "fzkzd9q2bbg4jp5j")
		viper.SetDefault("objectstorage.bucketname", "duc9pmlo-blue-note") // 保持原来的桶名，除非您有新的桶名
		viper.SetDefault("objectstorage.usessl", true) // 外部访问通常需要SSL
	} else {
		// 生产环境配置
		viper.SetDefault("mongodb.uri", "mongodb://root:spdlw7qd@blue-note-db-mongodb.ns-h49hpg7e.svc:27017")
		viper.SetDefault("mongodb.database", "bluenote")
		viper.SetDefault("redis.uri", "redis://default:dfwfdwgx@blue-note-redis-db-redis.ns-h49hpg7e.svc:6379")
		
		// 生产环境对象存储配置保持不变
		viper.SetDefault("objectstorage.endpoint", "object-storage.objectstorage-system.svc.cluster.local")
		viper.SetDefault("objectstorage.internalendpoint", "object-storage.objectstorage-system.svc.cluster.local")
		viper.SetDefault("objectstorage.externalendpoint", "static-host-h49hpg7e-blue-note.sealoshzh.site")
		viper.SetDefault("objectstorage.accesskey", "h49hpg7e")
		viper.SetDefault("objectstorage.secretkey", "p4j8pq9ctnbhshpj")
		viper.SetDefault("objectstorage.bucketname", "h49hpg7e-blue-note")
		viper.SetDefault("objectstorage.usessl", false)
	}
	
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", 168)
	viper.SetDefault("server.port", 8080)
	
	viper.SetDefault("default_avatar", "default-avatar.jpg") // 使用正确的文件扩展名
	viper.SetDefault("environment", env) // 设置环境变量

	if err := viper.ReadInConfig(); err != nil {
		// 配置文件读取失败时，使用默认配置
		fmt.Printf("读取配置文件失败: %v，将使用默认配置\n", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	// 确保环境变量被正确设置
	GlobalConfig.Environment = env
	
	fmt.Printf("当前运行环境: %s\n", GlobalConfig.Environment)
	fmt.Printf("MongoDB URI: %s\n", GlobalConfig.MongoDB.URI)
	fmt.Printf("Redis URI: %s\n", GlobalConfig.Redis.URI)

	return nil
}

func GetConfig() *Config {
	return &GlobalConfig
} 