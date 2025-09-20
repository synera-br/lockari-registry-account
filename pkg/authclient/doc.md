
// Example usage
func main() {
	// Initialize OpenTelemetry (replace with your actual initialization)
	// initTelemetry()
	// defer shutdownTelemetry()

	ctx := context.Background()

	// Load configuration (replace with your actual configuration loading)
	cfg := Config{
		ServiceAccountKeyPath: "/path/to/serviceAccountKey.json", // Replace with your service account key path
		ProjectID:             "your-project-id",                 // Replace with your Firebase project ID
		EnableTracing:         true,
	}

	// Initialize Firebase Authentication
	firebaseAuth := NewFirebaseAuth()
	err := firebaseAuth.Initialize(ctx, cfg)
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth: %v", err)
	}

	// Example: Authenticate a user
	token := "your-jwt-token" // Replace with a valid JWT token
	decodedToken, err := firebaseAuth.AuthenticateUser(ctx, token)
	if err != nil {
		log.Printf("Error authenticating user: %v", err)
	} else {
		log.Printf("User authenticated: %v", decodedToken.UID)
	}

	// Example: Get user by ID
	uid := "some-user-uid" // Replace with a valid user ID
	user, err := firebaseAuth.GetUserByID(ctx, uid)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
	} else {
		log.Printf("User: %v", user.Email)
	}

	// Example: Update user custom claims
	claims := map[string]interface{}{
		"premium": true,
		"role":    "admin",
	}
	err = firebaseAuth.UpdateUserCustomClaims(ctx, uid, claims)
	if err != nil {
		log.Printf("Error updating user custom claims: %v", err)
	} else {
		log.Printf("User custom claims updated")
	}

	// Example: Extract tenant ID from token
	tenantIDKey := "tenant_id"
	tenantID, err := ExtractTenantIDFromToken(token, tenantIDKey)
	if err != nil {
		log.Printf("Error extracting tenant ID: %v", err)
	} else {
		log.Printf("Tenant ID: %s", tenantID)
	}

	// Example: Gin middleware usage
	router := gin.Default()
	router.Use(firebaseAuth.AuthMiddleware(tenantIDKey))

	router.GET("/protected", func(c *gin.Context) {
		uid := c.GetString("uid")
		tenantID := c.GetString("tenant_id")
		c.JSON(http.StatusOK, gin.H{
			"message":   "Protected endpoint",
			"user_id":   uid,
			"tenant_id": tenantID,
		})
	})

	router.Run(":8080")
}

func initTelemetry() {
	tp, err := NewTracerProvider("saas-platform")
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTracerProvider(tp)

	// Register propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func shutdownTelemetry() {
	tp := otel.GetTracerProvider()
	if tp != nil {
		if t, ok := tp.(interface {
			Shutdown(context.Context) error
		}); ok {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if err := t.Shutdown(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func NewTracerProvider(serviceName string) (trace.TracerProvider, error) {
	exporter, err := newExporter()
	if err != nil {
		return nil, fmt.Errorf("creating exporter: %w", err)
	}

	resource := newResource(serviceName)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)

	return tp, nil
}

func newExporter() (trace.SpanExporter, error) {
	return nil, nil
}

func newResource(serviceName string) *otel.Resource {
	return otel.NewResource(
		attribute.String("service.name", serviceName),
	)
}
```