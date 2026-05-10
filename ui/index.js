// PROMPT SUBMISSION EVENT LISTENERS
document.getElementById("submit-prompt").addEventListener("click", (e) => {
    submitPrompt();
});
document.getElementById("prompt-text").addEventListener("keypress", (e) => {
    if (e.key === 'Enter') {
        document.getElementById("submit-prompt").click();
    }
});

async function submitPrompt() {
    // hide the welcome message
    document.querySelector("#welcome-message").style.display = "none";
    // show chat begining
    document.querySelector("#chat-begining").style.display = "flex";
    // show the messages
    document.querySelector("#messages").style.display = "flex";

    // get prompt message
    const promptInput = document.getElementById("prompt-text");
    const promptMessage = promptInput.value;

    // update user message block
    updateBlock("user", promptMessage);

    // reset prompt input
    promptInput.value = "";
    promptInput.placeholder = "Response ...";

    // get message response
    const r = await fetch("api/chat", { 
        method: "POST",
        body: JSON.stringify({
            message: promptMessage
        }) 
    });

    // catch http response error
    if (!r.ok) {
        throw new Error(`HTTP error: ${res.status} ${res.statusText}`);
    }

    // get response data
    const d = await r.json();

    // update chat message with response
    updateBlock("chat", d.message);
}

function updateBlock(t, m) {
    // copy template block
    const template = document.querySelector(`#${t}-block-template`);
    const block = template.content.cloneNode(true);

    // update block
    updateBlockTime(block.querySelector(".time-block"));
    block.querySelector(".message-text").innerHTML = m.replace(/\n/g, '<br>');

    // add to messages
    document.querySelector("#messages").append(block);
    const convo = document.querySelector("#conversation");
    convo.scrollTop = convo.scrollHeight;
}

function updateBlockTime(tb) {
    const d = formatDateNow();
    tb.querySelector(".date-part").innerHTML = d.date;
    tb.querySelector(".time-part").innerHTML = d.time;
}

function formatDateNow() {
    const d = new Date();
    const pad = (n) => String(n).padStart(2, '0');
    return {
        "date": `${d.getFullYear()}-${pad(d.getMonth())}-${pad(d.getDate())}`,
        "time": `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`,
    }
}
