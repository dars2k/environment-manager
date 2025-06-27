package handlers

import (
	"fmt"
	"net/http"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogHandler handles log-related HTTP requests
type LogHandler struct {
	logService *log.Service
	logger     *logrus.Logger
}

// NewLogHandler creates a new log handler
func NewLogHandler(logService *log.Service, logger *logrus.Logger) *LogHandler {
	return &LogHandler{
		logService: logService,
		logger:     logger,
	}
}

// List retrieves logs with filtering and pagination
func (h *LogHandler) List(c *gin.Context) {
	filter := interfaces.LogFilter{
		Page:  1,
		Limit: 50,
	}

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		var p int
		if _, err := fmt.Sscanf(page, "%d", &p); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := fmt.Sscanf(limit, "%d", &l); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	if envID := c.Query("environmentId"); envID != "" {
		if objID, err := primitive.ObjectIDFromHex(envID); err == nil {
			filter.EnvironmentID = &objID
		}
	}

	if logType := c.Query("type"); logType != "" {
		filter.Type = entities.LogType(logType)
	}

	if level := c.Query("level"); level != "" {
		filter.Level = entities.LogLevel(level)
	}

	if action := c.Query("action"); action != "" {
		filter.Action = entities.ActionType(action)
	}

	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	if startTime := c.Query("startTime"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = t
		}
	}

	if endTime := c.Query("endTime"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = t
		}
	}

	// Get logs
	logs, total, err := h.logService.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list logs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"page":  filter.Page,
				"limit": filter.Limit,
				"total": total,
			},
		},
	})
}

// GetByID retrieves a specific log entry
func (h *LogHandler) GetByID(c *gin.Context) {
	logID := c.Param("id")

	log, err := h.logService.GetByID(c.Request.Context(), logID)
	if err != nil {
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Log not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to get log")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"log": log,
		},
	})
}

// GetEnvironmentLogs retrieves recent logs for a specific environment
func (h *LogHandler) GetEnvironmentLogs(c *gin.Context) {
	envID := c.Param("id")
	limit := 100 // Default limit

	if l := c.Query("limit"); l != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(l, "%d", &parsedLimit); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	logs, err := h.logService.GetEnvironmentLogs(c.Request.Context(), envID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get environment logs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve environment logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"logs": logs,
		},
	})
}

// Count returns the count of logs based on filters
func (h *LogHandler) Count(c *gin.Context) {
	filter := interfaces.LogFilter{}

	// Parse query parameters
	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			filter.StartTime = t
		}
	}

	if envID := c.Query("environmentId"); envID != "" {
		if objID, err := primitive.ObjectIDFromHex(envID); err == nil {
			filter.EnvironmentID = &objID
		}
	}

	if logType := c.Query("type"); logType != "" {
		filter.Type = entities.LogType(logType)
	}

	if level := c.Query("level"); level != "" {
		filter.Level = entities.LogLevel(level)
	}

	// Get count
	count, err := h.logService.Count(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count logs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"count": count,
		},
	})
}
