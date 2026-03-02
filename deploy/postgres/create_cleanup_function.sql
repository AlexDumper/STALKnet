-- Единая функция для очистки данных Compliance (ФЗ-374)
-- Очищает: chat_messages, user_sessions (старше 1 года)
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

-- Комментарии
COMMENT ON FUNCTION fn_cleanup_old_compliance_data() IS 'Очистка старых данных Compliance ( ФЗ-374). user_events НЕ очищается.';
