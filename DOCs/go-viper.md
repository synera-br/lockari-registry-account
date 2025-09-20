# Golang + Viper: Gerenciando Configurações de Forma Simples e Escalável

## Introdução  
Recentemente, me deparei com uma situação, de ter que ficar mudando meu Golang package de configuração (configs), porém, gostaria de deixar mais agnostico a qualquer modelo, sendo que meu config deveria carregar e tratar as configs.

Após algumas pesquisas, optei pelo uso do **[Viper](https://github.com/spf13/viper)**, uma das bibliotecas mais populares em Go para gerenciamento de configurações. Com ele, é possível ler variáveis de ambiente, arquivos YAML/JSON/TOML, definir valores padrão e até integrar facilmente com sistemas externos.  

Neste artigo, vamos criar uma aplicação simples em Golang utilizando o Viper para carregar e acessar configurações. O objetivo é apresentar uma base funcional e genérica, que pode ser aplicada tanto em CLIs quanto em APIs.


## Pré-requisitos
- Go instalado (versão mínima recomendada).  
- Noções básicas de Go (módulos, imports).  
- Noções básicas em YAML.  


## Estrutura de Diretórios
Minha arquitetura de diretórios segue o padrão [golang-standards/project-layout](https://github.com/golang-standards/project-layout).  

Para o nosso exemplo, vamos trabalhar apenas com três diretórios:  

```
|- cmd     # inicialização da aplicação
|- configs # diretório que contém nosso arquivo principal de configuração
|- pkg     # pacotes auxiliares
```


## Configurando uma API com Viper
Nesse exemplo, vamos carregar o nosso arquivo de configuração, os parametros do GIN (Web Framework) e do OpenTelemetry.

No nosso package **configs/config.go**, temos a possibilidade de definir o tipo de extensão **YAML/JSON/TOML** e o path de onde estará nosso arquivo de configuração.

Vamos agora entender o nosso package **configs/config.go** e ver como ele está estruturado.


## Package Configs
Nessa etapa, vou detalhar cada função do package e posteriormente veremos o arquivo completo.

```go
// caminho do nosso arquivo de configuração
var AppPathConfig    string

// extensionConfig define a extensão do nosso arquivo de configuração
type extensionConfig string

// definie os tipos de extensões permitidas
const (
    Yaml extensionConfig = "yaml"
	Json extensionConfig = "json"
	Toml extensionConfig = "toml"
)

// retorna a extensão do nosso arquivo de configuração
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

// FileConfig define a estrutura e configurações do nosso arquivo
// Extentsion é a extensão do nosso arquivo de configuração
// FileName é o nome do nosso arquivo de configuração
// ConfigPath é o path do nosso arquivo de configuração
// ConfigFilePath é o path completo do nosso arquivo de configuração
type FileConfig struct {
	Extentsion     extensionConfig
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

// AppConfigField mapa do arquivo de configuração
// Fields é o arquivo de configuração por completo
// Keys são todas as chaves que temos do arquivo de configuração
type AppConfigField struct {
	Fields map[string]any
	Keys   []string
}
```

Nessa etapa, teremos a estrutura que vamos montar e retornar ao carregar as configurações
```go 
// AppConfig estrutura com as configurações carregadas
type AppConfig struct {
	PathConfigFile string 
	Paths          *string
	FileConfig     *FileConfig
	Settings       map[string]interface{}
	*AppConfigField
}

// MarshalYAML retorna o arquivo de configuração em YAML
func (c *AppConfig) MarshalYAML() ([]byte, error) {
	b, err := yaml.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

// Marshal retorna o arquivo de configuração em JSON
func (c *AppConfig) Marshal() ([]byte, error) {
	b, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

// MarshalYAMLField retorna uma chave especifica no formato YAML
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

// MarshalField retorna uma chave especifica no formato JSON
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
```

Agora, criaremos 3 funções,que são de onde o arquivo pode ser carregado. Explicaremos cada uma delas
```go
// getHomeFile retorna o path do arquivo de configuração e o path do diretório do usuário
func getHomeFile(ext extensionConfig) (string, string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if homeDir == "" {
			return "", "", err
		}
		return "", homeDir, err
	}

	defaultFile := filepath.Join(homeDir, ".appConfig",fmt.Sprintf("config.%s", ext.string()))
	if _, err := os.Stat(defaultFile); os.IsNotExist(err) {
		return "", homeDir, err
	}

	return defaultFile, homeDir, nil
}

// getCurrentFile carrega as configurações a partir do path que está sendo executado a APP
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

// setAppPathConfig carrega as configurações a partir do path que vem de uma variável de ambiente ou das funções acima
func setAppPathConfig(ext extensionConfig) error {

	AppPathConfig = os.Getenv("APP_CONFIG")
	if AppPathConfig != "" {
		return nil
	}

	f, _ := getCurrentFile()
	if f != nil && *f != "" {
		AppPathConfig = *f
		return nil
	}

	ktools, home, _ := getHomeFile()
	if ktools != "" {
		AppPathConfig = ktools
		return nil
	}

	return fmt.Errorf(`Config file not found. You can configure these options
1. Please set the APP_CONFIG environment variable
2. Create a config.yaml file at %s`, home+"/.appConfig")

}
```

Por último, iremos carregar as funções mencionadas anteriormente e retornar as nossas configurações
```go

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
```

Com isso, finalizamos nosso package config. O resultado final é:
```go
package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// caminho do nosso arquivo de configuração
var AppPathConfig    string

// extensionConfig define a extensão do nosso arquivo de configuração
type extensionConfig string

// definie os tipos de extensões permitidas
const (
    Yaml extensionConfig = "yaml"
	Json extensionConfig = "json"
	Toml extensionConfig = "toml"
)

// retorna a extensão do nosso arquivo de configuração
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

// FileConfig define a estrutura e configurações do nosso arquivo
// Extentsion é a extensão do nosso arquivo de configuração
// FileName é o nome do nosso arquivo de configuração
// ConfigPath é o path do nosso arquivo de configuração
// ConfigFilePath é o path completo do nosso arquivo de configuração
type FileConfig struct {
	Extentsion     extensionConfig
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

// AppConfigField mapa do arquivo de configuração
// Fields é o arquivo de configuração por completo
// Keys são todas as chaves que temos do arquivo de configuração
type AppConfigField struct {
	Fields map[string]any
	Keys   []string
}

// AppConfig estrutura com as configurações carregadas
type AppConfig struct {
	PathConfigFile string 
	Paths          *string
	FileConfig     *FileConfig
	Settings       map[string]interface{}
	*AppConfigField
}

// MarshalYAML retorna o arquivo de configuração em YAML
func (c *AppConfig) MarshalYAML() ([]byte, error) {
	b, err := yaml.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

// Marshal retorna o arquivo de configuração em JSON
func (c *AppConfig) Marshal() ([]byte, error) {
	b, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, err
	}

	return b, err
}

// MarshalYAMLField retorna uma chave especifica no formato YAML
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

// MarshalField retorna uma chave especifica no formato JSON
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

// getHomeFile retorna o path do arquivo de configuração e o path do diretório do usuário
func getHomeFile(ext extensionConfig) (string, string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if homeDir == "" {
			return "", "", err
		}
		return "", homeDir, err
	}

	defaultFile := filepath.Join(homeDir, ".appConfig",fmt.Sprintf("config.%s", ext.string()))
	if _, err := os.Stat(defaultFile); os.IsNotExist(err) {
		return "", homeDir, err
	}

	return defaultFile, homeDir, nil
}

// getCurrentFile carrega as configurações a partir do path que está sendo executado a APP
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

// setAppPathConfig carrega as configurações a partir do path que vem de uma variável de ambiente ou das funções acima
func setAppPathConfig(ext extensionConfig) error {

	AppPathConfig = os.Getenv("APP_CONFIG")
	if AppPathConfig != "" {
		return nil
	}

	f, _ := getCurrentFile()
	if f != nil && *f != "" {
		AppPathConfig = *f
		return nil
	}

	ktools, home, _ := getHomeFile()
	if ktools != "" {
		AppPathConfig = ktools
		return nil
	}

	return fmt.Errorf(`Config file not found. You can configure these options
1. Please set the APP_CONFIG environment variable
2. Create a config.yaml file at %s`, home+"/.appConfig")
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

```

## Arquivo de configuração
Vamos ver como seria o nosso arquivo de configuração da app **config.yaml**.

Segue o exemplo da configuração:
```yaml
telemetry:
  endpoint_type: honeycomb
  endpoint: api.honeycomb.io:443
  service_name: myapp-name
  headers:
    "x-honeycomb-team": "mytoken"
    "x-honeycomb-dataset": "mydataset-name"
  enabled_tracer: true
  enabled_logger: true

webserver:
  port: "8080"
  host: "0.0.0.0"
  name: "api"
  version: "v1"
  httpv2: true

database:
...

```

Esse seria um exemplo de configuração, onde podemos colocar as informações de banco de dados, mensageria, ....

## Inicializando a APP
Como já temos nosso package configs, configurado e o nosso arquivo config.yaml, podemos ver como inicializar as configurações

Vamos colcoar as configs no arquivo main.go
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"myapp/configs"
)

func main() {

	// INITIALIZE CONFIG
	cfg, err := initializeConfig()
	if err != nil {
		panic(err)
	}

		// INITIALIZE OPENTELEMETRY
	b, err := cfg.MarshalField("telemetry")
	if err != nil {
		log.Fatalf("error to marshal config: %v \n", err)
	}
	obs, cleanup, err := initializeOtel(b)
	if err != nil {
		log.Fatalf("error to initialize telemetry: %v \n", err)
	}

	if obs == nil {
		log.Fatalf("error to initialize telemetry: %v \n", err)
	}

	ctx := context.Background()
	ctx, span := obs.Span(ctx, "main")
	defer span.End()
	defer obs.CleanUp()
	defer cleanup()

	b, err := cfg.MarshalField("webserver")
	if err != nil {
		log.Fatalf("error to marshal config: %v \n", err)
	}
	router, err := initializeWebServer(b)
	if err != nil {
		log.Fatalf("error to initialize web server: %v \n", err)
	}

	if router == nil {
		log.Fatalf("warning: web server is nil \n")
	}

	if err := router.Run(); err != nil {
		log.Fatalf("error to execute webserver %w",err)
	}
	
}

func initializeConfig() (*configs.AppConfig, error) {

	cfg, err := configs.LoadConfig()
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	return cfg, nil
}

func initializeOtel(fields []byte) (telemetry.OtelObservability, func(), error) {

	if fields == nil {
		return nil, nil, fmt.Errorf("fields on initializeOtel is nil")
	}

	var config telemetry.OtelConfig
	err = json.Unmarshal(fields, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	// PROBLEMA 4: Usar InitObservability ao invés de usar config diretamente
	obs, cleanup, err := telemetry.InitObservability(&config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	if obs == nil {
		return nil, nil, fmt.Errorf("telemetry is nil")
	}

	return obs, cleanup, nil
}

func initializeWebServer( fields []byte)(webserver.Router, error){
	if fields == nil {
		return nil,  fmt.Errorf("fields on initializeWebServer is nil")
	}

	var config webserver.InitWebServer
	err = json.Unmarshal(fields, &config)
	if err != nil {
		return nil,  fmt.Errorf("failed to unmarshal config: %v", err)
	}

	router, err := webserver.InitWebServer(&config)
	if err != nil {
		return nil,  fmt.Errorf("failed to initialize webserver: %w", err)
	}

	if router == nil {
		return nil,  fmt.Errorf("webserver is nil")
	}

	return router, nil
}
```

Nesse modelo, consegui ter o meu package config de forma generica e em cada package que inicializo, envio as configurações respectivas do package.

Nesse formato, não me preocupo mais com meu **configs/config.go**, conforme a evolução da APP, apenas foco no arquivo **config.yaml** e na estrutura da dados que cada package, deva receber.

Me conta, como você usa a sua estrutura de configurações no Golang!  Constuma utilizar o Viper?