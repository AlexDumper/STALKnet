// STALKnet Web Client
const messages = document.getElementById("messages");
const input = document.getElementById("messageInput");
const sendBtn = document.getElementById("sendBtn");
const statusDisplay = document.getElementById("statusDisplay");
const inputLabel = document.getElementById("inputLabel");
const statusLight = document.getElementById("statusLight");

// API URL
const API_BASE = window.location.origin;

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

// Обновление отображаемого имени (только label у поля ввода)
function updateDisplayName(newName) {
    currentDisplayName = newName;
    inputLabel.textContent = newName + ":";
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

// История сообщений
let messageHistory = [];
let historyIndex = -1;
let currentInput = "";

// Ответ на сообщения
let replyToUser = null;

// Debounce для проверки имени
let usernameCheckTimeout = null;
let lastCheckedUsername = "";

// Обновление статуса подключения
function updateStatus() {
    console.log("updateStatus:", {
        serverConnected,
        authState,
        wsConnected
    });

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
        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Подключено к чату", "system");
        addMessage("╰────────────────────────────────────────────╯", "system");
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

setTimeout(() => {
    connected = true;
    updateStatus();
    // Загружаем приветственное сообщение из базы
    loadContent("help_welcome", function() {
        // Дефолтное приветствие
        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Добро пожаловать в STALKnet!", "system");
        addMessage("│ Введите /help для списка команд", "system");
        addMessage("│ Введите /auth для авторизации", "system");
        addMessage("╰────────────────────────────────────────────╯", "system");
    });
}, 1000);

function addMessage(text, type, msgUsername = null, isReply = false, recipientUsername = null) {
    type = type || "system";
    const div = document.createElement("div");
    div.className = "message " + type;

    if (isReply && type === "user") {
        div.className += " reply";
    }

    const time = new Date().toLocaleTimeString();
    let usernameDisplay = "";

    if (msgUsername) {
        usernameDisplay = "<span class=\"username\" onclick=\"setReplyTo('" + msgUsername + "')\">[" + msgUsername + "]</span> ";
    }

    if (recipientUsername) {
        usernameDisplay += "<span class=\"username\" onclick=\"setReplyTo('" + recipientUsername + "')\">> [" + recipientUsername + "]</span> ";
    }

    div.innerHTML = "<span class=\"timestamp\">[" + time + "]</span>" + usernameDisplay + text;
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
            // Загружаем справку из базы данных
            loadContent("help", function() {
                // Дефолтный контент если не загрузился
                const helpKey = authState === AuthState.Authorized ? "help_authorized" : "help_guest";
                loadContent(helpKey, function() {
                    // Совсем дефолтный если ничего не загрузилось
                    addMessage("╭────────────────────────────────────────────╮", "system");
                    addMessage("│ Доступные команды:", "system");
                    addMessage("│ /help - Показать эту справку", "system");
                    addMessage("│ /clear - Очистить экран", "system");
                    addMessage("│ /connect - Статус подключения", "system");
                    addMessage("│ /quit - Выйти из аккаунта и приложения", "system");
                    addMessage("│ /auth - Авторизация", "system");
                    addMessage("│ /logout - Выйти из аккаунта", "system");
                    addMessage("│ /login <user> <pass> - Быстрый вход", "system");
                    if (authState === AuthState.Authorized) {
                        addMessage("│ /nick <name> - Сменить имя", "system");
                        addMessage("│ /mock <text> - Отправить сообщение", "system");
                        addMessage("│ /mockmsg - Случайное сообщение", "system");
                        addMessage("│ /mocktask - Показать задание", "system");
                    }
                    addMessage("╰────────────────────────────────────────────╯", "system");
                });
            });
            break;
        case "/clear":
            messages.innerHTML = "";
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ Экран очищен", "system");
            addMessage("╰────────────────────────────────────╯", "system");
            break;
        case "/connect":
            const statusIcon = connected ? "●" : "○";
            const statusText = connected ? "Подключено" : "Отключено";
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ Статус: " + statusIcon + " " + statusText, "system");
            addMessage("╰────────────────────────────────────╯", "system");
            break;
        case "/quit":
            // Сначала logout, потом прощание
            await handleLogout();
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ До свидания!", "system");
            addMessage("╰────────────────────────────────────╯", "system");
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
                addMessage("╭────────────────────────────────────╮", "system");
                addMessage("│ Использование: /nick <имя>", "system");
                addMessage("╰────────────────────────────────────╯", "system");
            } else {
                const oldNick = username;
                username = args[0];
                userDisplay.innerHTML = "├ user: " + username + " ┤";
                addMessage("╭────────────────────────────────────╮", "system");
                addMessage("│ Имя изменено с '" + oldNick + "' на '" + username + "'", "system");
                addMessage("╰────────────────────────────────────╯", "system");
            }
            break;
        case "/mock":
            if (authState !== AuthState.Authorized) {
                addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
                return;
            }
            if (args.length === 0) {
                addMessage("╭────────────────────────────────────╮", "system");
                addMessage("│ Использование: /mock <текст>", "system");
                addMessage("╰────────────────────────────────────╯", "system");
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
            addMessage("╭────────────────────────────────────────────╮", "task");
            addMessage("│ Задание #" + task.id, "task");
            addMessage("│ " + task.title, "task");
            addMessage("│ Заказчик: " + task.client, "task");
            addMessage("│ Награда: " + task.reward, "task");
            addMessage("╰────────────────────────────────────────────╯", "task");
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
                addMessage("╭────────────────────────────────────╮", "system");
                addMessage("│ Использование: /login <user> <pass>", "system");
                addMessage("╰────────────────────────────────────╯", "system");
            } else {
                await handleQuickLogin(args[0], args[1]);
            }
            break;

        default:
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ Неизвестная команда: " + command, "system");
            addMessage("╰────────────────────────────────────╯", "system");
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
    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ STALKnet Авторизация", "system");
    addMessage("│ Требования:", "system");
    addMessage("│ - Имя: минимум 2 символа", "system");
    addMessage("│ - Пароль: минимум 6 символов", "system");
    addMessage("│ Введите имя сталкера:", "system");
    addMessage("│ Введите /cancel для отмены", "system");
    addMessage("╰────────────────────────────────────────────╯", "system");
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
        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Ошибка: имя слишком короткое!", "system");
        addMessage("│ Требуется минимум 2 символа", "system");
        addMessage("│ Попробуйте ещё раз или /cancel для отмены", "system");
        addMessage("╰────────────────────────────────────────────╯", "system");
        return;
    }

    // Показываем сообщение о проверке
    addMessage("│ Проверка имени '" + usernameInput + "'...", "system");

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
                addMessage("│ Пользователь '" + pendingUsername + "' найден.", "system");
                addMessage("│ Введите пароль:", "system");
                addMessage("│ Требование: минимум 6 символов", "system");
                addMessage("│ Введите /cancel для отмены", "system");
                addMessage("╰────────────────────────────────────────────╯", "system");
                return;
            }
        }
    } catch (e) {
        addMessage("│ Ошибка проверки: " + e.message, "system");
    }

    // Пользователь не найден - предлагаем создать профиль
    pendingUsername = usernameInput;
    authState = AuthState.ConfirmCreate;

    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ Имя '" + pendingUsername + "' свободно.", "system");
    addMessage("│ Создать профиль?", "system");
    addMessage("│ Введите /y для подтверждения или /n для отмены", "system");
    addMessage("╰────────────────────────────────────────────╯", "system");
}

