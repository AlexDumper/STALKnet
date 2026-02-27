// STALKnet Web Client
const messages = document.getElementById("messages");
const input = document.getElementById("messageInput");
const sendBtn = document.getElementById("sendBtn");
const statusDisplay = document.getElementById("statusDisplay");
const userDisplay = document.getElementById("userDisplay");

let username = "guest";
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
    
    let prefix = "";
    let icon = "○";
    let usernameDisplay = "";
    
    // Отображение имени отправителя
    if (msgUsername) {
        usernameDisplay = "<span class=\"username\" onclick=\"setReplyTo('" + msgUsername + "')\">[" + msgUsername + "]</span> ";
    }
    
    // Отображение имени получателя (для ответов)
    if (recipientUsername) {
        usernameDisplay += "<span class=\"username\" onclick=\"setReplyTo('" + recipientUsername + "')\">> [" + recipientUsername + "]</span> ";
    }
    
    if (type === "system") {
        icon = "●";
    } else if (type === "task") {
        icon = "◆";
    } else if (type === "user") {
        icon = isReply ? "⇔" : "▸";
    }
    
    div.innerHTML = prefix + "<span class=\"timestamp\">[" + time + "]</span> <span class=\"icon\">" + icon + "</span> " + usernameDisplay + text;
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

    switch(command) {
        case "/help":
            addMessage("╭────────────────────────────────────────────╮", "system");
            addMessage("│ Доступные команды:", "system");
            addMessage("│ /help - Показать эту справку", "system");
            addMessage("│ /clear - Очистить экран", "system");
            addMessage("│ /nick - Сменить имя пользователя", "system");
            addMessage("│ /connect - Статус подключения", "system");
            addMessage("│ /quit - Выйти", "system");
            addMessage("│ /mock - Отправить сообщение", "system");
            addMessage("│ /mockmsg - Случайное сообщение", "system");
            addMessage("│ /mocktask - Показать задание", "system");
            addMessage("│ /takemocktask - Взять задание", "system");
            addMessage("╰────────────────────────────────────────────╯", "system");
            break;
        case "/clear":
            messages.innerHTML = "";
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ Экран очищен", "system");
            addMessage("╰────────────────────────────────────╯", "system");
            break;
        case "/nick":
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
        case "/mock":
            if (args.length === 0) {
                addMessage("╭────────────────────────────────────╮", "system");
                addMessage("│ Использование: /mock <текст>", "system");
                addMessage("╰────────────────────────────────────╯", "system");
            } else {
                addMessage(args.join(" "), "user", username);
            }
            break;
        case "/mockmsg":
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
        case "/takemocktask":
            const taskId = Math.floor(Math.random() * 100);
            addMessage("╭────────────────────────────────────────────╮", "task");
            addMessage("│ " + username + " взял задание #" + taskId, "task");
            addMessage("╰────────────────────────────────────────────╯", "task");
            break;
        case "/mocktask":
            const tasks = [
                {
                    id: 42,
                    title: "Найти артефакт 'Медуза'",
                    description: "В Янтарном озере замечен редкий артефакт",
                    client: "Сахаров",
                    reward: "1500 RU, артефакт 'Кристалл'"
                },
                {
                    id: 17,
                    title: "Уничтожить гнездо кровососов",
                    description: "Сталкеры сообщают о логове в болотах",
                    client: "Сидорович",
                    reward: "2000 RU, аптечки x5"
                },
                {
                    id: 89,
                    title: "Доставить контейнер с образцами",
                    description: "Забрать у новичков на Кордоне",
                    client: "Волк",
                    reward: "1000 RU, патроны 5.45x39"
                },
                {
                    id: 56,
                    title: "Исследовать аномалию 'Трамплин'",
                    description: "Зафиксирована аномальная активность",
                    client: "Академик Круглов",
                    reward: "2500 RU, детектор 'Медведь'"
                },
                {
                    id: 33,
                    title: "Найти схрон с оружием",
                    description: "Координаты: свалка, старый ангар",
                    client: "Меченый",
                    reward: "Оружие на выбор"
                },
                {
                    id: 71,
                    title: "Ликвидировать бандгруппу",
                    description: "Бандиты терроризируют новичков",
                    client: "Долг",
                    reward: "1800 RU, броня 'Берилл'"
                }
            ];
            const tidx = Math.floor(Math.random() * tasks.length);
            const task = tasks[tidx];
            addMessage("╭────────────────────────────────────────────╮", "task");
            addMessage("│ Задача #" + task.id + ": " + task.title, "task");
            addMessage("│ Описание: " + task.description, "task");
            addMessage("│ Заказчик: " + task.client, "task");
            addMessage("│ Награда: " + task.reward, "task");
            addMessage("╰────────────────────────────────────────────╯", "task");
            break;
        default:
            addMessage("╭────────────────────────────────────╮", "system");
            addMessage("│ Неизвестная команда: " + command, "system");
            addMessage("╰────────────────────────────────────╯", "system");
    }
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
