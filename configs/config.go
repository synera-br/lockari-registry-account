package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	AppPathConfig string
)

type extensionConfig string

const (
	Yaml extensionConfig = "yaml"
	Json extensionConfig = "json"
	Toml extensionConfig = "toml"
)

func (e *extensionConfig) string() string {

	if e == nil {
		return ""
	}

	switch *e {
	case Yaml:
		return "yaml"
	case Json:
		return "json"
	case Toml:
		return "toml"
	default:
		return "yaml"
	}
}

type FileConfig struct {
	Extentsion     extensionConfig
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

type AppConfigField struct {
	Fields map[string]any
	Keys   []string
}

type AppConfig struct {
	PathConfigFile string `mapstructure:"path_config_file"`
	Paths          *string
	FileConfig     *FileConfig
	Settings       map[string]interface{}
	*AppConfigField
}

func (c *AppConfig) MarshalYAML() ([]byte, error) {
	b, err := yaml.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

func (c *AppConfig) Marshal() ([]byte, error) {
	b, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

func (c *AppConfig) MarshalYAMLField(field string) ([]byte, error) {
	if c.Settings[field] == nil {
		return nil, fmt.Errorf("field %s not found", field)
	}

	b, err := yaml.Marshal(c.Settings[field])
	if err != nil {
		return nil, err
	}

	return b, err
}

func (c *AppConfig) MarshalField(field string) ([]byte, error) {
	if c.Settings[field] == nil {
		return nil, fmt.Errorf("field %s not found", field)
	}

	b, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

func LoadConfig() (*AppConfig, error) {

	var extConfig extensionConfig
	if os.Getenv("CONFIG_EXTENSION") != "" {
		ex := os.Getenv("CONFIG_EXTENSION")
		extConfig = extensionConfig(ex)
	}

	err := setAppPathConfig(extConfig)
	if err != nil {
		return nil, fmt.Errorf("error to loading config: %v", err)
	}

	fc := FileConfig{
		ConfigPath:     AppPathConfig,
		Extentsion:     extConfig,
		FileName:       "config",
		ConfigFilePath: AppPathConfig,
	}

	viper.AddConfigPath(fc.ConfigPath)
	viper.SetConfigName(fc.FileName)
	viper.SetConfigType(fc.Extentsion.string())
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err.(viper.ConfigFileNotFoundError)
		}
		return nil, err
	}

	err = os.Setenv("JSON_CONFIG_PATH", fc.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	return &AppConfig{
		PathConfigFile: fc.ConfigFilePath,
		Paths:          &AppPathConfig,
		FileConfig:     &fc,
		Settings:       viper.AllSettings(),
		AppConfigField: &AppConfigField{
			Fields: viper.AllSettings(),
			Keys:   viper.AllKeys(),
		},
	}, err
}

func getHomeFile(ext extensionConfig) (string, string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if homeDir == "" {
			return "", "", err
		}
		return "", homeDir, err
	}

	defaultFile := filepath.Join(homeDir, ".appConfig", fmt.Sprintf("config.%s", ext.string()))
	if _, err := os.Stat(defaultFile); os.IsNotExist(err) {
		return "", homeDir, err
	}

	return defaultFile, homeDir, nil
}

func getCurrentFile(ext extensionConfig) (*string, error) {

	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	CurrentDirectory := filepath.Dir(execPath)
	defaultFile := filepath.Join(CurrentDirectory, "config.yaml")
	if _, err := os.Stat(defaultFile); os.IsNotExist(err) {
		return nil, err
	}

	return &defaultFile, nil
}

func setAppPathConfig(ext extensionConfig) error {

	AppPathConfig = os.Getenv("LOCKARI_CONFIG")
	if AppPathConfig != "" {
		return nil
	}

	f, _ := getCurrentFile(ext)
	if f != nil && *f != "" {
		AppPathConfig = *f
		return nil
	}

	ktools, home, _ := getHomeFile(ext)
	if ktools != "" {
		AppPathConfig = ktools
		return nil
	}

	return fmt.Errorf(`Config file not found. You can configure these options
1. Please set the LOCKARI_CONFIG environment variable
2. Create a config.yaml file at %s`, home+"/.lockari")

}
