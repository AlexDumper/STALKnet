package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	router.Use(gin.Logger())

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
		// Endpoint'ы для офлайн-сообщений требуют JWT авторизации
		api.GET("/offline-messages", JWTMiddleware(), chatHandler.GetOfflineMessages)
		api.POST("/offline-messages/read", JWTMiddleware(), chatHandler.MarkOfflineMessagesRead)
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
	hub         *hub.Hub
	repo        *repository.ChatRepository
	authBaseURL string  // URL Auth Service для поиска пользователей
}

func NewChatHandler(wsHub *hub.Hub, db *sql.DB) *ChatHandler {
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8081"
	}
	
	return &ChatHandler{
		hub:         wsHub,
		repo:        repository.NewChatRepository(db),
		authBaseURL: authURL,
	}
}

// findUserByUsername ищет пользователя через Auth Service
func (h *ChatHandler) findUserByUsername(ctx context.Context, username string) (int, string, error) {
	url := fmt.Sprintf("%s/api/users/search?username=%s", h.authBaseURL, url.QueryEscape(username))
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Users []struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Status   string `json:"status"`
		} `json:"users"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", err
	}
	
	// Ищем точное совпадение
	for _, user := range result.Users {
		if strings.EqualFold(user.Username, username) {
			return user.ID, user.Username, nil
		}
	}
	
	return 0, "", fmt.Errorf("user not found")
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

// GetOfflineMessages получает непрочитанные офлайн-сообщения пользователя
func (h *ChatHandler) GetOfflineMessages(c *gin.Context) {
	// Получаем user_id из контекста (должен быть установлен middleware)
	userID := c.GetInt("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := h.repo.GetUnreadOfflineMessages(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// MarkOfflineMessagesRead помечает все офлайн-сообщения как прочитанные
func (h *ChatHandler) MarkOfflineMessagesRead(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.repo.MarkOfflineMessagesAsRead(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

// JWTMiddleware проверяет JWT токен и извлекает user_id
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Убираем префикс "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Парсим токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "your-secret-key-change-in-production"
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Извлекаем claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Получаем user_id из токена
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
			c.Abort()
			return
		}

		userID := int(userIDFloat)
		c.Set("user_id", userID)
		c.Next()
	}
}
