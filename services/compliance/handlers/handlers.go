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
)

// UserSession представляет сессию пользователя
type UserSession struct {
	ID              int       `json:"id"`
	EventType       string    `json:"event_type"` // LOGIN, LOGOUT, DISCONNECT
	UserID          int       `json:"user_id"`
	Username        string    `json:"username"`
	SessionID       string    `json:"session_id,omitempty"`
	ClientIP        string    `json:"client_ip"`
	ClientPort      int       `json:"client_port"`
	UserAgent       string    `json:"user_agent,omitempty"`
	LoginTime       time.Time `json:"login_time"`
	LogoutTime      time.Time `json:"logout_time,omitempty"`
	DurationSeconds int       `json:"duration_seconds,omitempty"`
}

// UserEvent представляет событие пользователя
type UserEvent struct {
	ID            int       `json:"id"`
	EventType     string    `json:"event_type"` // CREATE, UPDATE
	UserID        int       `json:"user_id"`
	Username      string    `json:"username"`
	ClientIP      string    `json:"client_ip"`
	ClientPort    int       `json:"client_port"`
	OldUsername   string    `json:"old_username,omitempty"`
	NewUsername   string    `json:"new_username,omitempty"`
	Metadata      string    `json:"metadata,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

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
		// Сообщения чата
		api.POST("/messages", complianceHandler.SaveMessage)
		api.GET("/rooms/:id/messages", complianceHandler.GetMessages)
		api.GET("/users/:id/messages", complianceHandler.GetUserMessages)
		
		// События пользователей
		api.POST("/user-events", complianceHandler.SaveUserEvent)
		api.GET("/user-events", complianceHandler.GetUserEvents)
		api.GET("/user-events/:username", complianceHandler.GetUserEventsByUsername)
		
		// Сессии пользователей
		api.POST("/sessions", complianceHandler.SaveSession)
		api.GET("/sessions", complianceHandler.GetSessions)
		api.GET("/sessions/active", complianceHandler.GetActiveSessions)
		api.GET("/sessions/user/:userId", complianceHandler.GetUserSessions)
		api.PUT("/sessions/:id/logout", complianceHandler.UpdateLogout)
		
		// Общее
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
// Примечание: user_events НЕ очищается - данные о регистрации и смене имён накапливаются бессрочно
func (h *ComplianceHandler) CleanupOldMessages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Удаляем только сообщения чата старше 1 года
	count, err := h.repo.DeleteOldMessages(ctx, 365*24*time.Hour) // 1 год
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup old messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Old chat messages cleaned up (user_events preserved)",
		"deleted_count":   count,
		"retention_days":  365,
		"user_events":     "preserved indefinitely",
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

// SaveUserEvent сохраняет событие пользователя
func (h *ComplianceHandler) SaveUserEvent(c *gin.Context) {
	var event UserEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация
	if event.EventType == "" || event.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event type and username required"})
		return
	}

	if event.EventType != "CREATE" && event.EventType != "UPDATE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event type"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.SaveUserEvent(ctx, &event); err != nil {
		log.Printf("Failed to save user event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user event"})
		return
	}

	log.Printf("User event saved: type=%s, username=%s, ip=%s", event.EventType, event.Username, event.ClientIP)
	c.JSON(http.StatusCreated, gin.H{
		"message":  "User event saved successfully",
		"event_id": event.ID,
	})
}

// GetUserEvents получает все события пользователей
func (h *ComplianceHandler) GetUserEvents(c *gin.Context) {
	eventType := c.Query("event_type")
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

	events, err := h.repo.GetUserEvents(ctx, eventType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"total":  len(events),
	})
}

// GetUserEventsByUsername получает события по имени пользователя
func (h *ComplianceHandler) GetUserEventsByUsername(c *gin.Context) {
	username := c.Param("username")
	limit := 50
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := h.repo.GetUserEventsByUsername(ctx, username, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user events by username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":  events,
		"username": username,
		"total":   len(events),
	})
}

// SaveSession сохраняет сессию пользователя (LOGIN, LOGOUT, DISCONNECT)
func (h *ComplianceHandler) SaveSession(c *gin.Context) {
	var session UserSession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация типа события
	if session.EventType != "LOGIN" && session.EventType != "LOGOUT" && session.EventType != "DISCONNECT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event type must be LOGIN, LOGOUT, or DISCONNECT"})
		return
	}
	if session.Username == "" || session.ClientIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and client_ip required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Для LOGIN создаём новую запись, для LOGOUT/DISCONNECT обновляем существующую
	if session.EventType == "LOGIN" {
		if err := h.repo.SaveSession(ctx, &session); err != nil {
			log.Printf("Failed to save user session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		log.Printf("Session saved: type=%s, username=%s, ip=%s, session_id=%s", session.EventType, session.Username, session.ClientIP, session.SessionID)
		c.JSON(http.StatusCreated, gin.H{
			"message":    "Session saved successfully",
			"session_id": session.ID,
		})
	} else {
		// LOGOUT или DISCONNECT - обновляем существующую сессию
		if err := h.repo.UpdateSessionLogout(ctx, &session); err != nil {
			log.Printf("Failed to update session logout: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}
		log.Printf("Session updated: type=%s, username=%s, ip=%s, session_id=%s", session.EventType, session.Username, session.ClientIP, session.SessionID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Session updated successfully",
		})
	}
}

// GetSessions получает все сессии
func (h *ComplianceHandler) GetSessions(c *gin.Context) {
	eventType := c.Query("event_type")
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

	sessions, err := h.repo.GetSessions(ctx, eventType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetActiveSessions получает активные сессии (без logout_time)
func (h *ComplianceHandler) GetActiveSessions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessions, err := h.repo.GetActiveSessions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetUserSessions получает сессии конкретного пользователя
func (h *ComplianceHandler) GetUserSessions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessions, err := h.repo.GetUserSessions(ctx, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"user_id":  userID,
		"total":    len(sessions),
	})
}

// UpdateLogout обновляет сессию при LOGOUT (устанавливает logout_time)
func (h *ComplianceHandler) UpdateLogout(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateLogout(ctx, sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update logout"})
		return
	}

	log.Printf("Session %d logout updated", sessionID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout recorded successfully",
	})
}
