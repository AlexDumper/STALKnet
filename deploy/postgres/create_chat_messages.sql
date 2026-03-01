-- Создание таблицы для хранения истории сообщений чата
-- На основе ФЗ-374 от 06.07.2016 (срок хранения - 1 год)

-- Таблица для хранения сообщений чата
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    username VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    client_ip VARCHAR(45) NOT NULL,        -- IPv4 или IPv6
    client_port INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    message_type VARCHAR(20) DEFAULT 'message',  -- message, system, task
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_timestamp ON chat_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_chat_messages_username ON chat_messages(username);
CREATE INDEX IF NOT EXISTS idx_chat_messages_room_timestamp ON chat_messages(room_id, timestamp DESC);

-- Комментарии к таблице
COMMENT ON TABLE chat_messages IS 'История сообщений чата для соблюдения ФЗ-374';
COMMENT ON COLUMN chat_messages.room_id IS 'ID комнаты где было отправлено сообщение';
COMMENT ON COLUMN chat_messages.user_id IS 'ID пользователя отправившего сообщение';
COMMENT ON COLUMN chat_messages.username IS 'Имя пользователя (для отображения)';
COMMENT ON COLUMN chat_messages.content IS 'Текст сообщения';
COMMENT ON COLUMN chat_messages.client_ip IS 'IP адрес клиента';
COMMENT ON COLUMN chat_messages.client_port IS 'Порт клиента';
COMMENT ON COLUMN chat_messages.timestamp IS 'Время отправки сообщения';
COMMENT ON COLUMN chat_messages.message_type IS 'Тип сообщения: message, system, task';

-- Представление для просмотра сообщений за последний год
CREATE OR REPLACE VIEW v_chat_messages_last_year AS
SELECT * FROM chat_messages
WHERE timestamp >= NOW() - INTERVAL '1 year'
ORDER BY timestamp DESC;

-- Функция для автоматической очистки старых сообщений (старше 1 года)
CREATE OR REPLACE FUNCTION fn_cleanup_old_chat_messages() RETURNS void AS $$
BEGIN
    DELETE FROM chat_messages
    WHERE timestamp < NOW() - INTERVAL '1 year';
    
    RAISE NOTICE 'Удалены сообщения старше 1 года. Осталось: %', 
        (SELECT COUNT(*) FROM chat_messages);
END;
$$ LANGUAGE plpgsql;

-- Пример запроса для получения сообщений комнаты
-- SELECT username, content, timestamp, client_ip 
-- FROM chat_messages 
-- WHERE room_id = 1 
-- ORDER BY timestamp DESC 
-- LIMIT 50;
