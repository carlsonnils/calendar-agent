<script>
  import { marked } from 'marked'

  export let type   // 'user' | 'chat'
  export let text

  function formatDateNow() {
    const d = new Date()
    const pad = n => String(n).padStart(2, '0')
    return {
      date: `${d.getFullYear()}-${pad(d.getMonth())}-${pad(d.getDate())}`,
      time: `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    }
  }

  const { date, time } = formatDateNow()
</script>

<div class="message-block {type === 'user' ? 'user-block' : 'chat-block'}">
  <div class="time-block {type === 'user' ? 'user-time' : 'chat-time'}">
    <div class="time-part">{time}</div>
    <div class="date-part">{date}</div>
  </div>
  <div class="message-text {type === 'user' ? 'user-message' : 'chat-message'}">
    {#if type === 'chat'}
      {@html marked.parse(text)}
    {:else}
      {text}
    {/if}
  </div>
</div>