function handleConfirm() {
    if (authState !== AuthState.ConfirmCreate) {
        addMessage("Нечего подтверждать. Введите /auth для начала авторизации.", "system");
        return;
    }

    authState = AuthState.EnteringPassword;
    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ Введите пароль для: " + pendingUsername, "system");
    addMessage("│ Требование: минимум 6 символов", "system");
    addMessage("│ Введите /cancel для отмены", "system");
    addMessage("╰────────────────────────────────────────────╯", "system");
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
        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Ошибка: пароль слишком короткий!", "system");
        addMessage("│ Требуется минимум 6 символов", "system");
        addMessage("│ Попробуйте ещё раз или /cancel для отмены", "system");
        addMessage("╰────────────────────────────────────────────╯", "system");
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
            addMessage("│ Пользователь существует. Выполняется вход...", "system");
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
        authState = AuthState.Authorized;

        // Обновляем отображение
        const shortSid = sessionId.length > 12 ? "..." + sessionId.slice(-12) : sessionId;
        updateDisplayName(username);

        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Профиль создан/найден успешно!", "system");
        addMessage("│ Добро пожаловать, " + username + "!", "system");
        addMessage("│ Ваш Session ID: " + shortSid, "system");
        addMessage("╰────────────────────────────────────────────╯", "system");

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

        accessToken = tokenData.access_token;
        refreshToken = tokenData.refresh_token;
        sessionId = tokenData.session_id;
        username = tokenData.username;
        authState = AuthState.Authorized;

        updateDisplayName(username);

        addMessage("╭────────────────────────────────────────────╮", "system");
        addMessage("│ Вход выполнен успешно!", "system");
        addMessage("│ Добро пожаловать, " + username + "!", "system");
        addMessage("╰────────────────────────────────────────────╯", "system");

        // Подключаемся к WebSocket
        connectWebSocket(tokenData.user_id, username);

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

    // Очищаем сессию локально
    accessToken = "";
    refreshToken = "";
    sessionId = "";
    authState = AuthState.Guest;

    // Обновляем отображение на Guest
    updateDisplayName("Guest");

    // Обновляем статус
    updateStatus();

    addMessage("╭────────────────────────────────────╮", "system");
    addMessage("│ Выход из аккаунта выполнен", "system");
    addMessage("╰────────────────────────────────────╯", "system");
}

function handleCancel() {
    if (authState === AuthState.EnteringName ||
        authState === AuthState.ConfirmCreate ||
        authState === AuthState.EnteringPassword) {
        cancelAuth();
    } else {
        addMessage("Нечего отменять.", "system");
    }
}

function cancelAuth() {
    addMessage("Авторизация отменена.", "system");
    authState = AuthState.Guest;
    pendingUsername = "";
}

sendBtn.addEventListener("click", sendMessage);
input.addEventListener("keypress", (e) => {
    if (e.key === "Enter") sendMessage();
});

input.addEventListener("keydown", (e) => {
    if (e.key === "ArrowUp") {
        e.preventDefault();
        if (historyIndex > 0) {
            if (historyIndex === messageHistory.length) {
                currentInput = input.value;
            }
            historyIndex--;
            input.value = messageHistory[historyIndex];
        }
    } else if (e.key === "ArrowDown") {
        e.preventDefault();
        if (historyIndex < messageHistory.length - 1) {
            historyIndex++;
            input.value = messageHistory[historyIndex];
        } else if (historyIndex === messageHistory.length - 1) {
            historyIndex++;
            input.value = currentInput;
        }
    }
});

document.addEventListener("keydown", (e) => {
    if (e.ctrlKey && e.key === "l") {
        e.preventDefault();
        messages.innerHTML = "";
        addMessage("╭────────────────────────────────────╮", "system");
        addMessage("│ Экран очищен", "system");
        addMessage("╰────────────────────────────────────╯", "system");
    }
});
