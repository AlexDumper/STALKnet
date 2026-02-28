// STALKnet Web Client
const messages = document.getElementById("messages");
const input = document.getElementById("messageInput");
const sendBtn = document.getElementById("sendBtn");
const statusDisplay = document.getElementById("statusDisplay");
const usernameDisplay = document.getElementById("usernameDisplay");
const statusLight = document.getElementById("statusLight");
const versionDisplay = document.getElementById("versionDisplay");

// API URL
const API_BASE = window.location.origin;

// Версия приложения
const APP_VERSION = "0.1.3";

// Отображение версии
if (versionDisplay) {
    versionDisplay.textContent = "v" + APP_VERSION;
}

// WebSocket URL - используем тот же хост
const WS_HOST = window.location.hostname;
const WS_PORT = "8083";

// WebSocket
let ws = null;
let wsConnected = false;

// Статус соединения с серверами
let serverConnected = true;

// Текущее отображаемое имя
let currentDisplayName = "Guest";

// Обновление отображаемого имени
function updateDisplayName(newName) {
    currentDisplayName = newName;
    if (usernameDisplay) {
        usernameDisplay.textContent = newName;
    }
}

// Состояния авторизации
const AuthState = {
    Guest: 0,
    EnteringName: 1,
    ConfirmCreate: 2,
    EnteringPassword: 3,
    Authorized: 4
};

let username = "guest";
let sessionId = "";
let accessToken = "";
let refreshToken = "";
let authState = AuthState.Guest;
let pendingUsername = "";
let connected = false;
let userId = null;

// История сообщений
let messageHistory = [];
let historyIndex = -1;
let currentInput = "";

// Ответ на сообщения
let replyToUser = null;

// Debounce для проверки имени
let usernameCheckTimeout = null;
let lastCheckedUsername = "";

// === Сохранение сессии в localStorage ===

function saveSession() {
    const sessionData = {
        username: username,
        sessionId: sessionId,
        accessToken: accessToken,
        refreshToken: refreshToken,
        authState: authState,
        userId: userId,
        savedAt: Date.now()
    };
    localStorage.setItem('stalknet_session', JSON.stringify(sessionData));
    console.log("Сессия сохранена:", sessionData);
}

function loadSession() {
    const sessionData = localStorage.getItem('stalknet_session');
    console.log("Загрузка сессии из localStorage:", sessionData ? "найдена" : "не найдена");
    if (!sessionData) return false;
    
    try {
        const data = JSON.parse(sessionData);
        // Проверяем, не устарела ли сессия (более 7 дней)
        const maxAge = 7 * 24 * 60 * 60 * 1000; // 7 дней в мс
        const age = Date.now() - data.savedAt;
        console.log("Возраст сессии:", Math.floor(age / 1000 / 60), "минут");
        if (age > maxAge) {
            console.log("Сессия устарела, очистка");
            clearSession();
            return false;
        }
        
        username = data.username || "guest";
        sessionId = data.sessionId || "";
        accessToken = data.accessToken || "";
        refreshToken = data.refreshToken || "";
        authState = data.authState || AuthState.Guest;
        userId = data.userId;
        
        console.log("Сессия загружена:", { username, authState, hasToken: !!accessToken });
        return authState === AuthState.Authorized && accessToken !== "";
    } catch (e) {
        console.error("Ошибка загрузки сессии:", e);
        clearSession();
        return false;
    }
}

function clearSession() {
    localStorage.removeItem('stalknet_session');
}

// === Восстановление сессии при загрузке ===

async function restoreSession() {
    if (!loadSession()) {
        return false;
    }
    
    console.log("Восстановление сессии для:", username);
    
    // Проверяем валидность токена
    try {
        const resp = await fetch(API_BASE + "/api/auth/validate", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ token: accessToken })
        });
        
        if (resp.ok) {
            const data = await resp.json();
            if (data.valid) {
                console.log("Сессия восстановлена успешно");
                updateDisplayName(username);
                return true;
            }
        }
        
        // Токен недействителен, пробуем refresh
        return await refreshSession();
    } catch (e) {
        console.error("Ошибка восстановления сессии:", e);
        clearSession();
        return false;
    }
}

