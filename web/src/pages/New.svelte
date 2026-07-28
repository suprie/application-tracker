<script>
  import { api, pollTask } from '../lib/api.js'
  import { push } from 'svelte-spa-router'

  let text = ''
  let apply_url = ''
  let busy = false
  let status = ''
  let error = null

  async function submit() {
    if (!text.trim()) return
    busy = true
    error = null
    status = 'Parsing JD…'
    try {
      const r = await api.createJdTask(text, apply_url)
      const t = await pollTask(r.task_id, { onUpdate: (x) => (status = `Parsing… (${x.state})`) })
      if (t.state === 'failed') throw new Error(t.error || 'parse failed')
      push(`/jds/${t.result.jd_id}`)
    } catch (e) {
      error = e.message
      busy = false
    }
  }
</script>

<h1 class="mb-4 text-xl font-semibold text-slate-800">Add job description</h1>

{#if error}
  <p class="mb-3 text-red-600">{error}</p>
{/if}

<form on:submit|preventDefault={submit} class="space-y-3">
  <div>
    <label for="text" class="block text-sm text-slate-600">Paste the job description text</label>
    <textarea
      id="text"
      bind:value={text}
      rows="16"
      class="mt-1 w-full rounded border border-slate-300 p-2 font-mono text-sm"
      placeholder="Paste the full job description here…"
    ></textarea>
  </div>
  <div>
    <label for="url" class="block text-sm text-slate-600">Apply URL (optional)</label>
    <input id="url" bind:value={apply_url} class="mt-1 w-full rounded border border-slate-300 p-2" placeholder="https://…" />
  </div>
  <button
    disabled={busy || !text.trim()}
    class="rounded bg-slate-800 px-4 py-2 text-sm text-white disabled:opacity-50"
  >
    {busy ? status : 'Parse & save'}
  </button>
</form>
