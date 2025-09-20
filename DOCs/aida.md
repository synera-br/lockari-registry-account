# Golang + Viper: Gerenciando Configurações de Forma Simples e Escalável

## Atenção  
Você já precisou lidar com múltiplas configurações em uma aplicação Golang — variáveis de ambiente, arquivos YAML ou JSON — e acabou se perdendo entre `map[string]interface{}` e constantes espalhadas pelo código?  
Esse cenário é mais comum do que parece. E é aí que entra o **Viper**, uma das bibliotecas mais populares em Go para gerenciamento de configurações.  

---

## Interesse  
Recentemente, enfrentei dois desafios: configurar uma **CLI** e uma **API**. Apesar dos contextos diferentes, percebi que poderia unificar a forma de lidar com configurações.  
Com o Viper, ficou possível:  
- Ler variáveis de ambiente.  
- Carregar arquivos YAML/JSON.  
- Definir valores padrão de forma organizada.  
- Ter uma única base de configuração para múltiplos cenários.  

E neste artigo vou mostrar como construir uma **estrutura fortemente tipada**, simples de manter e escalável.  

---

## Desejo  
Imagine ter uma configuração clara, validada pelo compilador, com autocompletar na IDE e sem medo de quebrar código por causa de uma chave errada. É exatamente isso que conseguimos ao usar structs em conjunto com o Viper.  

### Pré-requisitos
- Go instalado.  
- Noções básicas de Go (structs, módulos, imports).  
- Noções básicas de YAML.  

### Instalação
```bash
go get github.com/spf13/viper
```

Estrutura de Diretórios
```
Copiar código
|- cmd/         # Entradas da aplicação (api/main.go, cli/main.go)
|- configs/     # Lógica de carregamento de configuração (config.go)
|- config.yaml  # Arquivo de configuração
```

Exemplo de config.yaml
```yaml
Copiar código
authentication:
  kind: "mytoken"
  encrypted: true
  encrypt_key: "0905f0a4dbb446419c82686e39d6afd66fd346a1bdf20771e58a38793fe317d7"
  endpoint: "my-endpoint.local"
  protocol: "https"

webserver:
  port: "8443"
  host: "0.0.0.0"
  name: "api"
  version: "v1"
  enabled_tls: true

otel:
  endpoint_type: "honeycomb"
  endpoint: "api.honeycomb.io:443"
  service_name: "lockari-vault-dev"
  headers:
    "x-honeycomb-team": "token-to-authentication"
    "x-honeycomb-dataset": "dataset-name"
  enabled_tracer: true
  enabled_logger: true
```

Arquivo configs/config.go
```go
Copiar código
package configs

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Authentication AuthConfig      `mapstructure:"authentication"`
	WebServer      WebServerConfig `mapstructure:"webserver"`
	Otel           OtelConfig      `mapstructure:"otel"`
}

type AuthConfig struct {
	Kind       string `mapstructure:"kind"`
	Encrypted  bool   `mapstructure:"encrypted"`
	EncryptKey string `mapstructure:"encrypt_key"`
	Endpoint   string `mapstructure:"endpoint"`
	Protocol   string `mapstructure:"protocol"`
}

type WebServerConfig struct {
	Port       string `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	Name       string `mapstructure:"name"`
	Version    string `mapstructure:"version"`
	EnabledTLS bool   `mapstructure:"enabled_tls"`
}

type OtelConfig struct {
	EndpointType  string            `mapstructure:"endpoint_type"`
	Endpoint      string            `mapstructure:"endpoint"`
	ServiceName   string            `mapstructure:"service_name"`
	Headers       map[string]string `mapstructure:"headers"`
	EnabledTracer bool              `mapstructure:"enabled_tracer"`
	EnabledLogger bool              `mapstructure:"enabled_logger"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("arquivo de configuração não encontrado: %w", err)
		}
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configurações: %w", err)
	}

	return &cfg, nil
}
Cenário 1: CLI
go
Copiar código
package main

import (
	"log"
	"fmt"
	"meu-projeto/configs"
)

func initializeAuth(cfg configs.AuthConfig) {
	fmt.Println("Inicializando autenticação com o kind:", cfg.Kind)
}

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	initializeAuth(cfg.Authentication)
	fmt.Println("CLI configurada com sucesso!")
}
Cenário 2: API
go
Copiar código
package main

import (
	"log"
	"fmt"
	"meu-projeto/configs"
)

func initializeOtel(cfg configs.OtelConfig) error {
	fmt.Printf("Inicializando Otel para o serviço: %s\n", cfg.ServiceName)
	return nil
}

func initializeWebServer(wsCfg configs.WebServerConfig, otelCfg configs.OtelConfig) error {
	fmt.Printf("Inicializando WebServer '%s' em %s:%s\n", wsCfg.Name, wsCfg.Host, wsCfg.Port)
	return nil
}

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Erro ao inicializar o config: %v", err)
	}

	if err := initializeOtel(cfg.Otel); err != nil {
		log.Fatalf("Erro ao inicializar o otel: %v", err)
	}

	if err := initializeWebServer(cfg.WebServer, cfg.Otel); err != nil {
		log.Fatalf("Erro ao inicializar o webserver: %v", err)
	}

	fmt.Println("API configurada e pronta para rodar!")
}
Ação
Agora é a sua vez!

Instale o Viper no seu projeto.

Crie seu config.yaml com as seções necessárias.

Defina suas structs em Go para mapear essas configurações.

Reaproveite o mesmo carregador em diferentes cenários (CLI, API, workers etc).

💡 Com essa base, você terá configurações seguras, organizadas e escaláveis em Go.
Nos próximos artigos, vou mostrar como integrar essa abordagem com Cobra-CLI e Kubernetes (ConfigMaps/Secrets).

yaml
Copiar código

---