async function refreshSession() {
    if (!refreshToken) {
        clearSession();
        return false;
    }
    
    try {
        const resp = await fetch(API_BASE + "/api/auth/refresh", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: refreshToken })
        });
        
        if (resp.ok) {
            const data = await resp.json();
            accessToken = data.access_token;
            refreshToken = data.refresh_token;
            sessionId = data.session_id;
            userId = data.user_id;
            username = data.username;
            authState = AuthState.Authorized;
            
            saveSession();
            console.log("Сессия обновлена через refresh");
            return true;
        }
    } catch (e) {
        console.error("Ошибка обновления сессии:", e);
    }
    
    clearSession();
    return false;
}

// Обновление статуса подключения
function updateStatus() {
    console.log("updateStatus:", {
        serverConnected,
        authState,
        wsConnected
    });

    // Принудительно сохраняем сессию при изменении authState на Authorized
    if (authState === AuthState.Authorized && accessToken) {
        console.log("updateStatus: authState=Authorized, сохраняем сессию...");
        saveSession();
    }

    if (!serverConnected) {
        statusLight.className = "status-indicator status-disconnected";
        statusLight.title = "Нет соединения с сервером";
    } else if (authState === AuthState.Authorized && wsConnected) {
        statusLight.className = "status-indicator status-connected";
        statusLight.title = "Авторизован и подключен";
    } else {
        statusLight.className = "status-indicator status-guest";
        statusLight.title = "Гость / Не авторизован";
    }
}

// Подключение к WebSocket
function connectWebSocket(userId, username) {
    if (ws) {
        ws.close();
    }

    const url = `ws://${WS_HOST}:${WS_PORT}/ws/chat?room_id=1&user_id=${userId}&username=${encodeURIComponent(username)}`;
    console.log("Connecting to WebSocket:", url);
    ws = new WebSocket(url);

    ws.onopen = function() {
        wsConnected = true;
        console.log("WebSocket connected");
        updateStatus();
        addMessage("---", "system");
        addMessage("Подключено к чату", "system");
        addMessage("---", "system");
    };

    ws.onmessage = function(event) {
        try {
            const msg = JSON.parse(event.data);
            if (msg.type === "message") {
                addMessage(msg.content, "user", msg.username);
            } else if (msg.type === "system") {
                addMessage(msg.content, "system");
            } else if (msg.type === "task") {
                addMessage(msg.content, "task");
            }
        } catch (e) {
            addMessage("Ошибка получения сообщения: " + e.message, "system");
        }
    };

    ws.onclose = function() {
        wsConnected = false;
        console.log("WebSocket disconnected");
        updateStatus();
        addMessage("Отключено от чата", "system");
    };

    ws.onerror = function(error) {
        console.error("WebSocket error:", error);
        addMessage("Ошибка WebSocket", "system");
    };
}

// Отправка сообщения через WebSocket
function sendWebSocketMessage(text) {
    if (!wsConnected || !ws) return;

    const msg = {
        type: "message",
        content: text
    };
    ws.send(JSON.stringify(msg));
}

// Загрузка контента из API
async function loadContent(key, callback) {
    try {
        const resp = await fetch(API_BASE + "/api/content/" + key + "?auth_state=" + authState);
        if (resp.ok) {
            const data = await resp.json();
            if (data.content) {
                // Разбиваем контент на строки и отображаем
                const lines = data.content.split("\n");
                lines.forEach(line => addMessage(line, "system"));
                return;
            }
        }
    } catch (e) {
        console.log("Failed to load content:", e.message);
    }
    // Если не удалось загрузить, вызываем callback с дефолтным контентом
    if (callback) callback();
}

