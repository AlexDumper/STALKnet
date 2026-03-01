package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
)

// ChatMessage представляет сообщение чата для сохранения в БД
type ChatMessage struct {
	ID          int       `json:"id"`
	RoomID      int       `json:"room_id"`
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Content     string    `json:"content"`
	ClientIP    string    `json:"client_ip"`
	ClientPort  int       `json:"client_port"`
	Timestamp   time.Time `json:"timestamp"`
	MessageType string    `json:"message_type"`
}

// SetupRouter настраивает роутер compliance service
func SetupRouter(dbHost, dbPort, dbUser, dbPassword, dbName string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
	})

	// Подключение к PostgreSQL
	dbConn := initDatabase(dbHost, dbPort, dbUser, dbPassword, dbName)

	// Создаём репозиторий
	repo := NewComplianceRepository(dbConn)

	// Создаём хендлер
	complianceHandler := NewComplianceHandler(repo)

	// API маршруты
	api := router.Group("/api/compliance")
	{
		api.POST("/messages", complianceHandler.SaveMessage)
		api.GET("/rooms/:id/messages", complianceHandler.GetMessages)
		api.GET("/users/:id/messages", complianceHandler.GetUserMessages)
		api.DELETE("/cleanup", complianceHandler.CleanupOldMessages)
		api.GET("/stats", complianceHandler.GetStats)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
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

	log.Println("Successfully connected to PostgreSQL (Compliance Service)")
	return db
}

// ComplianceHandler обрабатывает запросы к compliance сервису
type ComplianceHandler struct {
	repo *ComplianceRepository
}

func NewComplianceHandler(repo *ComplianceRepository) *ComplianceHandler {
	return &ComplianceHandler{
		repo: repo,
	}
}

// SaveMessage сохраняет сообщение в базу данных
func (h *ComplianceHandler) SaveMessage(c *gin.Context) {
	var msg ChatMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация
	if msg.RoomID == 0 || msg.UserID == 0 || msg.Username == "" || msg.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.SaveMessage(ctx, &msg); err != nil {
		log.Printf("Failed to save message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	log.Printf("Message saved: room=%d, user=%s, ip=%s:%d", msg.RoomID, msg.Username, msg.ClientIP, msg.ClientPort)
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Message saved successfully",
		"message_id": msg.ID,
	})
}

// GetMessages получает сообщения для указанной комнаты
func (h *ComplianceHandler) GetMessages(c *gin.Context) {
	roomIDStr := c.Param("id")
	var roomID int
	fmt.Sscanf(roomIDStr, "%d", &roomID)

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

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

// GetUserMessages получает сообщения от указанного пользователя
func (h *ComplianceHandler) GetUserMessages(c *gin.Context) {
	userIDStr := c.Param("id")
	var userID int
	fmt.Sscanf(userIDStr, "%d", &userID)

	limit := 50
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := h.repo.GetMessagesByUser(ctx, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"user_id":  userID,
		"total":    len(messages),
	})
}

// CleanupOldMessages удаляет сообщения старше 1 года (ФЗ-374)
func (h *ComplianceHandler) CleanupOldMessages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := h.repo.DeleteOldMessages(ctx, 365*24*time.Hour) // 1 год
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup old messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Old messages cleaned up",
		"deleted_count":   count,
		"retention_days":  365,
	})
}

// GetStats возвращает статистику по сообщениям
func (h *ComplianceHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	total, err := h.repo.GetTotalMessages(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_messages": total,
		"retention_days": 365,
		"compliance":     "ФЗ-374 от 06.07.2016",
	})
}
