-- Создание таблицы для хранения событий пользователей (ФЗ-374)
-- События: регистрация, смена имени

CREATE TABLE IF NOT EXISTS user_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- CREATE, UPDATE
    user_id INTEGER,                        -- ID пользователя (NULL для CREATE до создания)
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес клиента
    client_port INTEGER NOT NULL,           -- Порт клиента
    old_username VARCHAR(100),              -- Старое имя (для UPDATE)
    new_username VARCHAR(100),              -- Новое имя (для UPDATE)
    metadata JSONB,                         -- Дополнительные данные
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_user_events_event_type ON user_events(event_type);
CREATE INDEX IF NOT EXISTS idx_user_events_user_id ON user_events(user_id);
CREATE INDEX IF NOT EXISTS idx_user_events_username ON user_events(username);
CREATE INDEX IF NOT EXISTS idx_user_events_timestamp ON user_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_user_events_ip ON user_events(client_ip);
CREATE INDEX IF NOT EXISTS idx_user_events_event_timestamp ON user_events(event_type, timestamp DESC);

-- Комментарии к таблице
COMMENT ON TABLE user_events IS 'События пользователей для соблюдения ФЗ-374';
COMMENT ON COLUMN user_events.event_type IS 'Тип события: CREATE (регистрация), UPDATE (смена имени)';
COMMENT ON COLUMN user_events.user_id IS 'ID пользователя';
COMMENT ON COLUMN user_events.username IS 'Текущее имя пользователя';
COMMENT ON COLUMN user_events.client_ip IS 'IP адрес клиента';
COMMENT ON COLUMN user_events.client_port IS 'Порт клиента';
COMMENT ON COLUMN user_events.old_username IS 'Старое имя (для UPDATE)';
COMMENT ON COLUMN user_events.new_username IS 'Новое имя (для UPDATE)';
COMMENT ON COLUMN user_events.metadata IS 'Дополнительные данные в формате JSON';

-- Примеры запросов:
-- Получить все события пользователя
-- SELECT * FROM user_events WHERE user_id = 5 ORDER BY timestamp DESC;

-- Получить все регистрации за период
-- SELECT username, client_ip, timestamp FROM user_events 
-- WHERE event_type = 'CREATE' AND timestamp >= NOW() - INTERVAL '30 days'
-- ORDER BY timestamp DESC;

-- Получить все смены имён
-- SELECT username, old_username, new_username, client_ip, timestamp 
-- FROM user_events WHERE event_type = 'UPDATE'
-- ORDER BY timestamp DESC;
