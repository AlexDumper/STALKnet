package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает роутер notification service
func SetupRouter(redisHost, redisPort string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	notifHandler := NewNotificationHandler(redisHost, redisPort)

	api := router.Group("/api/notification")
	{
		api.GET("/unread", notifHandler.GetUnread)
		api.PUT("/unread/:id/read", notifHandler.MarkAsRead)
		api.PUT("/read-all", notifHandler.MarkAllAsRead)
	}

	// WebSocket для уведомлений
	router.GET("/ws/notification", notifHandler.HandleWebSocket)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

type NotificationHandler struct {
	// Redis client будет добавлен
}

func NewNotificationHandler(redisHost, redisPort string) *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) GetUnread(c *gin.Context) {
	// Получение непрочитанных уведомлений
	c.JSON(http.StatusOK, gin.H{"notifications": []gin.H{}})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read", "id": id})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
