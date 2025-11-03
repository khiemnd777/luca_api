package config

import "time"

type ProjectConfig struct {
	Name          string `mapstructure:"name"`
	BaseAPIPrefix string `mapstructure:"baseapiprefix"`
	Version       string `mapstructure:"version"`
}

type ServerConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Route       string `mapstructure:"route"`
	BodyLimitMB int    `mapstructure:"body_limit_mb"`
}

type AuthConfig struct {
	Secret            string `yaml:"secret"`
	InternalAuthToken string `yaml:"internalauthtoken"`
	InternalLogToken  string `yaml:"internallogtoken"`
}

type DatabaseConfig struct {
	Provider    string         `yaml:"provider"` // "postgres", "mongodb", etc.
	AutoMigrate bool           `mapstructure:"automigrate" yaml:"automigrate"`
	Postgres    PostgresConfig `yaml:"postgres"`
	MongoDB     MongoConfig    `yaml:"mongodb"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type RedisInstanceConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	IsCluster bool   `yaml:"iscluster"`
	UseTLS    bool   `yaml:"usetls"`
}

type CacheTTL struct {
	Short  time.Duration `mapstructure:"short"`
	Medium time.Duration `mapstructure:"medium"`
	Long   time.Duration `mapstructure:"long"`
	Static time.Duration `mapstructure:"static"`
}

type CacheConfig struct {
	TTL CacheTTL `mapstructure:"ttl"`
}

type RedisConfig struct {
	Instances map[string]RedisInstanceConfig `yaml:"instances"`
}

type CircuitBreakerConfig struct {
	Interval            time.Duration `yaml:"interval"`
	Timeout             time.Duration `yaml:"timeout"`
	ConsecutiveFailures int           `yaml:"consecutivefailures"`
}
type RetryConfig struct {
	MaxAttempts int           `yaml:"maxattempts"`
	Delay       time.Duration `yaml:"delay"`
}

type AppConfig struct {
	Project        ProjectConfig        `mapstructure:"project"`
	Server         ServerConfig         `mapstructure:"server"`
	Auth           AuthConfig           `yaml:"auth"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Redis          RedisConfig          `mapstructure:"redis"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuitbreaker"`
	Retry          RetryConfig          `yaml:"retry"`
}
