package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/stalknet/services/auth/repository"
)

type AuthHandler struct {
	repo      *repository.AuthRepository
	jwtSecret string
	db        *sql.DB
}

// SetDB устанавливает соединение с базой данных
func (h *AuthHandler) SetDB(db *sql.DB) {
	h.db = db
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	SessionID    string `json:"session_id"`
}

type ValidateRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateResponse struct {
	Valid    bool   `json:"valid"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func NewAuthHandler(repo *repository.AuthRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register регистрирует нового пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := context.Background()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь
	existingUser, err := h.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создаём пользователя
	userID, err := h.repo.CreateUser(ctx, req.Username, string(hashedPassword), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User registered successfully",
		"user_id":  userID,
		"username": req.Username,
	})
}

// Login авторизует пользователя и выдаёт токены
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := context.Background()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return;
	}

	// Находим пользователя
	user, err := h.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Генерируем токены
	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Создаём сессию
	session := &repository.Session{
		UserID:       user.ID,
		Username:     user.Username,
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		CreatedAt:    time.Now(),
	}

	err = h.repo.CreateSession(ctx, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Обновляем статус пользователя
	_ = h.repo.UpdateUserStatus(ctx, user.ID, "online")

	// Генерируем уникальный session ID на основе username и password hash
	sessionID := generateSessionID(user.Username, user.PasswordHash)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute),
		UserID:       user.ID,
		Username:     user.Username,
		SessionID:    sessionID,
	})
}

// generateSessionID генерирует уникальный ID сессии на основе username и password hash
func generateSessionID(username, passwordHash string) string {
	// Берем первые 10 символов password hash
	hashPart := ""
	if len(passwordHash) >= 10 {
		hashPart = passwordHash[:10]
	} else {
		hashPart = passwordHash
	}
	// Формируем session ID: username + hash
	return username + "_" + hashPart
}

// CheckUsername проверяет существование пользователя
func (h *AuthHandler) CheckUsername(c *gin.Context) {
	ctx := context.Background()

	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if user != nil {
		// Пользователь существует
		c.JSON(http.StatusOK, gin.H{
			"exists":   true,
			"username": user.Username,
		})
	} else {
		// Пользователь не найден
		c.JSON(http.StatusOK, gin.H{
			"exists": false,
		})
	}
}

// Logout завершает сессию
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := context.Background()

	// Получаем токен из заголовка
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	// Удаляем префикс "Bearer " если есть
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.repo.DeleteSession(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Refresh обновляет access токен
func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx := context.Background()

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Находим сессию по refresh токену
	session, err := h.repo.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Проверяем, не истёк ли access токен ещё
	if time.Now().Before(session.ExpiresAt) {
		c.JSON(http.StatusOK, TokenResponse{
			AccessToken:  session.Token,
			RefreshToken: session.RefreshToken,
			ExpiresIn:    int64(time.Until(session.ExpiresAt).Seconds()),
			UserID:       session.UserID,
			Username:     session.Username,
			SessionID:    session.Token[:16],
		})
		return
	}

	// Генерируем новые токены
	newAccessToken, _, err := h.generateTokens(session.UserID, session.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Обновляем сессию
	session.Token = newAccessToken
	session.ExpiresAt = time.Now().Add(15 * time.Minute)
	_ = h.repo.CreateSession(ctx, session)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    int64(15 * time.Minute),
		UserID:       session.UserID,
		Username:     session.Username,
		SessionID:    newAccessToken[:16],
	})
}

// Validate проверяет валидность токена
func (h *AuthHandler) Validate(c *gin.Context) {
	ctx := context.Background()

	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.repo.GetSession(ctx, req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}

	if session == nil || time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusOK, ValidateResponse{Valid: false})
		return
	}

	c.JSON(http.StatusOK, ValidateResponse{
		Valid:    true,
		UserID:   session.UserID,
		Username: session.Username,
	})
}

// GetSessionInfo возвращает информацию о сессии
func (h *AuthHandler) GetSessionInfo(c *gin.Context) {
	ctx := context.Background()

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	session, err := h.repo.GetSession(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.Token[:16],
		"user_id":    session.UserID,
		"username":   session.Username,
		"expires_at": session.ExpiresAt,
	})
}

func (h *AuthHandler) generateTokens(userID int, username string) (string, string, error) {
	// Access token (15 минут)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (7 дней)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// GetJWTSecret возвращает секретный ключ из переменных окружения
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	return secret
}

// GetContent получает статический контент по ключу с учетом статуса авторизации
func (h *AuthHandler) GetContent(c *gin.Context) {
	ctx := context.Background()

	key := c.Param("key")
	authStateStr := c.Query("auth_state")
	
	// Парсим статус авторизации (по умолчанию 0 = guest)
	authState := 0
	if authStateStr != "" {
		if parsed, err := strconv.Atoi(authStateStr); err == nil {
			authState = parsed
		}
	}

	// Запрос к базе данных
	query := `
		SELECT content_key, content_type, title, content 
		FROM static_content 
		WHERE content_key = $1 
		  AND is_active = true
		  AND $2 >= min_auth_state 
		  AND $2 <= max_auth_state
		ORDER BY created_at DESC 
		LIMIT 1
	`

	row := h.db.QueryRowContext(ctx, query, key, authState)

	var resultKey, contentType, title string
	var content string

	err := row.Scan(&resultKey, &contentType, &title, &content)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key": resultKey,
		"type": contentType,
		"title": title,
		"content": content,
	})
}
