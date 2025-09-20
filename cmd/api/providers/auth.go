package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/authclient"
	"registry-account/pkg/authzclient"
)

func ProvideAuth(params types.AuthParams) (authclient.AuthenticationProvider, error) {
	fields := params.Config.Settings["authentication"]
	if fields == nil {
		return nil, fmt.Errorf("authentication config is nil")
	}

	b, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth config: %w", err)
	}

	var fauth map[string]any
	err = json.Unmarshal(b, &fauth)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth config: %w", err)
	}

	configPath, ok := fauth["config_file_path"].(string)
	if !ok || configPath == "" {
		return nil, fmt.Errorf("firebase auth config path is empty")
	}

	projectID, ok := fauth["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("firebase auth project_id is empty")
	}

	enableTracing, _ := fauth["enabled_tracing"].(bool)

	auth, err := authclient.NewFirebaseAuth(&configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create firebase auth: %w", err)
	}

	err = auth.Initialize(context.Background(), authclient.Config{
		ServiceAccountKeyPath: configPath,
		ProjectID:             projectID,
		EnableTracing:         enableTracing,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase auth: %w", err)
	}

	return auth, nil
}

func ProvideAuthz(params types.AuthzParams) (authzclient.AuthzClient, error) {
	fields := params.Config.Settings["authorization"]
	if fields == nil {
		return nil, fmt.Errorf("authorization config is nil")
	}

	encrypt, ok := fields.(map[string]any)["encrypt_key"].(string)
	if !ok || encrypt == "" {
		return nil, fmt.Errorf("encrypt_key is missing or not a string")
	}

	config := authzclient.ConfigAuthzClient{
		EncryptKey:     encrypt,
		EnabledTracing: params.Obs != nil,
	}

	authz, err := authzclient.NewAuthzClient(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authz client: %w", err)
	}

	return authz, nil
}
