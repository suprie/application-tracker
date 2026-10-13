<script>
  import { api, pollTask } from '../lib/api.js'

  let file = null
  let busy = false
  let status = ''
  let error = null
  let done = false

  function onFileChange(e) {
    file = e.target.files[0] || null
    done = false
    error = null
  }

  async function submit() {
    if (!file) return
    busy = true
    error = null
    done = false
    status = 'Uploading…'
    try {
      const r = await api.uploadCvTask(file)
      const t = await pollTask(r.task_id, { onUpdate: (x) => (status = `Parsing… (${x.state})`) })
      if (t.state === 'failed') throw new Error(t.error || 'parse failed')
      done = true
    } catch (e) {
      error = e.message
    } finally {
      busy = false
    }
  }
</script>

<h1 class="mb-4 text-xl font-semibold text-slate-800">Master profile</h1>

<p class="mb-4 text-sm text-slate-600">
  Upload your CV as a PDF to (re)build the master profile used for matching, ranking, and cover letters.
  This overwrites the existing profile.
</p>

{#if error}
  <p class="mb-3 text-red-600">{error}</p>
{/if}
{#if done}
  <p class="mb-3 text-green-700">Profile saved.</p>
{/if}

<form on:submit|preventDefault={submit} class="space-y-3">
  <div>
    <label for="cv" class="block text-sm text-slate-600">CV (PDF)</label>
    <input id="cv" type="file" accept="application/pdf" on:change={onFileChange} class="mt-1 w-full text-sm" />
  </div>
  <button
    disabled={busy || !file}
    class="rounded bg-slate-800 px-4 py-2 text-sm text-white disabled:opacity-50"
  >
    {busy ? status : 'Upload & parse'}
  </button>
</form>
