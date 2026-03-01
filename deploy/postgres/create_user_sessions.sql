-- Создание таблицы для хранения сессий пользователей (LOGIN/LOGOUT)
-- События: вход (LOGIN), выход (LOGOUT), разрыв соединения (DISCONNECT)
-- Срок хранения: 1 год (автоматическая очистка)

CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- LOGIN, LOGOUT, DISCONNECT
    user_id INTEGER NOT NULL,               -- ID пользователя
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    session_id VARCHAR(255),                -- ID сессии (JWT token ID)
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес клиента
    client_port INTEGER NOT NULL,           -- Порт клиента
    user_agent TEXT,                        -- User agent (опционально)
    login_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,    -- Время входа
    logout_time TIMESTAMP WITH TIME ZONE,                           -- Время выхода
    duration_seconds INTEGER,               -- Длительность сессии в секундах
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_user_sessions_event_type ON user_sessions(event_type);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_username ON user_sessions(username);
CREATE INDEX IF NOT EXISTS idx_user_sessions_login_time ON user_sessions(login_time);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(user_id, logout_time) WHERE logout_time IS NULL;

-- Комментарии к таблице
COMMENT ON TABLE user_sessions IS 'Сессии пользователей (LOGIN/LOGOUT) для соблюдения ФЗ-374';
COMMENT ON COLUMN user_sessions.event_type IS 'Тип события: LOGIN (вход), LOGOUT (выход), DISCONNECT (разрыв)';
COMMENT ON COLUMN user_sessions.user_id IS 'ID пользователя';
COMMENT ON COLUMN user_sessions.username IS 'Имя пользователя';
COMMENT ON COLUMN user_sessions.session_id IS 'ID сессии (JWT token ID или session ID)';
COMMENT ON COLUMN user_sessions.client_ip IS 'IP адрес клиента';
COMMENT ON COLUMN user_sessions.client_port IS 'Порт клиента';
COMMENT ON COLUMN user_sessions.user_agent IS 'User agent клиента (браузер/приложение)';
COMMENT ON COLUMN user_sessions.login_time IS 'Время входа';
COMMENT ON COLUMN user_sessions.logout_time IS 'Время выхода (NULL если активная сессия)';
COMMENT ON COLUMN user_sessions.duration_seconds IS 'Длительность сессии в секундах';

-- Представление для просмотра активных сессий
CREATE OR REPLACE VIEW v_user_sessions_active AS
SELECT * FROM user_sessions
WHERE logout_time IS NULL
ORDER BY login_time DESC;

-- Функция для автоматической очистки старых сессий (старше 1 года)
CREATE OR REPLACE FUNCTION fn_cleanup_old_user_sessions() RETURNS void AS $$
BEGIN
    -- Обновляем длительность для сессий без logout_time
    UPDATE user_sessions
    SET duration_seconds = EXTRACT(EPOCH FROM (NOW() - login_time))::INTEGER
    WHERE logout_time IS NULL;

    -- Удаляем сессии старше 1 года
    DELETE FROM user_sessions
    WHERE login_time < NOW() - INTERVAL '1 year';
    
    RAISE NOTICE 'Удалены сессии старше 1 года. Осталось: %', 
        (SELECT COUNT(*) FROM user_sessions);
END;
$$ LANGUAGE plpgsql;

-- Примеры запросов:
-- Получить все сессии пользователя
-- SELECT event_type, username, client_ip, login_time, logout_time, duration_seconds
-- FROM user_sessions WHERE user_id = 5 ORDER BY login_time DESC;

-- Получить активные сессии
-- SELECT * FROM v_user_sessions_active;

-- Статистика по сессиям за сегодня
-- SELECT event_type, COUNT(*) as count FROM user_sessions
-- WHERE DATE(login_time) = CURRENT_DATE GROUP BY event_type;
