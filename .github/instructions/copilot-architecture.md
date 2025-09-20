
---
applyTo: '**'
---
# Lockari Vault - Visão Geral da Arquitetura

O Lockari Vault é uma plataforma PaaS para gerenciamento seguro de Vaults, Segredos, Credenciais, Certificados e Chaves, baseada em microserviços e desenvolvida em Go. A seguir, estão os principais pontos da arquitetura, requisitos e padrões do projeto.

## Planos da Plataforma

- **Free**
- **Pro**
- **Enterprise**
- **On-Premises**

Cada plano possui recursos e limites específicos, detalhados na documentação de produto.

## Funcionalidades Principais

- Gerenciamento de Tenants, Grupos de Usuários, Vaults, Segredos, Credenciais, Certificados e Chaves OpenSSH.
- Tags para facilitar busca e organização (até 12 por vault/grupo).
- Permissões granulares por recurso, baseadas em grupos de usuários (semelhante ao Azure Key Vault).
- Auditoria e telemetria de todas as ações, acessível apenas a usuários autorizados.
- Application Tokens (JWT) para autenticação de aplicações e automações.
- Criptografia hierárquica: KEK (chave do vault) e DEK (chave dos dados), ambas protegidas por chaves superiores.
- Dashboard para visualização de recursos.
- Conformidade com PCI-DSS, ISO 27001, LGPD e outros requisitos de segurança.
- Suporte a recursos exclusivos para On-Premise via tags `go:build`.

## Stack Tecnológica

- **Linguagem:** Go
- **API:** Gin
- **Banco de Dados:** MongoDB (padrão) e Postgres (On-Premise)
- **Mensageria:** RabbitMQ
- **Cache:** Redis
- **Autenticação:** Firebase Authentication (padrão) e Auth0 (On-Premise)
- **Autorização:** OpenFGA
- **Tracing/Métricas:** OpenTelemetry
- **Logs:** Slog
- **Criptografia:** AES256GCM

## Estrutura de Diretórios

Segue o padrão [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
cmd/
  api/
  cli/
  worker/
internal/
  core/
    entity/
    repository/
    service/
  infrastructure/
    database/
      mongodb/
        mongodb.go
        adapter.go
      postgres/
        postgres.go
        adapter.go
    messaging/
    cache/
    auth/
    authz/
pkg/
  database/
    mongodb/
    postgres/
  cache/
  messaging/
  authclient/
    firebase_auth/
      firebase.go
      adapter.go
    auth0_auth/
      auth0.go
      adapter.go
  authzclient/
  telemetry/
  logger/
```

## Configuração

- Todas as configurações são feitas via arquivos YAML.
- Variáveis de ambiente:
  - `LOCKARI_CONFIG`: caminho do arquivo de configuração principal.
  - `LOCKARI_QUEUE`: caminho do arquivo de configuração de mensageria.
- O carregamento das configurações é centralizado em `configs/config.go`.

## API e Autenticação

- API RESTful via Gin, com documentação Swagger.
- Todas as requisições exigem autenticação:
  - Application Token: JWT enviado no header `X-TOKEN`.
  - Usuário: token do Firebase no header `Authorization`.
  - Requisições sem autenticação retornam 401 Unauthorized.
- Custom claims obrigatórios em tokens.

## Autorização

- Controle de acesso via OpenFGA.
- Usuário pode pertencer a múltiplos tenants; o tenant ativo é indicado no custom claim do token.

## Domínios e Entidades

### Tenant
- ID, Slug, Owner, Plano vinculado.
- Propriedades gerenciadas apenas por usuários autorizados.
- Criação automática de grupo e vault default (pode usar mensageria).

### Plano
- Cada plano define features e limites.
- O plano Free não limita número de vaults, secrets, credentials, certificates e keys, mas limita auditoria e usuários por tenant.

### Usuários
- Nos planos abaixo de Enterprise, usuários são convidados.
- Enterprise e On-Premise suportam autenticação SSO.

### Grupos
- Até 12 tags por grupo.
- Nome obrigatório, descrição opcional, múltiplos usuários.

### Vaults
- Até 12 tags por vault.
- Nome obrigatório, descrição opcional.
- Contém secrets, credentials, certificates, keys.
- Permissões associadas a grupos.

### Secrets / Key and Value
- Acesso via string, JSON ou YAML.
- Suporta tags e nome.

### Certificate
- Pode ser gerado ou importado como secret.
- Suporta tags e nome.

### Auditoria
- Apenas usuários autorizados podem acessar.
- Cada registro contém: ação (create, copy, download, ...), tenant, vault, secret, usuário, timestamp.

### Mensageria
- Gerenciada via arquivo YAML.
- Elementos obrigatórios: Exchanges, Binds, Queues, Routing Keys, Dead Letter Exchanges (DLX).
- Configuração via variável de ambiente `LOCKARI_QUEUE`.

### Database
- Suporte a MongoDB (padrão) e Postgres (On-Premise).
- Implementação desacoplada por diretório.

### Authentication
- Suporte a Firebase Authentication (padrão) e Auth0 (On-Premise).
- Implementação desacoplada por diretório.

### DTO (Data Trasnform Object)
Temos o DTO no diretório internal/core/dto/DOMAIN_NAME/DOMAIN_NAME.go

## Segurança
Precisamos ter o máximo possível de cuidado, para que possamos ter todos os controles de segurança possíveis.
Nós devemos seguir as melhores práticas para ter conformidade com PCI-DSS, ISO 27001, LGPD e outros requisitos de segurança.


---
Este documento serve como referência para desenvolvedores e ferramentas de automação (ex: Copilot), padronizando a arquitetura, requisitos e práticas do Lockari Vault.

