package entity_registry_request

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// RegistryRequest representa os dados recebidos do frontend para registro
type RegistryRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Tenant   string `json:"tenant" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest representa os dados recebidos do frontend para login
type LoginRequest struct {
	UID           string `json:"uid" binding:"required"`
	Email         string `json:"email" binding:"required"`
	DisplayName   string `json:"displayName,omitempty"`
	EmailVerified bool   `json:"emailVerified"`
	Token         string `json:"token" binding:"required"`
}

// RegistryResponse representa a resposta do registro
type RegistryResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	User      UserData  `json:"user,omitempty"`
	AutoLogin LoginData `json:"autoLogin,omitempty"`
}

type UserData struct {
	Email       string   `json:"email"`
	DisplayName string   `json:"displayName"`
	Tenant      []string `json:"tenant"`
	UID         string   `json:"uid,omitempty"` // Preenchido após criação no Firebase
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"-"`
}

// Validate valida os dados do registro
func (r *RegistryRequest) Validate() error {
	if r == nil {
		return errors.New("registry request is nil")
	}

	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}

	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}

	if !r.IsValidEmail() {
		return errors.New("invalid email format")
	}

	if strings.TrimSpace(r.Tenant) == "" {
		return errors.New("tenant is required")
	}

	if !r.IsValidTenant() {
		return errors.New("invalid tenant format: use only lowercase letters, numbers and hyphens")
	}

	return nil
}

// IsValidEmail valida o formato do email
func (r *RegistryRequest) IsValidEmail() bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(r.Email)
}

// IsValidTenant valida o formato do tenant
func (r *RegistryRequest) IsValidTenant() bool {
	// Tenant deve ter apenas letras minúsculas, números e hífens
	tenantRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	tenant := strings.TrimSpace(r.Tenant)

	if len(tenant) < 3 || len(tenant) > 50 {
		return false
	}

	return tenantRegex.MatchString(tenant)
}

// IsGoogleAuth verifica se é um registro via Google
func (r *RegistryRequest) IsGoogleAuth() bool {
	return r.Password == "google-auth"
}

// GetAuthType retorna o tipo de autenticação
func (r *RegistryRequest) GetAuthType() string {
	if r.IsGoogleAuth() {
		return "google"
	}
	return "email"
}

// Validate valida os dados do login
func (l *LoginRequest) Validate() error {
	if l == nil {
		return errors.New("login request is nil")
	}

	if strings.TrimSpace(l.UID) == "" {
		return errors.New("uid is required")
	}

	if strings.TrimSpace(l.Email) == "" {
		return errors.New("email is required")
	}

	if strings.TrimSpace(l.Token) == "" {
		return errors.New("token is required")
	}

	return nil
}

// GetDisplayNameOrFallback retorna o displayName ou um fallback baseado no email
func (l *LoginRequest) GetDisplayNameOrFallback() string {
	if l.DisplayName != "" {
		return l.DisplayName
	}

	// Usar a parte antes do @ como fallback
	parts := strings.Split(l.Email, "@")
	if len(parts) > 0 {
		name := parts[0]
		// Capitalizar primeira letra de forma simples
		if len(name) > 0 {
			return strings.ToUpper(string(name[0])) + name[1:]
		}
	}

	return "Usuário"
}

// String implementa fmt.Stringer para logs seguros (sem senha)
func (r *RegistryRequest) String() string {
	return fmt.Sprintf("RegistryRequest{Name: %s, Email: %s, Tenant: %s, AuthType: %s}",
		r.Name, r.Email, r.Tenant, r.GetAuthType())
}

// String implementa fmt.Stringer para logs seguros (sem token)
func (l *LoginRequest) String() string {
	return fmt.Sprintf("LoginRequest{UID: %s, Email: %s, DisplayName: %s, EmailVerified: %t}",
		l.UID, l.Email, l.DisplayName, l.EmailVerified)
}
