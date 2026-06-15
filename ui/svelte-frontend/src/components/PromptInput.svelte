<script>
  import { createEventDispatcher } from 'svelte'
  import AudioRecorder from './AudioRecorder.svelte'

  const dispatch = createEventDispatcher()

  let text = ''
  let showRecorder = false

  function submit() {
    if (!text.trim()) return
    dispatch('submit', text)
    text = ''
  }

  function handleKeypress(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }
</script>

<div class="prompt">
  <div class="prompt-input">
    {#if showRecorder}
      <AudioRecorder on:close={() => showRecorder = false} />
    {:else}
      <div
        class="prompt-text"
        contenteditable="true"
        bind:textContent={text}
        on:keypress={handleKeypress}
      ></div>
    {/if}
  </div>
  <div class="prompt-buttons">
    {#if !showRecorder}
      <button id="activate-voice" on:click={() => showRecorder = true}>
        <img src="/microphone.png" alt="mic" />
      </button>
    {/if}
    <button id="submit-prompt" on:click={submit}>Submit</button>
  </div>
</div>
