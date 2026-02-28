package handlers

import (
	"embed"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/stalknet/gateway/middleware"
)

//go:embed web/index.html web/app.js
var webFS embed.FS

// SetupRouter настраивает роутер gateway
func SetupRouter(authURL, userURL, chatURL, taskURL string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Статические файлы веб-клиента
	router.GET("/", func(c *gin.Context) {
		data, err := webFS.ReadFile("web/index.html")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	router.GET("/app.js", func(c *gin.Context) {
		data, err := webFS.ReadFile("web/app.js")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Data(200, "application/javascript", data)
	})

	// Прокси для Auth Service
	authProxy := httputil.NewSingleHostReverseProxy(mustParseURL(authURL))
	authGroup := router.Group("/api/auth")
	{
		authGroup.Any("/*path", middleware.Proxy(authProxy))
	}

	// Прокси для User Service
	userProxy := httputil.NewSingleHostReverseProxy(mustParseURL(userURL))
	userGroup := router.Group("/api/user")
	{
		userGroup.Use(middleware.JWTAuth())
		userGroup.Any("/*path", middleware.Proxy(userProxy))
	}

	// Прокси для Chat Service
	chatProxy := httputil.NewSingleHostReverseProxy(mustParseURL(chatURL))
	chatGroup := router.Group("/api/chat")
	{
		chatGroup.Use(middleware.JWTAuth())
		chatGroup.Any("/*path", middleware.Proxy(chatProxy))
	}

	// WebSocket для чата
	router.GET("/ws/chat", func(c *gin.Context) {
		// WebSocket upgrade будет обработан в chat service
		chatProxy.ServeHTTP(c.Writer, c.Request)
	})

	// Прокси для Task Service
	taskProxy := httputil.NewSingleHostReverseProxy(mustParseURL(taskURL))
	taskGroup := router.Group("/api/task")
	{
		taskGroup.Use(middleware.JWTAuth())
		taskGroup.Any("/*path", middleware.Proxy(taskProxy))
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func mustParseURL(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return parsed
}
