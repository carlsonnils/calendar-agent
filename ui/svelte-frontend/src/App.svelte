<script>
    import ConversationView from "./components/ConversationView.svelte";
    import PromptInput from "./components/PromptInput.svelte";
    import LoginForm from "./components/LoginForm.svelte";

    let messages = [];
    let loading = false;
    let loggedIn = document.cookie.split('; ').some(row => row.startsWith('is_authed='));
    let submittedPrompt = false;

    async function handleSubmit(text) {
        submittedPrompt = true;
        messages = [...messages, { type: "user", text }];
        loading = true;

        const r = await fetch("/api/chat", {
            method: "POST",
            body: JSON.stringify({ message: text }),
        });

        if (r.status === 401) {
            loggedIn = false;
            return;
        }
        const d = await r.json();

        messages = [...messages, { type: "chat", text: d.message }];
        loading = false;
    }
</script>

<div class="app">
    {#if loggedIn || !submittedPrompt}
        <div class="app-body">
            <ConversationView {messages} {loading} />
            <PromptInput onSubmit={handleSubmit} />
        </div>
    {:else}
        <LoginForm {loggedIn} />        
    {/if}
</div>
