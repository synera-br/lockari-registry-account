package authorization

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

// OpenFGAClient wraps the OpenFGA client with additional functionality
type OpenFGAClient struct {
	client      *client.OpenFgaClient
	config      *Config
	logger      Logger
	cache       interface{} // Simplified for now
	mu          sync.RWMutex
	healthState HealthState
}

// ClientOptions defines options for creating a new OpenFGA client
type ClientOptions struct {
	Config *Config
	Logger Logger
	Cache  interface{} // Simplified for now
}

// NewOpenFGAClient creates a new OpenFGA client with all configured services
func NewOpenFGAClient(opts ClientOptions) (*OpenFGAClient, error) {
	fmt.Println("\nInitializing OpenFGA client with options:", *opts.Config)

	if opts.Config == nil {
		return nil, ErrInvalidConfig
	}

	if err := opts.Config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Add credentials if provided

	creds := &credentials.Credentials{
		Method: credentials.CredentialsMethodClientCredentials,
		Config: &credentials.Config{
			ClientCredentialsClientId:       opts.Config.ClientID,
			ClientCredentialsClientSecret:   opts.Config.ClientSecret,
			ClientCredentialsApiTokenIssuer: opts.Config.APITokenIssuer,
			ClientCredentialsApiAudience:    opts.Config.APIAudience,
		},
	}
	fmt.Println("\n Credentials for OpenFGA client:", *creds)

	// Create OpenFGA client configuration
	config := &client.ClientConfiguration{
		ApiUrl:               opts.Config.APIURL,
		StoreId:              opts.Config.StoreID,
		AuthorizationModelId: opts.Config.AuthorizationModelID,
		Credentials:          creds,
	}

	fmt.Println("\nOpenFGA client configuration:", *config)
	// Create OpenFGA client
	fgaClient, err := client.NewSdkClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	c := &OpenFGAClient{
		client:      fgaClient,
		config:      opts.Config,
		logger:      opts.Logger,
		cache:       opts.Cache,
		healthState: HealthStateHealthy,
	}

	return c, nil
}

// Check performs an authorization check
func (c *OpenFGAClient) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	// Validate request
	if err := c.validateCheckRequest(req); err != nil {
		return nil, fmt.Errorf("invalid check request: %w", err)
	}

	// Generate cache key for future use
	cacheKey := c.generateCacheKey(req)
	_ = cacheKey // Avoid unused variable warning for now

	// Check cache first (simplified for now)
	if c.cache != nil && c.config.CacheEnabled {
		// TODO: Implement cache logic when CacheService is defined
	}

	// Perform the actual check
	result, err := c.performCheck(ctx, req)
	if err != nil {
		c.updateHealthState(HealthStateUnhealthy)
		return nil, err
	}

	// Cache the result (simplified for now)
	if c.cache != nil && c.config.CacheEnabled {
		// TODO: Implement cache logic when CacheService is defined
	}

	c.updateHealthState(HealthStateHealthy)

	if c.logger != nil {
		c.logger.Debug("Authorization check completed",
			map[string]interface{}{
				"user":     req.User,
				"relation": req.Relation,
				"object":   req.Object,
				"allowed":  result.Allowed,
			})
	}

	return result, nil
}

// BatchCheck performs multiple authorization checks efficiently
func (c *OpenFGAClient) BatchCheck(ctx context.Context, requests []*CheckRequest) (*BatchCheckResponse, error) {
	if len(requests) == 0 {
		return &BatchCheckResponse{
			Results: make([]*CheckResponse, 0),
		}, nil
	}

	start := time.Now()
	results := make([]*CheckResponse, len(requests))
	errors := make([]error, len(requests))

	// Use goroutines for concurrent checks
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.BatchSize)

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *CheckRequest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := c.Check(ctx, request)
			results[index] = result
			errors[index] = err
		}(i, req)
	}

	wg.Wait()

	// Check for errors
	var firstError error
	for _, err := range errors {
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	response := &BatchCheckResponse{
		Results:  results,
		Duration: time.Since(start),
	}

	if firstError != nil {
		response.Error = firstError.Error()
	}

	return response, nil
}

