package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stalknet/services/chat/hub"
)

// SetupRouter настраивает роутер chat service
func SetupRouter(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	redisHost, redisPort string,
	wsHub *hub.Hub,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	chatHandler := NewChatHandler(
		dbHost, dbPort, dbUser, dbPassword, dbName,
		redisHost, redisPort,
		wsHub,
	)

	api := router.Group("/api/chat")
	{
		api.GET("/rooms", chatHandler.GetRooms)
		api.POST("/rooms", chatHandler.CreateRoom)
		api.GET("/rooms/:id/messages", chatHandler.GetMessages)
		api.POST("/rooms/:id/messages", chatHandler.SendMessage)
		api.GET("/rooms/:id/members", chatHandler.GetMembers)
		api.POST("/rooms/:id/join", chatHandler.JoinRoom)
		api.POST("/rooms/:id/leave", chatHandler.LeaveRoom)
	}

	// WebSocket endpoint
	router.GET("/ws/chat", chatHandler.HandleWebSocket)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

type ChatHandler struct {
	hub *hub.Hub
	// repository будет добавлен
}

func NewChatHandler(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	redisHost, redisPort string,
	wsHub *hub.Hub,
) *ChatHandler {
	return &ChatHandler{
		hub: wsHub,
	}
}

func (h *ChatHandler) GetRooms(c *gin.Context) {
	// Получение списка комнат из БД
	c.JSON(http.StatusOK, gin.H{
		"rooms": []gin.H{
			{"id": 1, "name": "general", "description": "Общая комната"},
			{"id": 2, "name": "tasks", "description": "Задачи"},
		},
	})
}

func (h *ChatHandler) CreateRoom(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Room created", "name": req.Name})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomID := c.Param("id")
	// Получение сообщений из БД
	c.JSON(http.StatusOK, gin.H{"messages": []gin.H{}, "room_id": roomID})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent", "content": req.Content})
}

func (h *ChatHandler) GetMembers(c *gin.Context) {
	roomID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"members": []gin.H{}, "room_id": roomID})
}

func (h *ChatHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Joined room", "room_id": roomID})
}

func (h *ChatHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Left room", "room_id": roomID})
}
