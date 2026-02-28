// STALKnet Web Client
const messages = document.getElementById("messages");
const input = document.getElementById("messageInput");
const sendBtn = document.getElementById("sendBtn");
const statusDisplay = document.getElementById("statusDisplay");
const userDisplay = document.getElementById("userDisplay");
const sessionDisplay = document.getElementById("sessionDisplay");

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
let authState = AuthState.Guest;
let pendingUsername = "";
let connected = false;

// История сообщений
let messageHistory = [];
let historyIndex = -1;
let currentInput = "";

// Ответ на сообщения
let replyToUser = null;  // ник получателя

setTimeout(() => {
    connected = true;
    statusDisplay.innerHTML = "<span class=\"header-decorator\">[</span>●<span class=\"header-decorator\">]</span> Подключено";
    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ Добро пожаловать в STALKnet!", "system");
    addMessage("│ Введите /help для списка команд", "system");
    addMessage("│ Введите /auth для авторизации", "system");
    addMessage("╰────────────────────────────────────────────╯", "system");
}, 1000);

function addMessage(text, type, msgUsername = null, isReply = false, recipientUsername = null) {
    type = type || "system";
    const div = document.createElement("div");
    div.className = "message " + type;

    // Добавляем класс reply если это ответ
    if (isReply && type === "user") {
        div.className += " reply";
    }

    const time = new Date().toLocaleTimeString();

    let usernameDisplay = "";

    // Отображение имени отправителя
    if (msgUsername) {
        usernameDisplay = "<span class=\"username\" onclick=\"setReplyTo('" + msgUsername + "')\">[" + msgUsername + "]</span> ";
    }

    // Отображение имени получателя (для ответов)
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

    // Сохраняем в историю
    messageHistory.push(text);
    historyIndex = messageHistory.length;

    input.value = "";

    // Проверяем, есть ли получатель в начале сообщения (формат "[ник]: текст")
    let isReply = false;
    let recipientUser = null;
    const replyMatch = text.match(/^\[([^\]]+)\]:\s*(.*)/);
    if (replyMatch) {
        isReply = true;
        recipientUser = replyMatch[1];
        text = replyMatch[2];  // отрезаем "[ник]: " и оставляем только текст
    }

    if (text.startsWith("/")) {
        handleCommand(text);
    } else {
        // Обработка состояний авторизации - ввод имени/пароля
        if (authState === AuthState.EnteringName) {
            handleEnteringName("", [text]);
            return;
        }
        if (authState === AuthState.EnteringPassword) {
            handleEnteringPassword("", [text]);
            return;
        }
        
        // Блокировка отправки сообщений для неавторизованных
        if (authState !== AuthState.Authorized) {
            addMessage("Требуется авторизация. Введите /auth для авторизации.", "system");
            return;
        }
        addMessage(text, "user", username, isReply, recipientUser);
    }
}

// Функция для установки получателя ответа
function setReplyTo(nick) {
    replyToUser = nick;
    input.value = "[" + nick + "]: ";
    input.focus();
}

// Делаем функцию доступной глобально для onclick
window.setReplyTo = setReplyTo;

