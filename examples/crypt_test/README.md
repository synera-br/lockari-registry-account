# CLI de Teste para Criptografia/Descriptografia

Este é um utilitário CLI para testar a funcionalidade de criptografia/descriptografia de payloads recebidos do frontend.

## Pré-requisitos

- Go 1.23 ou superior
- Acesso ao módulo `lockari-api-application`

## Instalação

```bash
cd /home/carlos.tomelin/codes/tomelin/lockari/lockari-backend-app/examples/crypt_test
go mod tidy
```

## Uso

### Mostrar Ajuda

```bash
go run main.go -h
```

### Criptografar Dados

```bash
# Criptografar JSON
go run main.go -mode encrypt -payload '{"user":"test","data":"value"}'

# Criptografar texto simples
go run main.go -mode encrypt -payload "Hello World"
```

### Descriptografar Dados

```bash
# Descriptografar payload do frontend
go run main.go -mode decrypt -payload "Base64PayloadAqui"

# Exemplo com payload criptografado
go run main.go -mode decrypt -payload "zO/kbBgRG8jyk1DrM0PhpmKfFw+RiLCVskwjRaYNLSto5PyBH3MWyxvgBfc8175W"
```

### Usar Chave Personalizada

```bash
# Usar sua própria chave Base64
go run main.go -key "SuaChaveBase64Aqui" -mode decrypt -payload "PayloadAqui"
```

### Modo Interativo

```bash
# Modo interativo para múltiplos testes
go run main.go -i
```

No modo interativo, você pode:
1. Escolher usar a chave padrão ou inserir sua própria chave
2. Alternar entre criptografia e descriptografia
3. Testar múltiplos payloads sem reiniciar o programa
4. Digite 'exit' para sair

## Exemplos de Uso

### Teste Completo (Criptografar e Descriptografar)

```bash
# 1. Criptografar um payload JSON
ENCRYPTED=$(go run main.go -mode encrypt -payload '{"user":"admin","token":"abc123"}')
echo "Criptografado: $ENCRYPTED"

# 2. Descriptografar o resultado
go run main.go -mode decrypt -payload "$ENCRYPTED"
```

### Teste com Payload do Frontend

Se você receber um payload do frontend como:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Use:
```bash
go run main.go -mode decrypt -payload "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Chave Padrão

A chave padrão para teste é: `MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=`

**⚠️ IMPORTANTE**: Esta chave é apenas para teste. Em produção, use uma chave segura e não a coloque no código.

## Estrutura do Projeto

```
examples/crypt_test/
├── main.go      # Código principal do CLI
├── go.mod       # Módulo Go
└── README.md    # Este arquivo
```

## Debugging

O programa inclui logs detalhados durante a descriptografia, mostrando:
- Chave sendo usada
- Tamanho dos dados decodificados
- Processo de descriptografia

Para ver os logs, execute normalmente - eles aparecerão no stderr.

## Solução de Problemas

### Erro: "invalid AES key size"
- Certifique-se que a chave Base64 decodifica para 16, 24 ou 32 bytes
- Use a chave padrão ou gere uma nova chave válida

### Erro: "failed to decode base64"
- Verifique se o payload está em formato Base64 válido
- Certifique-se que não há espaços ou caracteres extras

### Erro: "failed to unpad data"
- Provavelmente a chave está incorreta
- Verifique se está usando a mesma chave que foi usada para criptografar

## Exemplo de Sessão Interativa

```
=== Modo Interativo de Teste de Criptografia ===
Digite 'exit' para sair
Digite a chave Base64 (Enter para usar chave padrão): 
Usando chave: MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=

Escolha o modo (encrypt/decrypt): encrypt
Digite o payload: {"test": "data"}
Resultado criptografado:
abcd1234...

Escolha o modo (encrypt/decrypt): decrypt
Digite o payload: abcd1234...
Resultado descriptografado:
{
  "test": "data"
}

Escolha o modo (encrypt/decrypt): exit
Saindo...
```
