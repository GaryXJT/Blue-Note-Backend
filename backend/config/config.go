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
		viper.SetDefault("objectstorage.accesskey", "u6holj03")
		viper.SetDefault("objectstorage.secretkey", "bls9lcm5w4bt6qdw")
		viper.SetDefault("objectstorage.bucketname", "u6holj03-blue-note")
		viper.SetDefault("objectstorage.usessl", true) // 外部访问通常需要SSL
	} else {
		// 生产环境配置
		viper.SetDefault("mongodb.uri", "mongodb://root:ndpvkj55@dbconn.sealoshzh.site:30667/?directConnection=true")
		viper.SetDefault("mongodb.database", "bluenote")
		viper.SetDefault("redis.uri", "redis://default:6qcbbdwx@dbconn.sealoshzh.site:45216")
		
		// 生产环境对象存储配置 - 使用外部端点
		viper.SetDefault("objectstorage.endpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.internalendpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.externalendpoint", "objectstorageapi.hzh.sealos.run")
		viper.SetDefault("objectstorage.accesskey", "u6holj03")
		viper.SetDefault("objectstorage.secretkey", "bls9lcm5w4bt6qdw")
		viper.SetDefault("objectstorage.bucketname", "u6holj03-blue-note")
		viper.SetDefault("objectstorage.usessl", true)
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