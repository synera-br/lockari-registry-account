package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	cryptserver "lockari-api-application/pkg/crypt/crypt_server"
)

const (
	// Exemplo de chave base64 para teste (32 bytes = 256 bits)
	// Em produção, use uma chave segura e não a coloque no código
	defaultTestKey = "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="
)

func main() {
	// Definir flags de linha de comando
	var (
		mode        = flag.String("mode", "decrypt", "Modo de operação: 'encrypt' ou 'decrypt'")
		cryptMode   = flag.String("crypt", "CBC", "Modo de criptografia: 'CBC' ou 'GCM'")
		key         = flag.String("key", defaultTestKey, "Chave de criptografia em Base64")
		payload     = flag.String("payload", "", "Payload para criptografar/descriptografar")
		interactive = flag.Bool("i", false, "Modo interativo")
		help        = flag.Bool("h", false, "Mostrar ajuda")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *interactive {
		runInteractiveMode()
		return
	}

	if *payload == "" {
		fmt.Println("Erro: payload é obrigatório")
		showHelp()
		os.Exit(1)
	}

	// Validar modo de criptografia
	if *cryptMode != "CBC" && *cryptMode != "GCM" {
		fmt.Printf("Erro: modo de criptografia inválido: %s (deve ser CBC ou GCM)\n", *cryptMode)
		showHelp()
		os.Exit(1)
	}

	// Inicializar CryptData com modo específico
	var cryptData cryptserver.CryptDataInterface
	var err error

	if *cryptMode == "CBC" {
		cryptData, err = cryptserver.InicializationCryptData(key)
	} else {
		cryptData, err = cryptserver.InicializationCryptDataWithMode(key, *cryptMode)
	}

	if err != nil {
		log.Fatalf("Erro ao inicializar CryptData: %v", err)
	}

	fmt.Printf("Modo de criptografia: %s\n", *cryptMode)

	switch *mode {
	case "encrypt":
		result, err := encryptData(cryptData, *payload, *cryptMode)
		if err != nil {
			log.Fatalf("Erro na criptografia: %v", err)
		}
		fmt.Printf("Payload criptografado (%s): %s\n", *cryptMode, result)

	case "decrypt":
		result, err := decryptData(cryptData, *payload)
		if err != nil {
			log.Fatalf("Erro na descriptografia: %v", err)
		}
		fmt.Printf("Payload descriptografado: %s\n", result)

	default:
		fmt.Printf("Modo inválido: %s. Use 'encrypt' ou 'decrypt'\n", *mode)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println(`
CLI para testar criptografia/descriptografia de payloads

Uso:
  go run main.go [flags]

Flags:
  -mode string
        Modo de operação: 'encrypt' ou 'decrypt' (default "decrypt")
  -crypt string
        Modo de criptografia: 'CBC' ou 'GCM' (default "CBC")
  -key string
        Chave de criptografia em Base64 (default usa chave de teste)
  -payload string
        Payload para criptografar/descriptografar (obrigatório)
  -i    Modo interativo
  -h    Mostrar esta ajuda

Exemplos:
  # Descriptografar payload do frontend (CBC)
  go run main.go -mode decrypt -payload "Base64PayloadAqui"

  # Criptografar dados JSON usando GCM
  go run main.go -mode encrypt -crypt GCM -payload '{"user":"test","data":"value"}'

  # Teste CBC vs GCM
  go run main.go -mode encrypt -crypt CBC -payload "test data"
  go run main.go -mode encrypt -crypt GCM -payload "test data"

  # Modo interativo
  go run main.go -i

  # Usar chave personalizada com GCM
  go run main.go -crypt GCM -key "SuaChaveBase64Aqui" -mode decrypt -payload "PayloadAqui"
`)
}

func encryptData(cryptData cryptserver.CryptDataInterface, payload string, mode string) (string, error) {
	// Converter payload string para bytes
	jsonData := []byte(payload)

	// Criptografar usando o modo especificado
	var encrypted string
	var err error

	if mode == "GCM" {
		encrypted, err = cryptData.EncryptPayloadWithMode(jsonData, "GCM")
	} else {
		encrypted, err = cryptData.EncryptPayload(jsonData) // Usa modo padrão (CBC)
	}

	if err != nil {
		return "", fmt.Errorf("falha na criptografia %s: %w", mode, err)
	}

	return encrypted, nil
}

func decryptData(cryptData cryptserver.CryptDataInterface, payload string) (string, error) {
	// Descriptografar
	decrypted, err := cryptData.PayloadData(payload)
	if err != nil {
		return "", fmt.Errorf("falha na descriptografia: %w", err)
	}

	// Tentar fazer parse como JSON para formatação
	var jsonData interface{}
	if err := json.Unmarshal(decrypted, &jsonData); err == nil {
		// É um JSON válido, formatar bonito
		formatted, err := json.MarshalIndent(jsonData, "", "  ")
		if err == nil {
			return string(formatted), nil
		}
	}

	// Se não for JSON válido, retornar como string
	return string(decrypted), nil
}

func runInteractiveMode() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Modo Interativo de Teste de Criptografia ===")
	fmt.Println("Digite 'exit' para sair")

	// Solicitar chave
	fmt.Print("Digite a chave Base64 (Enter para usar chave padrão): ")
	keyInput, _ := reader.ReadString('\n')
	keyInput = strings.TrimSpace(keyInput)

	key := defaultTestKey
	if keyInput != "" {
		key = keyInput
	}

	// Solicitar modo de criptografia
	fmt.Print("Digite o modo de criptografia [CBC/GCM] (Enter para CBC): ")
	cryptModeInput, _ := reader.ReadString('\n')
	cryptModeInput = strings.TrimSpace(strings.ToUpper(cryptModeInput))

	cryptMode := "CBC"
	if cryptModeInput == "GCM" {
		cryptMode = "GCM"
	}

	// Inicializar CryptData com modo específico
	var cryptData cryptserver.CryptDataInterface
	var err error

	if cryptMode == "CBC" {
		cryptData, err = cryptserver.InicializationCryptData(&key)
	} else {
		cryptData, err = cryptserver.InicializationCryptDataWithMode(&key, cryptMode)
	}

	if err != nil {
		log.Fatalf("Erro ao inicializar CryptData: %v", err)
	}

	fmt.Printf("Usando chave: %s\n", key)
	fmt.Printf("Modo de criptografia: %s\n\n", cryptMode)

	for {
		fmt.Print("Escolha o modo (encrypt/decrypt): ")
		modeInput, _ := reader.ReadString('\n')
		modeInput = strings.TrimSpace(strings.ToLower(modeInput))

		if modeInput == "exit" {
			fmt.Println("Saindo...")
			break
		}

		if modeInput != "encrypt" && modeInput != "decrypt" {
			fmt.Println("Modo inválido. Use 'encrypt' ou 'decrypt'")
			continue
		}

		fmt.Print("Digite o payload: ")
		payloadInput, _ := reader.ReadString('\n')
		payloadInput = strings.TrimSpace(payloadInput)

		if payloadInput == "exit" {
			fmt.Println("Saindo...")
			break
		}

		if payloadInput == "" {
			fmt.Println("Payload não pode estar vazio")
			continue
		}

		switch modeInput {
		case "encrypt":
			result, err := encryptData(cryptData, payloadInput, cryptMode)
			if err != nil {
				fmt.Printf("Erro na criptografia %s: %v\n", cryptMode, err)
			} else {
				fmt.Printf("Resultado criptografado (%s):\n%s\n", cryptMode, result)
			}

		case "decrypt":
			result, err := decryptData(cryptData, payloadInput)
			if err != nil {
				fmt.Printf("Erro na descriptografia: %v\n", err)
			} else {
				fmt.Printf("Resultado descriptografado:\n%s\n", result)
			}
		}

		fmt.Println(strings.Repeat("-", 50))
	}
}