function handleCommand(cmd) {
    const parts = cmd.trim().split(/\s+/);
    const command = parts[0].toLowerCase();
    const args = parts.slice(1);

    // Обработка состояний авторизации
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
            addMessage("╭────────────────────────────────────────────╮", "system");
            addMessage("│ Доступные команды:", "system");
            addMessage("│ /help - Показать эту справку", "system");
            addMessage("│ /clear - Очистить экран", "system");
            addMessage("│ /connect - Статус подключения", "system");
            addMessage("│ /quit - Выйти", "system");
            addMessage("│ /auth - Авторизация", "system");
            if (authState === AuthState.Authorized) {
                addMessage("│ /nick <name> - Сменить имя", "system");
                addMessage("│ /mock <text> - Отправить сообщение", "system");
                addMessage("│ /mockmsg - Случайное сообщение", "system");
                addMessage("│ /mocktask - Показать задание", "system");
            }
            addMessage("╰────────────────────────────────────────────╯", "system");
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
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ До свидания!", "system");
            addMessage("╰────────────────────────────────────╯", "system");
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
                {
                    id: 42,
                    title: "Найти артефакт 'Медуза'",
                    client: "Сахаров",
                    reward: "1500 RU, артефакт 'Кристалл'"
                },
                {
                    id: 17,
                    title: "Уничтожить гнездо кровососов",
                    client: "Сидорович",
                    reward: "2000 RU, аптечки x5"
                },
                {
                    id: 89,
                    title: "Доставить контейнер с образцами",
                    client: "Волк",
                    reward: "1000 RU, патроны 5.45x39"
                },
                {
                    id: 56,
                    title: "Исследовать аномалию 'Трамплин'",
                    client: "Академик Круглов",
                    reward: "2500 RU, детектор 'Медведь'"
                },
                {
                    id: 33,
                    title: "Найти схрон с оружием",
                    client: "Меченый",
                    reward: "Оружие на выбор"
                },
                {
                    id: 71,
                    title: "Ликвидировать бандгруппу",
                    client: "Долг",
                    reward: "1800 RU, броня 'Берилл'"
                }
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

        // Команды авторизации
        case "/auth":
            handleAuth();
            break;
        case "/y":
            handleConfirm();
            break;
        case "/n":
            handleDecline();
            break;
        case "/cancel":
            handleCancel();
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
    addMessage("│ Введите имя сталкера:", "system");
    addMessage("│ Введите /cancel для отмены", "system");
    addMessage("╰────────────────────────────────────────────╯", "system");
}

function handleEnteringName(command, args) {
    // /cancel обрабатывается отдельно
    if (command === "/cancel") {
        handleCancel();
        return;
    }

    // Игнорируем другие команды в этом состоянии
    if (command.startsWith("/")) {
        addMessage("Введите имя сталкера или /cancel для отмены.", "system");
        return;
    }

    const usernameInput = args.join(" ").trim();
    if (usernameInput === "") {
        addMessage("Имя не может быть пустым. Попробуйте ещё раз или /cancel для отмены.", "system");
        return;
    }

    // В прототипе просто принимаем имя (без реальной проверки на уникальность)
    // Для демонстрации считаем, что имя уникально
    pendingUsername = usernameInput;
    authState = AuthState.ConfirmCreate;

    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ Имя '" + pendingUsername + "' доступно.", "system");
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

function handleEnteringPassword(command, args) {
    // /cancel обрабатывается отдельно
    if (command === "/cancel") {
        handleCancel();
        return;
    }

    // Игнорируем другие команды в этом состоянии
    if (command.startsWith("/")) {
        addMessage("Введите пароль или /cancel для отмены.", "system");
        return;
    }

    const password = args.join(" ").trim();
    if (password === "") {
        addMessage("Пароль не может быть пустым. Попробуйте ещё раз или /cancel для отмены.", "system");
        return;
    }

    // В прототипе просто создаём "сессию"
    // Генерируем случайный Session ID
    sessionId = generateSessionId();
    username = pendingUsername;
    authState = AuthState.Authorized;

    // Обновляем отображение
    userDisplay.innerHTML = "├ user: " + username + " ┤";
    sessionDisplay.style.display = "inline";
    sessionDisplay.innerHTML = "├ SID: " + sessionId + " ┤";

    addMessage("╭────────────────────────────────────────────╮", "system");
    addMessage("│ Профиль создан успешно!", "system");
    addMessage("│ Добро пожаловать, " + username + "!", "system");
    addMessage("│ Ваш Session ID: " + sessionId, "system");
    addMessage("╰────────────────────────────────────────────╯", "system");

    pendingUsername = "";
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

function generateSessionId() {
    // Генерируем случайный ID в формате hex
    const chars = '0123456789abcdef';
    let result = '';
    for (let i = 0; i < 32; i++) {
        result += chars[Math.floor(Math.random() * chars.length)];
    }
    return result;
}

sendBtn.addEventListener("click", sendMessage);
input.addEventListener("keypress", (e) => {
    if (e.key === "Enter") sendMessage();
});

// Навигация по истории сообщений (стрелки вверх/вниз)
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