// Проверка доступности сервера
async function checkServer() {
    try {
        const resp = await fetch(API_BASE + "/health");
        const data = await resp.json();
        serverConnected = data.status === "ok";
    } catch (e) {
        console.log("Server check failed:", e.message);
        serverConnected = false;
    }
    updateStatus();
}

// Запускаем проверку при загрузке
checkServer();

// Восстанавливаем сессию при загрузке
(async function initSession() {
    const restored = await restoreSession();
    if (restored) {
        updateDisplayName(username);
        updateStatus();
        // Подключаемся к WebSocket
        connectWebSocket(userId, username);
        addMessage("---", "system");
        addMessage("Сессия восстановлена", "system");
        addMessage("С возвращением, " + username + "!", "system");
        addMessage("---", "system");
    }
})();

setTimeout(() => {
    connected = true;
    updateStatus();
    // Загружаем приветственное сообщение из базы
    loadContent("help_welcome", function() {
        // Дефолтное приветствие
        addMessage("---", "system");
        addMessage("Добро пожаловать в STALKnet!", "system");
        addMessage("Введите /help для списка команд", "system");
        addMessage("Введите /auth для авторизации", "system");
        addMessage("---", "system");
    });
}, 1000);

function addMessage(text, type, msgUsername = null, isReply = false, recipientUsername = null) {
    type = type || "system";
    const div = document.createElement("div");
    div.className = "message " + type;

    if (isReply && type === "user") {
        div.className += " reply";
    }

    const time = new Date().toLocaleTimeString('ru-RU', { 
        hour: '2-digit', 
        minute: '2-digit'
    });
    let usernameDisplay = "";

    if (msgUsername) {
        usernameDisplay = "<span class=\"username\" onclick=\"setReplyTo('" + msgUsername + "')\">[" + msgUsername + "]</span> ";
    }

    if (recipientUsername) {
        usernameDisplay += "<span class=\"username\" onclick=\"setReplyTo('" + recipientUsername + "')\">> [" + recipientUsername + "]</span> ";
    }

    // Экранируем HTML спецсимволы в тексте
    const safeText = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');

    div.innerHTML = "<span class=\"timestamp\">[" + time + "]</span>" + usernameDisplay + safeText;
    messages.appendChild(div);
    messages.scrollTop = messages.scrollHeight;
}

function sendMessage() {
    let text = input.value.trim();
    if (!text) return;

    messageHistory.push(text);
    historyIndex = messageHistory.length;
    input.value = "";

    let isReply = false;
    let recipientUser = null;
    const replyMatch = text.match(/^\[([^\]]+)\]:\s*(.*)/);
    if (replyMatch) {
        isReply = true;
        recipientUser = replyMatch[1];
        text = replyMatch[2];
    }

    if (text.startsWith("/")) {
        handleCommand(text);
    } else {
        if (authState === AuthState.EnteringName) {
            handleEnteringName("", [text]);
            return;
        }
        if (authState === AuthState.EnteringPassword) {
            handleEnteringPassword("", [text]);
            return;
        }

        if (authState !== AuthState.Authorized) {
            addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
            return;
        }

        // Отправляем через WebSocket
        sendWebSocketMessage(text);

        // Отображаем локально (сообщение придет обратно через WebSocket)
        // addMessage(text, "user", username, isReply, recipientUser);
    }
}

function setReplyTo(nick) {
    replyToUser = nick;
    input.value = "[" + nick + "]: ";
    input.focus();
}

window.setReplyTo = setReplyTo;

