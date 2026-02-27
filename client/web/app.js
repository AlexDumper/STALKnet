// STALKnet Web Client
const messages = document.getElementById("messages");
const input = document.getElementById("messageInput");
const sendBtn = document.getElementById("sendBtn");
const statusDisplay = document.getElementById("statusDisplay");
const userDisplay = document.getElementById("userDisplay");

let username = "guest";
let connected = false;

setTimeout(() => {
    connected = true;
    statusDisplay.innerHTML = "<span class=\"header-decorator\">[</span>●<span class=\"header-decorator\">]</span> Подключено";
    addMessage("Добро пожаловать в STALKnet!", "system");
    addMessage("Введите /help для списка команд", "system");
}, 1000);

function addMessage(text, type, msgUsername = null) {
    type = type || "system";
    const div = document.createElement("div");
    div.className = "message " + type;
    const time = new Date().toLocaleTimeString();
    
    let prefix = "│ ";
    let icon = "○";
    let usernameDisplay = "";
    
    // Отображение имени отправителя
    if (msgUsername) {
        usernameDisplay = "<span class=\"username\">[" + msgUsername + "]</span> ";
    }
    
    if (type === "system") {
        icon = "●";
    } else if (type === "task") {
        icon = "◆";
    } else if (type === "user") {
        icon = "▸";
    }
    
    div.innerHTML = prefix + "<span class=\"timestamp\">[" + time + "]</span> <span class=\"icon\">" + icon + "</span> " + usernameDisplay + text;
    messages.appendChild(div);
    messages.scrollTop = messages.scrollHeight;
}

function sendMessage() {
    const text = input.value.trim();
    if (!text) return;
    input.value = "";
    if (text.startsWith("/")) {
        handleCommand(text);
    } else {
        addMessage(text, "user", username);
    }
}

function handleCommand(cmd) {
    const parts = cmd.trim().split(/\s+/);
    const command = parts[0].toLowerCase();
    const args = parts.slice(1);

    switch(command) {
        case "/help":
            addMessage("╭────────────────────────────────────────────╮", "system");
            addMessage("│ Доступные команды:                         │", "system");
            addMessage("│ /help     - Показать эту справку           │", "system");
            addMessage("│ /clear    - Очистить экран                 │", "system");
            addMessage("│ /nick     - Сменить имя пользователя       │", "system");
            addMessage("│ /connect  - Статус подключения             │", "system");
            addMessage("│ /quit     - Выйти                          │", "system");
            addMessage("│ /mock     - Отправить сообщение            │", "system");
            addMessage("│ /mockmsg  - Случайное сообщение            │", "system");
            addMessage("│ /mocktask - Уведомление о задаче           │", "system");
            addMessage("╰────────────────────────────────────────────╯", "system");
            break;
        case "/clear":
            messages.innerHTML = "";
            addMessage("Экран очищен", "system");
            break;
        case "/nick":
            if (args.length === 0) {
                addMessage("Использование: /nick <имя>", "system");
            } else {
                const oldNick = username;
                username = args[0];
                userDisplay.innerHTML = "├ user: " + username + " ┤";
                addMessage("Имя изменено с '" + oldNick + "' на '" + username + "'", "system");
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
            addMessage("До свидания!", "system");
            break;
        case "/mock":
            if (args.length === 0) {
                addMessage("Использование: /mock <текст>", "system");
            } else {
                addMessage(args.join(" "), "user", username);
            }
            break;
        case "/mockmsg":
            const msgs = ["Всем привет!", "Как дела?", "Работаю над задачей #42", "Кто в комнате?"];
            const users = ["alice", "bob", "charlie", "diana"];
            const idx = Math.floor(Math.random() * msgs.length);
            addMessage(msgs[idx], "user", users[idx]);
            break;
        case "/mocktask":
            const tasks = ["Ревью кода", "Обновление документации", "Исправление бага в auth", "Добавление тестов"];
            const tidx = Math.floor(Math.random() * tasks.length);
            const tid = Math.floor(Math.random() * 100);
            addMessage("╭────────────────────────────────────────────╮", "task");
            addMessage("◆ Задача #" + tid + " создана: " + tasks[tidx], "task");
            addMessage("│ Назначена: " + username, "task");
            addMessage("╰────────────────────────────────────────────╯", "task");
            break;
        default:
            addMessage("Неизвестная команда: " + command, "system");
    }
}

sendBtn.addEventListener("click", sendMessage);
input.addEventListener("keypress", (e) => {
    if (e.key === "Enter") sendMessage();
});

document.addEventListener("keydown", (e) => {
    if (e.ctrlKey && e.key === "l") {
        e.preventDefault();
        messages.innerHTML = "";
        addMessage("Экран очищен", "system");
    }
});
