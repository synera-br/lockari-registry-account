package webhandler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"lockari-api-application/internal/handler/middleware"
	"lockari-api-application/pkg/authenticator"
	"lockari-api-application/pkg/authorization"
	cryptserver "lockari-api-application/pkg/crypt/crypt_server"
	"lockari-api-application/pkg/tokengen"
	"lockari-api-application/pkg/utils"
)

type auditLogHandler struct {
	auditLogService authorization.AuditLogService
	authzService    authorization.LockariAuthorizationService
	encryptor       cryptserver.CryptDataInterface
	authClient      authenticator.Authenticator
	token           tokengen.TokenGenerator
}

type AuditLogHandlerInterface interface {
	GetLogs(c *gin.Context)
	ExportLogs(c *gin.Context)
	GetLogStats(c *gin.Context)
	GetLogTrends(c *gin.Context)
}

// InitializeAuditLogHandler initializes the audit log handler
func InitializeAuditLogHandler(
	auditLogService authorization.AuditLogService,
	authzService authorization.LockariAuthorizationService,
	encryptor cryptserver.CryptDataInterface,
	authClient authenticator.Authenticator,
	token tokengen.TokenGenerator,
	routerGroup *gin.RouterGroup,
	middlewares ...gin.HandlerFunc,
) (AuditLogHandlerInterface, error) {

	if auditLogService == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "audit log service")
	}

	if authzService == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "authorization service")
	}

	if encryptor == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "encryptor")
	}

	if authClient == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "auth client")
	}

	if token == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "token generator")
	}

	handler := &auditLogHandler{
		auditLogService: auditLogService,
		authzService:    authzService,
		encryptor:       encryptor,
		authClient:      authClient,
		token:           token,
	}

	handler.setupRoutes(routerGroup, middlewares...)

	return handler, nil
}

func (h *auditLogHandler) setupRoutes(routerGroup *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	auditRoutes := routerGroup.Group("/audit")
	auditRoutes.Use(middleware.ValidateTokenJWT(h.token))

	// Apply additional middlewares if provided
	for _, middleware := range middlewares {
		auditRoutes.Use(middleware)
	}

	// Audit logs endpoints
	auditRoutes.GET("/logs", h.GetLogs)
	auditRoutes.GET("/logs/export", h.ExportLogs)
	auditRoutes.GET("/logs/stats", h.GetLogStats)
	auditRoutes.GET("/logs/trends", h.GetLogTrends)
}