async function handleCommand(cmd) {
    const parts = cmd.trim().split(/\s+/);
    const command = parts[0].toLowerCase();
    const args = parts.slice(1);

    if (authState === AuthState.EnteringName) {
        handleEnteringName(command, args);
        return;
    }
    if (authState === AuthState.EnteringPassword) {
        handleEnteringPassword(command, args);
        return;
    }

    switch(command) {
        case "/help":
            // Загружаем справку из базы данных в зависимости от статуса авторизации
            const helpKey = authState === AuthState.Authorized ? "help_authorized" : "help_guest";
            loadContent(helpKey, function() {
                // Ошибка если не загрузилось из БД
                addMessage("---", "system");
                addMessage("ОШИБКА: Нет связи с базой статического контента", "system");
                addMessage("---", "system");
            });
            break;
        case "/clear":
            messages.innerHTML = "";
            addMessage("---", "system");
            addMessage("Экран очищен", "system");
            addMessage("---", "system");
            break;
        case "/connect":
            const statusIcon = connected ? "?" : "0";
            const statusText = connected ? "Подключено" : "Отключено";
            addMessage("---", "system");
            addMessage("Статус: " + statusIcon + " " + statusText, "system");
            addMessage("---", "system");
            break;
        case "/quit":
            // Сначала logout, потом прощание
            await handleLogout();
            addMessage("---", "system");
            addMessage("До свидания!", "system");
            addMessage("---", "system");
            break;
        case "/logout":
            await handleLogout();
            break;
        case "/nick":
            if (authState !== AuthState.Authorized) {
                addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
                return;
            }
            if (args.length === 0) {
                addMessage("---", "system");
                addMessage("Использование: /nick <имя>", "system");
                addMessage("---", "system");
            } else {
                const oldNick = username;
                username = args[0];
                userDisplay.innerHTML = "+ user: " + username + " +";
                addMessage("---", "system");
                addMessage("Имя изменено с '" + oldNick + "' на '" + username + "'", "system");
                addMessage("---", "system");
            }
            break;
        case "/mock":
            if (authState !== AuthState.Authorized) {
                addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
                return;
            }
            if (args.length === 0) {
                addMessage("---", "system");
                addMessage("Использование: /mock <текст>", "system");
                addMessage("---", "system");
            } else {
                addMessage(args.join(" "), "user", username);
            }
            break;
        case "/mockmsg":
            if (authState !== AuthState.Authorized) {
                addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
                return;
            }
            const msgs = [
                "Ни пуха, ни пера, сталкеры!",
                "Здорово, бродяги!",
                "С прибытием в Зону!",
                "Аномалия рядом, будь осторожен!",
                "Кровосос в болотах замечен!",
                "Выход есть, я видел карту!",
                "Артефакт фонит, нужен контейнер!",
                "Куплю патроны, дорого!",
                "Продам аптечки, дёшево!",
                "Кто на Свалку идёт? Вместе безопаснее!",
                "Монолитовцы атакуют блокпост!",
                "Выброс скоро, нужен укрытие!",
                "Стрелок... ты слышишь меня?",
                "За Монолит! Во имя Зоны!",
                "Долг превыше всего!"
            ];
            const users = ["alice", "bob", "charlie", "diana", "Стрелок", "Волк", "Призрак", "Меченый"];
            const idx = Math.floor(Math.random() * msgs.length);
            const user = users[Math.floor(Math.random() * users.length)];
            addMessage(msgs[idx], "user", user);
            break;
        case "/mocktask":
            if (authState !== AuthState.Authorized) {
                addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
                return;
            }
            const tasks = [
                { id: 42, title: "Найти артефакт 'Медуза'", client: "Сахаров", reward: "1500 RU, артефакт 'Кристалл'" },
                { id: 17, title: "Уничтожить гнездо кровососов", client: "Сидорович", reward: "2000 RU, аптечки x5" },
                { id: 89, title: "Доставить контейнер с образцами", client: "Волк", reward: "1000 RU, патроны 5.45x39" },
                { id: 56, title: "Исследовать аномалию 'Трамплин'", client: "Академик Круглов", reward: "2500 RU, детектор 'Медведь'" },
                { id: 33, title: "Найти схрон с оружием", client: "Меченый", reward: "Оружие на выбор" },
                { id: 71, title: "Ликвидировать бандгруппу", client: "Долг", reward: "1800 RU, броня 'Берилл'" }
            ];
            const tidx = Math.floor(Math.random() * tasks.length);
            const task = tasks[tidx];
            addMessage("---", "task");
            addMessage("Задание #" + task.id, "task");
            addMessage("" + task.title, "task");
            addMessage("Заказчик: " + task.client, "task");
            addMessage("Награда: " + task.reward, "task");
            addMessage("---", "task");
            break;

        case "/auth":
            handleAuth();
            break;
        case "/y":
            await handleConfirm();
            break;
        case "/n":
            handleDecline();
            break;
        case "/cancel":
            handleCancel();
            break;
        case "/login":
            if (args.length < 2) {
                addMessage("---", "system");
                addMessage("Использование: /login <user> <pass>", "system");
                addMessage("---", "system");
            } else {
                await handleQuickLogin(args[0], args[1]);
            }
            break;

        default:
            addMessage("---", "system");
            addMessage("Неизвестная команда: " + command, "system");
            addMessage("---", "system");
    }
}

