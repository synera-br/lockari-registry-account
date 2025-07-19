package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"lockari-api-application/internal/core/entity/auth"
)

func main() {
	fmt.Println("=== Auth Entity Example ===")

	// Create user and client info
	user := auth.User{
		Uid:   "user-123",
		Email: "john.doe@example.com",
		Name:  "John Doe",
		Plan:  "premium",
	}

	clientInfo := auth.Client{
		IpAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	// Example 1: Create a Login event using factory method
	fmt.Println("\n1. Login Event:")
	loginEvent := auth.NewLogin(user, clientInfo)

	if err := loginEvent.IsValid(); err != nil {
		log.Printf("Login validation error: %v", err)
	} else {
		fmt.Printf("✓ Login event is valid\n")
		fmt.Printf("  Event Type: %s\n", loginEvent.GetEventType().GetEventType())
		fmt.Printf("  User: %s (%s)\n", loginEvent.GetUser().Name, loginEvent.GetUser().Email)
		fmt.Printf("  Timestamp: %s\n", loginEvent.GetTimestamp().Format(time.RFC3339))
	}

	// Example 2: Create a Signup event
	fmt.Println("\n2. Signup Event:")
	signupEvent := auth.NewSignup(user, clientInfo, "tenant-001")

	if err := signupEvent.IsValid(); err != nil {
		log.Printf("Signup validation error: %v", err)
	} else {
		fmt.Printf("✓ Signup event is valid\n")
		fmt.Printf("  Event Type: %s\n", signupEvent.GetEventType().GetEventType())
		fmt.Printf("  Tenant: %s\n", signupEvent.Tenant)
	}

	// Example 4: EventType utility methods
	fmt.Println("\n4. EventType Utility Methods:")
	fmt.Printf("LOGIN_SUCCESS is login event: %t\n", auth.LOGIN_SUCCESS.IsLoginEvent())
	fmt.Printf("LOGIN_SUCCESS is success event: %t\n", auth.LOGIN_SUCCESS.IsSuccessEvent())
	fmt.Printf("LOGIN_FAILURE is failure event: %t\n", auth.LOGIN_FAILURE.IsFailureEvent())

	// Example 5: JSON serialization
	fmt.Println("\n5. JSON Serialization:")
	jsonData, err := json.MarshalIndent(loginEvent, "", "  ")
	if err != nil {
		log.Printf("JSON serialization error: %v", err)
	} else {
		fmt.Printf("Login Event JSON:\n%s\n", string(jsonData))
	}

	// Example 6: Using AuthEvent interface
	fmt.Println("\n6. AuthEvent Interface:")
	var events []auth.AuthEvent
	events = append(events, loginEvent)
	events = append(events, signupEvent)

	for i, event := range events {
		fmt.Printf("Event %d: %s at %s\n",
			i+1,
			event.GetEventType().GetEventType(),
			event.GetTimestamp().Format("15:04:05"))
	}

	// Example 8: Testing validation errors
	fmt.Println("\n8. Validation Error Testing:")
	invalidUser := auth.User{
		Uid:   "", // Invalid: empty UID
		Email: "test@example.com",
	}

	invalidLoginEvent := auth.NewLogin(invalidUser, clientInfo)
	if err := invalidLoginEvent.IsValid(); err != nil {
		fmt.Printf("✓ Validation correctly caught error: %v\n", err)
	}

	fmt.Println("\n✓ All examples completed successfully!")
}
