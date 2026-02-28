package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает роутер task service
func SetupRouter(dbHost, dbPort, dbUser, dbPassword, dbName string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	taskHandler := NewTaskHandler(dbHost, dbPort, dbUser, dbPassword, dbName)

	api := router.Group("/api/task")
	{
		api.GET("", taskHandler.GetTasks)
		api.POST("", taskHandler.CreateTask)
		api.GET("/:id", taskHandler.GetTask)
		api.PUT("/:id", taskHandler.UpdateTask)
		api.PUT("/:id/complete", taskHandler.CompleteTask)
		api.PUT("/:id/confirm", taskHandler.ConfirmTask)
		api.DELETE("/:id", taskHandler.DeleteTask)
		
		// Задачи по комнате
		api.GET("/room/:room_id", taskHandler.GetRoomTasks)
		
		// Мои задачи
		api.GET("/my/created", taskHandler.GetMyCreatedTasks)
		api.GET("/my/assigned", taskHandler.GetMyAssignedTasks)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

type TaskHandler struct {
	// repository будет добавлен
}

func NewTaskHandler(dbHost, dbPort, dbUser, dbPassword, dbName string) *TaskHandler {
	return &TaskHandler{}
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	// status := c.Query("status")
	// Фильтрация по статусу будет добавлена

	c.JSON(http.StatusOK, gin.H{
		"tasks": []gin.H{
			{"id": 1, "title": "Пример задачи", "status": "open"},
		},
	})
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		AssigneeID  int    `json:"assignee_id"`
		RoomID      int    `json:"room_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Task created", "title": req.Title})
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	taskID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": taskID, "title": "Task " + id})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Task updated", "id": id})
}

func (h *TaskHandler) CompleteTask(c *gin.Context) {
	id := c.Param("id")
	// Установка статуса done
	c.JSON(http.StatusOK, gin.H{"message": "Task completed", "id": id})
}

func (h *TaskHandler) ConfirmTask(c *gin.Context) {
	id := c.Param("id")
	// Подтверждение задачи (статус confirmed)
	c.JSON(http.StatusOK, gin.H{"message": "Task confirmed", "id": id})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted", "id": id})
}

func (h *TaskHandler) GetRoomTasks(c *gin.Context) {
	roomID := c.Param("room_id")
	c.JSON(http.StatusOK, gin.H{"room_id": roomID, "tasks": []gin.H{}})
}

func (h *TaskHandler) GetMyCreatedTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tasks": []gin.H{}})
}

func (h *TaskHandler) GetMyAssignedTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tasks": []gin.H{}})
}