// === Функции авторизации ===

function handleAuth() {
    if (authState === AuthState.Authorized) {
        addMessage("Вы уже авторизованы как: " + username, "system");
        addMessage("Session ID: " + sessionId, "system");
        return;
    }

    authState = AuthState.EnteringName;
    addMessage("---", "system");
    addMessage("STALKnet Авторизация", "system");
    addMessage("Требования:", "system");
    addMessage("- Имя: минимум 2 символа", "system");
    addMessage("- Пароль: минимум 6 символов", "system");
    addMessage("Введите имя сталкера:", "system");
    addMessage("Введите /cancel для отмены", "system");
    addMessage("---", "system");
}

async function handleEnteringName(command, args) {
    if (command === "/cancel") {
        handleCancel();
        return;
    }

    if (command.startsWith("/")) {
        addMessage("Введите имя сталкера или /cancel для отмены.", "system");
        return;
    }

    const usernameInput = args.join(" ").trim();
    if (usernameInput === "") {
        addMessage("Имя не может быть пустым. Попробуйте ещё раз или /cancel для отмены.", "system");
        return;
    }

    // Проверка длины имени
    if (usernameInput.length < 2) {
        addMessage("---", "system");
        addMessage("Ошибка: имя слишком короткое!", "system");
        addMessage("Требуется минимум 2 символа", "system");
        addMessage("Попробуйте ещё раз или /cancel для отмены", "system");
        addMessage("---", "system");
        return;
    }

    // Показываем сообщение о проверке
    addMessage("Проверка имени '" + usernameInput + "'...", "system");

    // Проверяем, существует ли пользователь через API
    try {
        const checkResp = await fetch(API_BASE + "/api/auth/check-username", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: usernameInput })
        });

        if (checkResp.ok) {
            const data = await checkResp.json();
            if (data.exists) {
                // Пользователь существует - сразу переходим к вводу пароля
                pendingUsername = usernameInput;
                authState = AuthState.EnteringPassword;
                addMessage("Пользователь '" + pendingUsername + "' найден.", "system");
                addMessage("Введите пароль:", "system");
                addMessage("Требование: минимум 6 символов", "system");
                addMessage("Введите /cancel для отмены", "system");
                addMessage("---", "system");
                return;
            }
        }
    } catch (e) {
        addMessage("Ошибка проверки: " + e.message, "system");
    }

    // Пользователь не найден - предлагаем создать профиль
    pendingUsername = usernameInput;
    authState = AuthState.ConfirmCreate;

    addMessage("---", "system");
    addMessage("Имя '" + pendingUsername + "' свободно.", "system");
    addMessage("Создать профиль?", "system");
    addMessage("Введите /y для подтверждения или /n для отмены", "system");
    addMessage("---", "system");
}