// GetLogs retrieves audit logs with filtering, pagination, and sorting
func (h *auditLogHandler) GetLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate JWT token
	token := c.GetHeader("X-TOKEN")
	claims, err := h.token.Validate(token)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Extract user ID from claims
	userID := claims.UserID
	if userID == "" {
		log.Printf("Error: user ID not found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Parse query parameters
	query, err := h.parseAuditLogQuery(c)
	if err != nil {
		log.Printf("Error parsing query parameters: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid query parameters: %v", err)})
		return
	}

	// Check authorization - user can only access their own logs unless they have admin permissions
	if err := h.checkAuditLogAccess(ctx, userID, query); err != nil {
		log.Printf("Access denied for user %s: %v", userID, err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// If user is not an admin, restrict to their own logs
	if !h.isUserAdmin(ctx, userID) {
		query.UserID = userID
	}

	// Query audit logs
	response, err := h.auditLogService.QueryLogs(ctx, query)
	if err != nil {
		log.Printf("Error querying audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportLogs exports audit logs in various formats
func (h *auditLogHandler) ExportLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate JWT token
	token := c.GetHeader("X-TOKEN")
	claims, err := h.token.Validate(token)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Extract user ID from claims
	userID := claims.UserID
	if userID == "" {
		log.Printf("Error: user ID not found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Parse query parameters
	query, err := h.parseAuditLogQuery(c)
	if err != nil {
		log.Printf("Error parsing query parameters: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid query parameters: %v", err)})
		return
	}

	// Check authorization
	if err := h.checkAuditLogAccess(ctx, userID, query); err != nil {
		log.Printf("Access denied for user %s: %v", userID, err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// If user is not an admin, restrict to their own logs
	if !h.isUserAdmin(ctx, userID) {
		query.UserID = userID
	}

	// Get export format
	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	// Export logs
	data, err := h.auditLogService.ExportLogs(ctx, query, format)
	if err != nil {
		log.Printf("Error exporting audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export audit logs"})
		return
	}

	// Determine content type and filename based on format
	var contentType, filename string
	switch format {
	case "csv":
		contentType = "text/csv"
		filename = fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102_150405"))
	case "json":
		contentType = "application/json"
		filename = fmt.Sprintf("audit_logs_%s.json", time.Now().Format("20060102_150405"))
	default:
		contentType = "application/octet-stream"
		filename = fmt.Sprintf("audit_logs_%s.txt", time.Now().Format("20060102_150405"))
	}

	// Set response headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, data)
}

// GetLogStats retrieves statistics about audit logs
func (h *auditLogHandler) GetLogStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate JWT token
	token := c.GetHeader("X-TOKEN")
	claims, err := h.token.Validate(token)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Extract user ID from claims
	userID := claims.UserID
	if userID == "" {
		log.Printf("Error: user ID not found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Check if user has admin permissions to view global stats
	if !h.isUserAdmin(ctx, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - admin permissions required"})
		return
	}

	// Create a basic query for stats
	query := &authorization.AuditLogQuery{
		Limit: 1000, // Limit for stats calculation
	}

	// Get logs for statistics calculation
	response, err := h.auditLogService.QueryLogs(ctx, query)
	if err != nil {
		log.Printf("Error getting audit logs for stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	// Calculate basic statistics
	stats := h.calculateStats(response.Logs)

	c.JSON(http.StatusOK, stats)
}

// GetLogTrends retrieves trends data for audit logs
func (h *auditLogHandler) GetLogTrends(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate JWT token
	token := c.GetHeader("X-TOKEN")
	claims, err := h.token.Validate(token)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Extract user ID from claims
	userID := claims.UserID
	if userID == "" {
		log.Printf("Error: user ID not found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Parse time range
	days := 30 // default
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	// Check if user has admin permissions to view global trends
	var restrictToUser string
	if !h.isUserAdmin(ctx, userID) {
		restrictToUser = userID
	}

	// Create query for trends
	query := &authorization.AuditLogQuery{
		UserID: restrictToUser,
		Limit:  1000,
	}

	// Set date range for trends
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	query.StartTime = &startTime
	query.EndTime = &endTime

	// Get logs for trend calculation
	response, err := h.auditLogService.QueryLogs(ctx, query)
	if err != nil {
		log.Printf("Error getting audit logs for trends: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve trends"})
		return
	}

	// Calculate trends
	trends := h.calculateTrends(response.Logs, days)

	c.JSON(http.StatusOK, trends)
}

// parseAuditLogQuery parses query parameters into an AuditLogQuery struct
func (h *auditLogHandler) parseAuditLogQuery(c *gin.Context) (*authorization.AuditLogQuery, error) {
	query := &authorization.AuditLogQuery{}

	// Basic filters
	query.UserID = c.Query("userId")
	query.UserEmail = c.Query("userEmail")
	query.ResourceType = c.Query("resourceType")
	query.ResourceName = c.Query("resourceName")
	query.Action = c.Query("action")
	query.IPAddress = c.Query("ipAddress")

	// Date filters
	if startStr := c.Query("startDate"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			query.StartTime = &start
		} else {
			return nil, fmt.Errorf("invalid startDate format: %v", err)
		}
	}

	if endStr := c.Query("endDate"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			query.EndTime = &end
		} else {
			return nil, fmt.Errorf("invalid endDate format: %v", err)
		}
	}

	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			query.Limit = limit
		}
	}

	// Sorting
	if sortBy := c.Query("sortBy"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	return query, nil
}

// checkAuditLogAccess checks if the user has access to query audit logs
func (h *auditLogHandler) checkAuditLogAccess(ctx context.Context, userID string, query *authorization.AuditLogQuery) error {
	// Admin users can access all logs
	if h.isUserAdmin(ctx, userID) {
		return nil
	}

	// Regular users can only access their own logs
	if query.UserID != "" && query.UserID != userID {
		return fmt.Errorf("access denied: cannot access other user's logs")
	}

	return nil
}

// isUserAdmin checks if the user has admin permissions
func (h *auditLogHandler) isUserAdmin(ctx context.Context, userID string) bool {
	// Check if user has admin role in the system
	// For now, we'll use a simple check - you can extend this with proper permission checking
	adminPermission, err := h.authzService.CanAccessVault(ctx, userID, "system", authorization.VaultPermissionManage)
	if err != nil {
		log.Printf("Error checking admin permissions for user %s: %v", userID, err)
		return false
	}

	return adminPermission
}

// calculateStats calculates basic statistics from audit logs
func (h *auditLogHandler) calculateStats(logs []authorization.AuditLogData) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_events"] = len(logs)

	// Count by action
	actionCounts := make(map[string]int)
	resourceCounts := make(map[string]int)
	userCounts := make(map[string]int)

	for _, log := range logs {
		actionCounts[log.Action]++
		resourceCounts[log.ResourceType]++
		userCounts[log.UserID]++
	}

	stats["actions"] = actionCounts
	stats["resource_types"] = resourceCounts
	stats["unique_users"] = len(userCounts)

	return stats
}

// calculateTrends calculates trends data from audit logs
func (h *auditLogHandler) calculateTrends(logs []authorization.AuditLogData, days int) map[string]interface{} {
	trends := make(map[string]interface{})

	// Group logs by day
	dailyCounts := make(map[string]int)

	for _, log := range logs {
		dayKey := log.Timestamp.Format("2006-01-02")
		dailyCounts[dayKey]++
	}

	trends["daily_counts"] = dailyCounts
	trends["total_events"] = len(logs)
	trends["period_days"] = days

	return trends
}
