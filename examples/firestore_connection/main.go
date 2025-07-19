package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"lockari-api-application/pkg/database"
)

func main() {
	// Configuração do Firebase
	config := database.FirebaseConfig{
		ProjectID:             "your-project-id",
		ServiceAccountKeyPath: "path/to/your/service-account-key.json",
		// DatabaseURL pode ser omitido para usar o database padrão
	}

	// Inicializar conexão com Firestore
	db, err := database.InitializeFirebaseDB(config)
	if err != nil {
		log.Fatal("Falha ao conectar com Firestore:", err)
	}

	// Exemplo 1: Verificação simples de conexão (apenas verifica se o cliente não é nil)
	fmt.Println("=== Verificação Simples de Conexão ===")
	if db.IsConnected() {
		fmt.Println("✅ Cliente Firestore está inicializado")
	} else {
		fmt.Println("❌ Cliente Firestore não está inicializado")
	}

	// Exemplo 2: Verificação com ping real
	fmt.Println("\n=== Verificação com Ping Real ===")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if db.IsConnectedWithPing(ctx) {
		fmt.Println("✅ Firestore está conectado e respondendo")
	} else {
		fmt.Println("❌ Firestore não está respondendo")
	}

	// Exemplo 3: Ping detalhado com tratamento de erro
	fmt.Println("\n=== Ping Detalhado ===")
	if err := db.Ping(ctx); err != nil {
		fmt.Printf("❌ Erro no ping do Firestore: %v\n", err)
	} else {
		fmt.Println("✅ Ping do Firestore bem-sucedido")
	}

	// Exemplo 4: Verificação periódica de saúde
	fmt.Println("\n=== Verificação Periódica de Saúde ===")
	healthCheck(db)
}

// healthCheck demonstra como implementar uma verificação periódica de saúde
func healthCheck(db database.FirebaseDBInterface) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 3; i++ { // Apenas 3 iterações para o exemplo
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if !db.IsConnected() {
				fmt.Println("🔴 Health Check: Cliente não inicializado")
				cancel()
				continue
			}

			if err := db.Ping(ctx); err != nil {
				fmt.Printf("🟡 Health Check: Ping falhou - %v\n", err)
			} else {
				fmt.Println("🟢 Health Check: Firestore saudável")
			}

			cancel()
		}
	}
}
