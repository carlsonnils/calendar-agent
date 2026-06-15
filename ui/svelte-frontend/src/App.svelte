<script>
  import ConversationView from './components/ConversationView.svelte'
  import PromptInput from './components/PromptInput.svelte'

  let messages = []
  let loading = false

  async function handleSubmit(text) {
    messages = [...messages, { type: 'user', text }]
    loading = true

    const r = await fetch('/api/chat', {
      method: 'POST',
      body: JSON.stringify({ message: text })
    })

    if (r.status === 401) { window.location.href = '/login'; return }
    const d = await r.json()

    messages = [...messages, { type: 'chat', text: d.message }]
    loading = false
  }
</script>

<div class="app">
  <div class="app-body">
    <ConversationView {messages} {loading} />
    <PromptInput on:submit={e => handleSubmit(e.detail)} />
  </div>
</div>
