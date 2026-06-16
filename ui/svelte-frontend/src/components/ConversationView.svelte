<script>
    import MessageBlock from "./MessageBlock.svelte";
    import { afterUpdate } from "svelte";

    export let messages = [];
    export let loading = false;

    let convoEl;

    afterUpdate(() => {
        if (convoEl) convoEl.scrollTop = convoEl.scrollHeight;
    });
</script>

<div class="conversation" bindthis={convoEl}>
    {#if messages.length === 0}
        <div class="welcome-message">
            <p>Knowledge is power</p>
            <p>Ingenuity is freedom</p>
        </div>
    {:else}
        <div class="chat-begining"></div>
        <div class="messages">
            {#each messages as msg}
                <MessageBlock type={msg.type} text={msg.text} />
            {/each}
        </div>
    {/if}

    {#if loading}
        <span class="loader-animation"></span>
    {/if}
</div>
