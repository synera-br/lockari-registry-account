# 🗓️ Cronograma de Desenvolvimento Backend Lockari

| Fase | Etapa         | Funcionalidade                | Plano        | Entrega/Descrição                       |
|------|---------------|-------------------------------|--------------|------------------------------------------|
| 1    | Secrets       | CRUD de Secrets               | Free         | API para gerenciamento de secrets        |
| 2    | Key/Value     | CRUD de Key/Value             | Free         | API para gerenciamento de pares KV       |
| 3    | Secrets       | Compartilhamento de Secrets   | Pro          | Compartilhar secrets entre usuários      |
| 4    | Key/Value     | Compartilhamento de Key/Value | Pro          | Compartilhar pares KV entre usuários     |
| 5    | Certificate   | CRUD de Certificados          | Pro          | API para gerenciamento de certificados   |
| 6    | Certificate   | Compartilhamento de Certificados | Enterprise | Compartilhar certificados entre usuários |
| 7    | OpenSSH       | CRUD de Chaves SSH            | Enterprise   | API para gerenciamento de chaves SSH     |
| 8    | OpenSSH       | Compartilhamento de Chaves SSH| Enterprise   | Compartilhar chaves SSH entre usuários   |
| 9    | Connection    | CRUD de Connection Strings     | Enterprise   | API para gerenciamento de conexões       |
| 10   | Connection    | Compartilhamento de Connection Strings | Enterprise | Compartilhar conexões entre usuários     |

> Cada fase representa uma entrega funcional, completa e testada, alinhada ao plano correspondente.
> O cronograma pode ser ajustado conforme evolução do projeto e feedback dos usuários.

---

## 📋 Detalhamento das Fases

### Fase 1 – CRUD de Secrets (Free)
- **Componentes:**
  - API REST para secrets (endpoints: create, read, update, delete)
  - Serviço de criptografia para armazenamento seguro
  - Repositório para persistência (ex: Firestore, Redis)
  - Testes unitários e de integração
  - Documentação da API
- **Fluxo:** Usuário pode criar, consultar, atualizar e remover secrets via API. Todos os dados são criptografados em repouso.
- **OpenFGA Policies:**
  - `owner`: Usuário é proprietário do secret e pode realizar todas as operações.
  - **Condicionais:**
    - Acesso restrito ao usuário autenticado (owner).
    - Sem compartilhamento.

### Fase 2 – CRUD de Key/Value (Free)
- **Componentes:**
  - API REST para pares chave/valor
  - Serviço de validação de dados
  - Repositório para persistência
  - Testes unitários e integração
  - Documentação da API
- **Fluxo:** Usuário pode gerenciar pares chave/valor, com validação e segurança.
- **OpenFGA Policies:**
  - `owner`: Usuário é proprietário do par KV.
  - **Condicionais:**
    - Acesso restrito ao usuário autenticado.
    - Sem compartilhamento.

### Fase 3 – Compartilhamento de Secrets (Pro)
- **Componentes:**
  - API para compartilhamento seguro entre usuários
  - Serviço de controle de permissões
  - Auditoria de acessos
  - Testes unitários e integração
  - Documentação
- **Fluxo:** Secrets podem ser compartilhados com outros usuários, respeitando permissões e registrando auditoria.
- **OpenFGA Policies:**
  - `owner`, `shared`, `reader`, `writer`
  - **Condicionais:**
    - Compartilhamento por usuário ou grupo.
    - Permissões diferenciadas (leitura, escrita).
    - Auditoria de acessos.
    - Respeito ao plano Pro.

### Fase 4 – Compartilhamento de Key/Value (Pro)
- **Componentes:**
  - API para compartilhamento de pares KV
  - Serviço de controle de permissões
  - Auditoria
  - Testes e documentação
- **Fluxo:** Pares KV podem ser compartilhados entre usuários, com controle de acesso.
- **OpenFGA Policies:**
  - `owner`, `shared`, `reader`, `writer`
  - **Condicionais:**
    - Compartilhamento por usuário ou grupo.
    - Permissões diferenciadas.
    - Auditoria.
    - Respeito ao plano Pro.

### Fase 5 – CRUD de Certificados (Pro)
- **Componentes:**
  - API REST para certificados digitais
  - Serviço de validação e armazenamento seguro
  - Repositório
  - Testes e documentação
- **Fluxo:** Usuário pode gerenciar certificados digitais, incluindo upload, consulta e remoção.
- **OpenFGA Policies:**
  - `owner`
  - **Condicionais:**
    - Acesso restrito ao proprietário.
    - Respeito ao plano Pro.

### Fase 6 – Compartilhamento de Certificados (Enterprise)
- **Componentes:**
  - API para compartilhamento de certificados
  - Controle de permissões avançado
  - Auditoria
  - Testes e documentação
- **Fluxo:** Certificados podem ser compartilhados entre usuários e equipes, com rastreabilidade.
- **OpenFGA Policies:**
  - `owner`, `shared`, `team`, `reader`, `writer`
  - **Condicionais:**
    - Compartilhamento por usuário, grupo ou equipe.
    - Permissões avançadas.
    - Auditoria detalhada.
    - Respeito ao plano Enterprise.

### Fase 7 – CRUD de Chaves SSH (Enterprise)
- **Componentes:**
  - API REST para chaves SSH
  - Serviço de geração, armazenamento e validação
  - Repositório
  - Testes e documentação
- **Fluxo:** Usuário pode criar, importar, consultar e remover chaves SSH.
- **OpenFGA Policies:**
  - `owner`
  - **Condicionais:**
    - Acesso restrito ao proprietário.
    - Respeito ao plano Enterprise.

### Fase 8 – Compartilhamento de Chaves SSH (Enterprise)
- **Componentes:**
  - API para compartilhamento de chaves SSH
  - Permissões e auditoria
  - Testes e documentação
- **Fluxo:** Chaves SSH podem ser compartilhadas entre usuários, com controle de acesso.
- **OpenFGA Policies:**
  - `owner`, `shared`, `team`, `reader`, `writer`
  - **Condicionais:**
    - Compartilhamento por usuário, grupo ou equipe.
    - Permissões avançadas.
    - Auditoria detalhada.
    - Respeito ao plano Enterprise.

### Fase 9 – CRUD de Connection Strings (Enterprise)
- **Componentes:**
  - API REST para connection strings
  - Serviço de validação e armazenamento
  - Testes e documentação
- **Fluxo:** Usuário pode gerenciar strings de conexão de bancos e serviços.
- **OpenFGA Policies:**
  - `owner`
  - **Condicionais:**
    - Acesso restrito ao proprietário.
    - Respeito ao plano Enterprise.

### Fase 10 – Compartilhamento de Connection Strings (Enterprise)
- **Componentes:**
  - API para compartilhamento de connection strings
  - Permissões e auditoria
  - Testes e documentação
- **Fluxo:** Strings de conexão podem ser compartilhadas entre usuários e equipes.
- **OpenFGA Policies:**
  - `owner`, `shared`, `team`, `reader`, `writer`
  - **Condicionais:**
    - Compartilhamento por usuário, grupo ou equipe.
    - Permissões avançadas.
    - Auditoria detalhada.
    - Respeito ao plano Enterprise.

> Cada fase contempla: desenvolvimento dos endpoints, serviços internos, testes automatizados, documentação, integração contínua e validação de segurança.
