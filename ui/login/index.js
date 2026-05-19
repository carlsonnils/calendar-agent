// PROMPT SUBMISSION EVENT LISTENERS
document.getElementById("submit-login").addEventListener("click", (e) => {
    submitLogin(e);
});
// document.querySelectorAll(".input-container").forEach((x) => { 
//     x.addEventListener("keypress", (e) => {
//         if (e.key === 'Enter') {
//             document.getElementById("submit-login").click();
//         }
//     });
// });


async function submitLogin(e) {
    // prevent page reload
    e.preventDefault();

    // build body with username and password
    const body = {
        username: document.getElementById("username-input").value,
        password: document.getElementById("password-input").value,
    }

    // request login authentication
    const r = await fetch("/login", {
        method: "POST",
        body: JSON.stringify(body),
    });

    // reroute to chat
    if (r.ok) {
        window.location.href = "/";
    } 
    // TODO: add failed login message
}
