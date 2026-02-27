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
    statusDisplay.textContent = "[*] Connected";
    addMessage("Welcome to STALKnet!", "system");
    addMessage("Type /help for available commands", "system");
}, 1000);

function addMessage(text, type) {
    type = type || "system";
    const div = document.createElement("div");
    div.className = "message " + type;
    const time = new Date().toLocaleTimeString();
    div.innerHTML = "<span class=\"timestamp\">[" + time + "]</span> " + text;
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
        addMessage("<" + username + "> " + text, "user");
    }
}

function handleCommand(cmd) {
    const parts = cmd.trim().split(/\s+/);
    const command = parts[0].toLowerCase();
    const args = parts.slice(1);
    
    switch(command) {
        case "/help":
            addMessage("Commands: /help /clear /nick /connect /quit /mock /mockmsg /mocktask", "system");
            break;
        case "/clear":
            messages.innerHTML = "";
            addMessage("Screen cleared", "system");
            break;
        case "/nick":
            if (args.length === 0) {
                addMessage("Usage: /nick <name>", "system");
            } else {
                const oldNick = username;
                username = args[0];
                userDisplay.textContent = "user: " + username;
                addMessage("Username changed from " + oldNick + " to " + username, "system");
            }
            break;
        case "/connect":
            addMessage("Connection status: " + (connected ? "Connected" : "Disconnected"), "system");
            break;
        case "/quit":
            addMessage("Goodbye!", "system");
            break;
        case "/mock":
            if (args.length === 0) {
                addMessage("Usage: /mock <text>", "system");
            } else {
                addMessage("<" + username + "> " + args.join(" "), "user");
            }
            break;
        case "/mockmsg":
            const msgs = ["Hello everyone!", "How are you?", "Working on task #42", "Who is in room?"];
            const users = ["alice", "bob", "charlie", "diana"];
            const idx = Math.floor(Math.random() * msgs.length);
            addMessage("<" + users[idx] + "> " + msgs[idx], "user");
            break;
        case "/mocktask":
            const tasks = ["Review code", "Update documentation", "Fix bug in auth", "Add unit tests"];
            const tidx = Math.floor(Math.random() * tasks.length);
            const tid = Math.floor(Math.random() * 100);
            addMessage("Task #" + tid + " created: " + tasks[tidx], "task");
            addMessage("Assigned to: " + username, "task");
            break;
        default:
            addMessage("Unknown command: " + command, "system");
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
        addMessage("Screen cleared", "system");
    }
});
