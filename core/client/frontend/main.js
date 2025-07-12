import { SendInput, FetchAll, FetchTimestamp } from "../wailsjs/go/main/App.js";

function updateTimestamp() {
  const ts = Math.floor(Date.now() / 1000);
  document.getElementById("timestamp").innerText = ts;
}

setInterval(updateTimestamp, 1000);
updateTimestamp();

function showStatus(message, color) {
  const status = document.getElementById("status");
  status.textContent = message;
  status.style.color = color;
  status.classList.remove("fade");

  setTimeout(() => {
    status.classList.add("fade");
  }, 5000); // 5 seconds
}

window.sendInput = async function () {
  const msgInput = document.getElementById("msgInput");
  const msg = msgInput.value.trim();
  if (msg === "") return;

  const result = await SendInput(msg);
  msgInput.value = ""; // Clear input field

  if (result.includes("❌")) {
    showStatus("❌ Rejected", "red");
  } else {
    showStatus("✔️ Approved", "lightgreen");
  }
};

window.fetchAll = async function () {
  const messages = await FetchAll();
  const formatted = messages.map(m => formatMessage(m)).join("\n");
  document.getElementById("output").innerText = formatted;
};

window.fetchTimestamp = async function () {
  const tsInput = document.getElementById("tsInput");
  const ts = tsInput.value.trim();
  if (ts === "") return;

  const messages = await FetchTimestamp(ts);
  tsInput.value = ""; // Clear input field
  const formatted = messages.map(m => formatMessage(m)).join("\n");
  document.getElementById("output").innerText = formatted;
};

window.clearOutput = function () {
  document.getElementById("output").innerText = "";
};

window.handleKey = function (e, type) {
  if (e.key === "Enter") {
    e.preventDefault();
    if (type === "send") {
      sendInput();
    } else if (type === "fetch") {
      fetchTimestamp();
    }
  }
};

function formatMessage(msg) {
  const match = msg.match(/Time: (\d+)/);
  const ts = match ? match[1] : "???";
  const sender = msg.includes("Sender:") ? msg.split("Sender: ")[1].split(" |")[0] : "Unknown";
  const content = msg.includes("Msg: ") ? msg.split("Msg: ")[1].split(" |")[0] : "Unknown";
  return `[${ts}] ${sender}: ${content}`;
}