function handleConfirm() {
    if (authState !== AuthState.ConfirmCreate) {
        addMessage("Нечего подтверждать. Введите /auth для начала авторизации.", "system");
        return;
    }

    authState = AuthState.EnteringPassword;
    addMessage("---", "system");
    addMessage("Введите пароль для: " + pendingUsername, "system");
    addMessage("Требование: минимум 6 символов", "system");
    addMessage("Введите /cancel для отмены", "system");
    addMessage("---", "system");
}

function handleDecline() {
    if (authState !== AuthState.ConfirmCreate) {
        addMessage("Нечего отклонять.", "system");
        return;
    }
    cancelAuth();
}

async function handleEnteringPassword(command, args) {
    if (command === "/cancel") {
        handleCancel();
        return;
    }

    if (command.startsWith("/")) {
        addMessage("Введите пароль или /cancel для отмены.", "system");
        return;
    }

    const password = args.join(" ").trim();
    if (password === "") {
        addMessage("Пароль не может быть пустым. Попробуйте ещё раз или /cancel для отмены.", "system");
        return;
    }

    // Проверка длины пароля
    if (password.length < 6) {
        addMessage("---", "system");
        addMessage("Ошибка: пароль слишком короткий!", "system");
        addMessage("Требуется минимум 6 символов", "system");
        addMessage("Попробуйте ещё раз или /cancel для отмены", "system");
        addMessage("---", "system");
        return;
    }

    // Сначала пробуем зарегистрировать
    try {
        const registerResp = await fetch(API_BASE + "/api/auth/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                username: pendingUsername,
                password: password
            })
        });

        if (registerResp.status === 409) {
            // Пользователь уже существует, пробуем login
            addMessage("Пользователь существует. Выполняется вход...", "system");
        } else if (!registerResp.ok) {
            const errData = await registerResp.json();
            addMessage("Ошибка регистрации: " + (errData.error || "Неизвестная ошибка"), "system");
            cancelAuth();
            return;
        }
    } catch (e) {
        addMessage("Ошибка соединения с сервером: " + e.message, "system");
        cancelAuth();
        return;
    }

    // Теперь логинимся
    try {
        const loginResp = await fetch(API_BASE + "/api/auth/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                username: pendingUsername,
                password: password
            })
        });

        if (!loginResp.ok) {
            const errData = await loginResp.json();
            addMessage("Ошибка входа: " + (errData.error || "Неизвестная ошибка"), "system");
            cancelAuth();
            return;
        }

        const tokenData = await loginResp.json();

        // Сохраняем сессию
        accessToken = tokenData.access_token;
        refreshToken = tokenData.refresh_token;
        sessionId = tokenData.session_id;
        username = tokenData.username;
        userId = tokenData.user_id;
        authState = AuthState.Authorized;
        
        // Сохраняем в localStorage
        saveSession();

        // Обновляем отображение
        const shortSid = sessionId.length > 12 ? "..." + sessionId.slice(-12) : sessionId;
        updateDisplayName(username);

        addMessage("---", "system");
        addMessage("Профиль создан/найден успешно!", "system");
        addMessage("Добро пожаловать, " + username + "!", "system");
        addMessage("Ваш Session ID: " + shortSid, "system");
        addMessage("---", "system");

        // Подключаемся к WebSocket
        connectWebSocket(tokenData.user_id, username);

        // Обновляем статус
        updateStatus();

        pendingUsername = "";
    } catch (e) {
        addMessage("Ошибка соединения с сервером: " + e.message, "system");
        cancelAuth();
    }
}

