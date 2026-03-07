-- Таблица для временного хранения приватных сообщений для офлайн-получателей
-- Сообщения хранятся 3 суток и затем автоматически удаляются

CREATE TABLE IF NOT EXISTS private_messages_offline (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    sender_username VARCHAR(100) NOT NULL,
    recipient_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + INTERVAL '3 days')
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_private_offline_recipient ON private_messages_offline(recipient_id);
CREATE INDEX IF NOT EXISTS idx_private_offline_expires ON private_messages_offline(expires_at);
CREATE INDEX IF NOT EXISTS idx_private_offline_unread ON private_messages_offline(recipient_id, is_read) WHERE is_read = FALSE;

-- Комментарии
COMMENT ON TABLE private_messages_offline IS 'Временное хранение приватных сообщений для офлайн-получателей (3 суток)';
COMMENT ON COLUMN private_messages_offline.is_read IS 'Флаг прочтения сообщения';
COMMENT ON COLUMN private_messages_offline.expires_at IS 'Время удаления сообщения (3 суток)';

-- Функция для автоматической очистки старых сообщений (старше 3 суток)
CREATE OR REPLACE FUNCTION fn_cleanup_old_private_offline()
RETURNS void AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Подсчёт удаляемых сообщений
    SELECT COUNT(*) INTO deleted_count
    FROM private_messages_offline
    WHERE expires_at < NOW();

    -- Удаляем сообщения старше 3 суток
    DELETE FROM private_messages_offline
    WHERE expires_at < NOW();

    RAISE NOTICE 'Очистка старых приватных сообщений: удалено % записей', deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION fn_cleanup_old_private_offline IS 'Удаление приватных сообщений старше 3 суток';

-- Примеры запросов:
-- Получить все непрочитанные сообщения пользователя:
-- SELECT * FROM private_messages_offline WHERE recipient_id = 5 AND is_read = FALSE;

-- Пометить все сообщения как прочитанные:
-- UPDATE private_messages_offline SET is_read = TRUE WHERE recipient_id = 5;

-- Запустить очистку вручную:
-- SELECT fn_cleanup_old_private_offline();