// ListObjects retrieves objects that a user has a specific relation to
func (c *OpenFGAClient) ListObjects(ctx context.Context, req *ListObjectsRequest) (*ListObjectsResponse, error) {
	if err := c.validateListObjectsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid list objects request: %w", err)
	}

	start := time.Now()
	result, err := c.performListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(start)
	return result, nil
}

// HealthCheck performs a health check on the OpenFGA service
func (c *OpenFGAClient) HealthCheck(ctx context.Context) *HealthCheckResponse {
	start := time.Now()

	// Simple check request to verify connectivity
	testReq := &CheckRequest{
		User:     "user:health-check",
		Relation: "viewer",
		Object:   "resource:health-check",
	}

	_, err := c.performCheck(ctx, testReq)
	duration := time.Since(start)

	status := HealthStateHealthy
	message := "OpenFGA service is healthy"

	if err != nil {
		status = HealthStateUnhealthy
		message = fmt.Sprintf("OpenFGA service is unhealthy: %v", err)
	}

	c.updateHealthState(status)

	return &HealthCheckResponse{
		Status:    status,
		Message:   message,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// Close closes the OpenFGA client and cleans up resources
func (c *OpenFGAClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		// OpenFGA client doesn't have a close method, so we just nil it
		c.client = nil
	}

	return nil
}

// performCheck executes the actual OpenFGA check
func (c *OpenFGAClient) performCheck(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	// Create OpenFGA check request
	checkReq := client.ClientCheckRequest{
		User:     req.User,
		Relation: req.Relation,
		Object:   req.Object,
	}

	// TODO: Add context support when CheckRequestWithContext is used

	// Execute check
	response, err := c.client.Check(ctx).Body(checkReq).Execute()
	if err != nil {
		return nil, fmt.Errorf("OpenFGA check failed: %w", err)
	}

	return &CheckResponse{
		Allowed: response.GetAllowed(),
	}, nil
}

// performListObjects executes the actual OpenFGA list objects
func (c *OpenFGAClient) performListObjects(ctx context.Context, req *ListObjectsRequest) (*ListObjectsResponse, error) {
	// Create OpenFGA list objects request
	listReq := client.ClientListObjectsRequest{
		User:     req.User,
		Relation: req.Relation,
		Type:     req.Type,
	}

	// TODO: Add context support when needed

	// Execute list objects
	response, err := c.client.ListObjects(ctx).Body(listReq).Execute()
	if err != nil {
		return nil, fmt.Errorf("OpenFGA list objects failed: %w", err)
	}

	return &ListObjectsResponse{
		Objects: response.GetObjects(),
	}, nil
}

// validateCheckRequest validates a check request
func (c *OpenFGAClient) validateCheckRequest(req *CheckRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}

	if req.User == "" {
		return ErrInvalidUser
	}

	if req.Relation == "" {
		return ErrInvalidRelation
	}

	if req.Object == "" {
		return ErrInvalidObject
	}

	return nil
}

// validateListObjectsRequest validates a list objects request
func (c *OpenFGAClient) validateListObjectsRequest(req *ListObjectsRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}

	if req.User == "" {
		return ErrInvalidUser
	}

	if req.Relation == "" {
		return ErrInvalidRelation
	}

	if req.Type == "" {
		return ErrInvalidRequest
	}

	return nil
}

// generateCacheKey generates a cache key for a check request
func (c *OpenFGAClient) generateCacheKey(req *CheckRequest) string {
	key := fmt.Sprintf("%s:%s:%s", req.User, req.Relation, req.Object)
	// TODO: Add context support when needed
	return key
}

// updateHealthState updates the health status thread-safely
func (c *OpenFGAClient) updateHealthState(state HealthState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.healthState = state
}

// GetHealthState returns the current health status
func (c *OpenFGAClient) GetHealthState() HealthState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.healthState
}
