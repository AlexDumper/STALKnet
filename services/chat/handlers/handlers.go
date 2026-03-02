package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
	"github.com/stalknet/services/chat/hub"
	"github.com/stalknet/services/chat/repository"
)

// SetupRouter настраивает роутер chat service
func SetupRouter(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	redisHost, redisPort string,
	wsHub *hub.Hub,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Подключение к PostgreSQL (для получения сообщений комнат)
	dbConn := initDatabase(dbHost, dbPort, dbUser, dbPassword, dbName)

	chatHandler := NewChatHandler(
		wsHub,
		dbConn,
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
		// Проверка подключения к БД
		if err := dbConn.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func initDatabase(host, port, user, password, dbname string) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	// Настраиваем пул подключений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL (Chat Service)")
	return db
}

type ChatHandler struct {
	hub  *hub.Hub
	repo *repository.ChatRepository
}

func NewChatHandler(wsHub *hub.Hub, db *sql.DB) *ChatHandler {
	return &ChatHandler{
		hub:  wsHub,
		repo: repository.NewChatRepository(db),
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
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}
	
	// Получаем параметры пагинации
	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	// Получение сообщений из БД через репозиторий
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := h.repo.GetMessagesByRoom(ctx, roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"room_id":  roomID,
		"total":    len(messages),
	})
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
