// PROMPT SUBMISSION EVENT LISTENERS
document.getElementById("submit-login").addEventListener("click", (e) => {
    submitLogin(e);
});


// capitalize first letter of a string
function capitalize(s) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}

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
    // show message on failed login
    if (r.ok) {
        window.location.href = "/";
    } else {
        const data = await r.json();
        document.getElementById("failed-login-message").textContent = capitalize(data.message);
    }
}
