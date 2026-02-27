package publishers

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// Notification тип уведомления
type Notification struct {
	UserID  int         `json:"user_id"`
	Type    string      `json:"type"` // message, task_assigned, task_completed, task_confirmed
	Payload interface{} `json:"payload"`
}

// Publisher публикует уведомления в Redis
type Publisher struct {
	client *redis.Client
}

func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

// Send отправляет уведомление пользователю
func (p *Publisher) Send(ctx context.Context, notif *Notification) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	// Публикуем в канал пользователя
	channel := "notifications:" + string(rune(notif.UserID))
	return p.client.Publish(ctx, channel, data).Err()
}

// SendTaskAssigned отправляет уведомление о назначении задачи
func (p *Publisher) SendTaskAssigned(ctx context.Context, userID, taskID int, title string) error {
	notif := &Notification{
		UserID: userID,
		Type:   "task_assigned",
		Payload: map[string]interface{}{
			"task_id": taskID,
			"title":   title,
		},
	}
	return p.Send(ctx, notif)
}

// SendTaskCompleted отправляет уведомление о выполнении задачи
func (p *Publisher) SendTaskCompleted(ctx context.Context, userID, taskID int, title string) error {
	notif := &Notification{
		UserID: userID,
		Type:   "task_completed",
		Payload: map[string]interface{}{
			"task_id": taskID,
			"title":   title,
		},
	}
	return p.Send(ctx, notif)
}

// SendMessage отправляет уведомление о новом сообщении
func (p *Publisher) SendMessage(ctx context.Context, userID, roomID int, content string) error {
	notif := &Notification{
		UserID: userID,
		Type:   "message",
		Payload: map[string]interface{}{
			"room_id": roomID,
			"content": content,
		},
	}
	return p.Send(ctx, notif)
}
