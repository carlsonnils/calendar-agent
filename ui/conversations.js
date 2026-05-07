
export async function init() {
    const convResp = await fetch("api/conversations");
    const convJson = await convResp.json();
    const cont = document.getElementById("conversations");
    for (let convo of convJson) {
        buildConversationButton(convo, cont);
    }
}

function buildConversationButton(convo, cont) {
    const template = document.getElementById("conversation-button-template");    
    const clone = template.content.cloneNode(true);
    clone.querySelector(".conversation-button").addEventListener("click", openConversation);
    clone.querySelector(".name").textContent = convo.name;
    clone.querySelector(".updated-at").textContent = convo.updated_at;
    clone.querySelector(".message-count").textContent = convo.message_count;
    clone.querySelector(".session-id").textContent = convo.session_id;
    clone.querySelector(".history").textContent = JSON.stringify(convo.history);
    clone.querySelector(".created-at").textContent = convo.created_at;
    cont.appendChild(clone);
}

function openConversation(evnt) {
    // load the conversation history
    // load the conversation ui
    console.log(evnt);
}
