-- Инициализация базы данных STALKnet

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- Комнаты чата
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by INTEGER REFERENCES users(id),
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rooms_name ON rooms(name);

-- Участники комнат
CREATE TABLE IF NOT EXISTS room_members (
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

-- Сообщения
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_room_id ON messages(room_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- История сообщений чата (для соблюдения ФЗ-374)
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    username VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    client_port INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    message_type VARCHAR(20) DEFAULT 'message',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_chat_messages_timestamp ON chat_messages(timestamp);
CREATE INDEX idx_chat_messages_username ON chat_messages(username);
CREATE INDEX idx_chat_messages_room_timestamp ON chat_messages(room_id, timestamp DESC);

-- Задачи
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    assignee_id INTEGER REFERENCES users(id),
    room_id INTEGER REFERENCES rooms(id),
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_creator_id ON tasks(creator_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_room_id ON tasks(room_id);

-- Статусы задач: open, in_progress, done, confirmed

-- Статический контент (инструкции, справка)
CREATE TABLE IF NOT EXISTS static_content (
    id SERIAL PRIMARY KEY,
    content_key VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    min_auth_state INT DEFAULT 0,
    max_auth_state INT DEFAULT 4,
    language VARCHAR(10) DEFAULT 'ru',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_static_content_key ON static_content(content_key);
CREATE INDEX idx_static_content_auth ON static_content(min_auth_state, max_auth_state);
CREATE INDEX idx_static_content_language ON static_content(language);

-- Начальные данные для справки
-- Приветственное сообщение (загружается при подключении)
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state) VALUES
('help_welcome', 'Приветствие',
'───
Добро пожаловать в STALKnet!
• /help - Список команд
• /auth - Авторизация
— Автор полностью согласен с требованиями регулятора о предоставлении информации о действиях пользователей. Все сообщения логируются, могут быть просмотрены и прочитаны. Срок хранения - 1 год. Требование составлено на основе ФЗ-374 от 06.07.2016
───', 0, 4);

-- Справка для гостей (неавторизованные пользователи)
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state) VALUES
('help_guest', 'Базовые команды',
'───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
───', 0, 0);

-- Справка для авторизованных пользователей
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state) VALUES
('help_authorized', 'Полный список команд',
'───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
• /nick <name> - Сменить имя
• /mock <text> - Отправить сообщение
• /mockmsg - Случайное сообщение
• /mocktask - Показать задание
───', 4, 4);


-- ============================================================================
-- COMPLIANCE SERVICE TABLES (ФЗ-374 от 06.07.2016)
-- ============================================================================
-- Таблицы для хранения событий пользователей и сессий в соответствии с
-- Федеральным законом № 374-ФЗ "О противодействии терроризму"
-- ============================================================================

-- События пользователей (бессрочное хранение)
-- События: CREATE (регистрация), UPDATE (смена имени)
-- ВАЖНО: Данные в этой таблице НЕ очищаются и накапливаются бессрочно!
CREATE TABLE IF NOT EXISTS user_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- CREATE, UPDATE
    user_id INTEGER,                        -- ID пользователя (NULL для CREATE до создания)
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес клиента (IPv4/IPv6)
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

-- Комментарии к таблице user_events
COMMENT ON TABLE user_events IS 'События пользователей для соблюдения ФЗ-374. Данные НЕ очищаются!';
COMMENT ON COLUMN user_events.event_type IS 'Тип события: CREATE (регистрация), UPDATE (смена имени)';
COMMENT ON COLUMN user_events.username IS 'Текущее имя пользователя';
COMMENT ON COLUMN user_events.client_ip IS 'IP адрес клиента';
COMMENT ON COLUMN user_events.client_port IS 'Порт клиента';
COMMENT ON COLUMN user_events.old_username IS 'Старое имя (для UPDATE)';
COMMENT ON COLUMN user_events.new_username IS 'Новое имя (для UPDATE)';
COMMENT ON COLUMN user_events.metadata IS 'Дополнительные данные в формате JSON';


-- Сессии пользователей (хранение 1 год)
-- События: LOGIN (вход), LOGOUT (выход), DISCONNECT (разрыв соединения)
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- LOGIN, LOGOUT, DISCONNECT
    user_id INTEGER NOT NULL,               -- ID пользователя
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    session_id VARCHAR(255),                -- ID сессии (JWT token ID)
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес клиента (IPv4/IPv6)
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

-- Комментарии к таблице user_sessions
COMMENT ON TABLE user_sessions IS 'Сессии пользователей (LOGIN/LOGOUT) для соблюдения ФЗ-374. Хранение 1 год.';
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


-- Функция для автоматической очистки старых данных (старше 1 года)
-- Очищает: chat_messages, user_sessions
-- НЕ очищает: user_events (хранится бессрочно!)
CREATE OR REPLACE FUNCTION fn_cleanup_old_compliance_data()
RETURNS void AS $$
DECLARE
    deleted_messages_count INTEGER;
    deleted_sessions_count INTEGER;
BEGIN
    -- Подсчёт удаляемых сообщений
    SELECT COUNT(*) INTO deleted_messages_count
    FROM chat_messages
    WHERE timestamp < NOW() - INTERVAL '1 year';

    -- Подсчёт удаляемых сессий
    SELECT COUNT(*) INTO deleted_sessions_count
    FROM user_sessions
    WHERE login_time < NOW() - INTERVAL '1 year';

    -- Обновляем длительность для сессий без logout_time
    UPDATE user_sessions
    SET duration_seconds = EXTRACT(EPOCH FROM (NOW() - login_time))::INTEGER
    WHERE logout_time IS NULL;

    -- Удаляем сообщения чата старше 1 года
    DELETE FROM chat_messages
    WHERE timestamp < NOW() - INTERVAL '1 year';

    -- Удаляем сессии старше 1 года
    DELETE FROM user_sessions
    WHERE login_time < NOW() - INTERVAL '1 year';

    -- ВАЖНО: user_events НЕ очищается - хранится бессрочно!
    RAISE NOTICE 'Очистка данных Compliance завершена. Удалено сообщений: %, сессий: %. Таблица user_events сохранена полностью.',
        deleted_messages_count, deleted_sessions_count;
END;
$$ LANGUAGE plpgsql;


-- Комментарии к таблицам Compliance
COMMENT ON TABLE chat_messages IS 'Сообщения чата (ФЗ-374). Хранение 1 год.';
COMMENT ON TABLE user_events IS 'События пользователей (ФЗ-374). Хранится бессрочно.';
COMMENT ON TABLE user_sessions IS 'Сессии пользователей (ФЗ-374). Хранение 1 год.';


-- ============================================================================
-- ПРИМЕРЫ ЗАПРОСОВ COMPLIANCE
-- ============================================================================

-- Получить все события пользователя:
-- SELECT * FROM user_events WHERE user_id = 5 ORDER BY timestamp DESC;

-- Получить все регистрации за период:
-- SELECT username, client_ip, timestamp FROM user_events
-- WHERE event_type = 'CREATE' AND timestamp >= NOW() - INTERVAL '30 days'
-- ORDER BY timestamp DESC;

-- Получить все смены имён:
-- SELECT username, old_username, new_username, client_ip, timestamp
-- FROM user_events WHERE event_type = 'UPDATE'
-- ORDER BY timestamp DESC;

-- Получить активные сессии:
-- SELECT * FROM v_user_sessions_active;

-- Получить сессии пользователя:
-- SELECT event_type, username, client_ip, login_time, logout_time, duration_seconds
-- FROM user_sessions WHERE user_id = 5 ORDER BY login_time DESC;

-- Запустить очистку старых данных (1 год):
-- SELECT fn_cleanup_old_compliance_data();

-- ============================================================================