async function handleQuickLogin(user, pass) {
    addMessage("Выполняется вход как " + user + "...", "system");

    try {
        const loginResp = await fetch(API_BASE + "/api/auth/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: user, password: pass })
        });

        if (!loginResp.ok) {
            const errData = await loginResp.json();
            addMessage("Ошибка входа: " + (errData.error || "Неизвестная ошибка"), "system");
            return;
        }

        const tokenData = await loginResp.json();
        console.log("Получены токены:", { username: tokenData.username, hasAccessToken: !!tokenData.access_token });

        // Сохраняем сессию
        accessToken = tokenData.access_token;
        refreshToken = tokenData.refresh_token;
        sessionId = tokenData.session_id;
        username = tokenData.username;
        userId = tokenData.user_id;
        authState = AuthState.Authorized;

        console.log("Перед сохранением сессии...");
        // Сохраняем в localStorage
        saveSession();
        console.log("После сохранения сессии, проверяем localStorage:", localStorage.getItem('stalknet_session') ? "OK" : "NULL");

        // Обновляем отображение
        const shortSid = sessionId.length > 12 ? "..." + sessionId.slice(-12) : sessionId;
        updateDisplayName(username);

        addMessage("---", "system");
        addMessage("Профиль создан/найден успешно!", "system");
        addMessage("Добро пожаловать, " + username + "!", "system");
        addMessage("Ваш Session ID: " + shortSid, "system");
        addMessage("---", "system");

        // Подключаемся к WebSocket
        try {
            connectWebSocket(tokenData.user_id, username);
        } catch (wsError) {
            console.error("Ошибка подключения к WebSocket:", wsError);
        }

        // Обновляем статус
        updateStatus();
    } catch (e) {
        addMessage("Ошибка соединения с сервером: " + e.message, "system");
    }
}

async function handleLogout() {
    if (authState !== AuthState.Authorized) {
        addMessage("Вы не авторизованы.", "system");
        return;
    }

    // Закрываем WebSocket
    if (ws) {
        ws.close();
        ws = null;
        wsConnected = false;
    }

    try {
        await fetch(API_BASE + "/api/auth/logout", {
            method: "POST",
            headers: {
                "Authorization": "Bearer " + accessToken,
                "Content-Type": "application/json"
            }
        });
    } catch (e) {
        // Игнорируем ошибки сети при logout
    }

    // Очищаем сессию локально и в localStorage
    clearSession();
    accessToken = "";
    refreshToken = "";
    sessionId = "";
    authState = AuthState.Guest;
    userId = null;

    // Обновляем отображение на Guest
    updateDisplayName("Guest");

    // Обновляем статус
    updateStatus();

    addMessage("---", "system");
    addMessage("Выход из аккаунта выполнен", "system");
    addMessage("---", "system");
}

function handleCancel() {
    if (authState === AuthState.EnteringName ||
        authState === AuthState.ConfirmCreate ||
        authState === AuthState.EnteringPassword) {
        cancelAuth();
    } else {
        addMessage("Отмена.", "system");
    }
}

function cancelAuth() {
    addMessage("---", "system");
    addMessage("Авторизация отменена", "system");
    addMessage("---", "system");
    authState = AuthState.Guest;
    pendingUsername = "";
}

// Обработка Enter в поле ввода
input.addEventListener("keydown", function(event) {
    if (event.key === "Enter") {
        sendMessage();
    }
});

// Обработка клика по кнопке отправки
sendBtn.addEventListener("click", sendMessage);

// История команд
input.addEventListener("keydown", function(event) {
    if (event.key === "ArrowUp") {
        if (historyIndex > 0) {
            if (historyIndex === messageHistory.length) {
                currentInput = input.value;
            }
            historyIndex--;
            input.value = messageHistory[historyIndex];
        }
    } else if (event.key === "ArrowDown") {
        if (historyIndex < messageHistory.length - 1) {
            historyIndex++;
            input.value = messageHistory[historyIndex];
        } else {
            historyIndex = messageHistory.length;
            input.value = currentInput;
        }
    }
});

// Focus на поле ввода при загрузке
window.addEventListener("load", function() {
    input.focus();
});

