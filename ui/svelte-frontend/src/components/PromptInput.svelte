<script>
    let { onSubmit } = $props();

    let text = "";
    let showRecorder = false;

    function handleKeyPress(e) {
        if (e.getModifierState("Shift")) {
            return;
        }
        if (e.key === "Enter") {
            e.preventDefault();
            onSubmit(text);
        }
    }
</script>

<div class="prompt">
    <div class="prompt-input">
        <div
            class="prompt-text"
            contenteditable="true"
            bindtextContent={text}
            onkeypress={handleKeyPress}
        ></div>
    </div>
    <div class="prompt-buttons">
        <button id="submit-prompt" onclick={onSubmit(text)}>Submit</button>
    </div>
</div>
