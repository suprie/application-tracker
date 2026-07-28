<script>
  import { onMount } from 'svelte'
  import { api, pollTask } from '../lib/api.js'
  import { link } from 'svelte-spa-router'

  export let params = {}
  const id = params.id

  let jd = null
  let loading = true
  let error = null
  let taskState = {}

  const statuses = [
    { value: 'draft', label: 'Draft' },
    { value: 'fitmatch', label: 'Fit match' },
    { value: 'applied', label: 'Applied' },
    { value: 'rejected', label: 'Rejected' },
    { value: 'offer', label: 'Offer' },
  ]

  async function load() {
    loading = true
    error = null
    try {
      jd = await api.getJd(id)
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }

  onMount(load)

  async function runAction(kind, fn) {
    taskState = { ...taskState, [kind]: 'pending' }
    try {
      const r = await fn()
      await pollTask(r.task_id, {
        onUpdate: (t) => (taskState = { ...taskState, [kind]: t.state }),
      })
      await load()
      taskState = { ...taskState, [kind]: 'done' }
    } catch (e) {
      taskState = { ...taskState, [kind]: 'failed: ' + e.message }
    }
  }

  async function changeStatus(e) {
    try {
      await api.patchJd(id, { status: e.target.value })
      await load()
    } catch (err) {
      error = err.message
    }
  }

  async function apply() {
    try {
      await api.applyJd(id)
      await load()
    } catch (e) {
      error = e.message
    }
  }

  function norm(s) {
    return (s || '').toLowerCase().replace(' ', '')
  }
</script>

<p class="mb-3"><a href="/" use:link class="text-sm text-slate-500 underline">← back to list</a></p>

{#if loading}
  <p class="text-slate-500">Loading…</p>
{:else if error}
  <p class="text-red-600">Error: {error}</p>
{:else if jd}
  <h1 class="text-xl font-semibold text-slate-800">
    {jd.company || '—'} — {jd.role_title || '—'}
  </h1>

  <div class="mt-2 flex flex-wrap items-center gap-3 text-sm text-slate-600">
    <select value={norm(jd.status)} on:change={changeStatus} class="rounded border border-slate-300 px-2 py-1">
      {#each statuses as s}
        <option value={s.value} selected={norm(jd.status) === s.value}>{s.label}</option>
      {/each}
    </select>
    {#if jd.fit_score != null}
      <span class="rounded bg-emerald-50 px-2 py-0.5 text-emerald-700">Fit {jd.fit_score}%</span>
    {/if}
    {#if jd.apply_url}
      <a href={jd.apply_url} target="_blank" class="text-blue-600 underline">apply link ↗</a>
    {/if}
    <button on:click={apply} class="rounded border border-slate-300 px-2 py-1">Mark applied</button>
  </div>

  {#if jd.fit_summary}
    <p class="mt-3 rounded bg-slate-50 p-3 text-sm text-slate-700">{jd.fit_summary}</p>
  {/if}

  <div class="mt-6 flex flex-wrap gap-2">
    <button
      on:click={() => runAction('match', () => api.matchTask(id))}
      class="rounded bg-slate-800 px-3 py-1.5 text-sm text-white"
    >Run match</button>
    <button
      on:click={() => runAction('rank', () => api.rankTask(id))}
      class="rounded bg-slate-800 px-3 py-1.5 text-sm text-white"
    >Run ranker</button>
    <button
      on:click={() => runAction('cover_letter', () => api.coverLetterTask(id))}
      class="rounded bg-slate-800 px-3 py-1.5 text-sm text-white"
    >Generate cover letter</button>
  </div>

  {#if Object.keys(taskState).length > 0}
    <ul class="mt-3 text-sm text-slate-600">
      {#each Object.keys(taskState) as k}
        <li>{k}: {taskState[k]}</li>
      {/each}
    </ul>
  {/if}

  {#if taskState.cover_letter === 'done' || jd.id}
    <p class="mt-2 text-sm">
      <a href={`/api/jds/${id}/cover-letter.pdf`} target="_blank" class="text-blue-600 underline">download cover letter (PDF)</a>
    </p>
  {/if}

  <section class="mt-6 grid grid-cols-1 gap-4 md:grid-cols-2">
    <div>
      <h2 class="mb-1 text-sm font-semibold uppercase text-slate-500">Requirements</h2>
      {#if jd.requirements}
        <ul class="list-disc pl-5 text-sm">
          {#each jd.requirements.must_have || [] as r}<li>{r}</li>{/each}
        </ul>
        {#if jd.requirements.nice_have && jd.requirements.nice_have.length}
          <p class="mt-1 text-xs text-slate-500">Nice to have: {jd.requirements.nice_have.join(', ')}</p>
        {/if}
      {/if}
    </div>
    <div>
      <h2 class="mb-1 text-sm font-semibold uppercase text-slate-500">Responsibilities</h2>
      <ul class="list-disc pl-5 text-sm">
        {#each jd.responsibilities || [] as r}<li>{r}</li>{/each}
      </ul>
    </div>
  </section>

  {#if jd.keywords && jd.keywords.length}
    <div class="mt-4">
      <h2 class="mb-1 text-sm font-semibold uppercase text-slate-500">Keywords</h2>
      <div class="flex flex-wrap gap-1">
        {#each jd.keywords as k}<span class="rounded bg-slate-100 px-2 py-0.5 text-xs">{k}</span>{/each}
      </div>
    </div>
  {/if}

  {#if jd.ranker_result}
    <details class="mt-4">
      <summary class="cursor-pointer text-sm font-semibold text-slate-600">Ranker result</summary>
      <pre class="mt-1 overflow-auto rounded bg-slate-900 p-3 text-xs text-slate-100">{JSON.stringify(jd.ranker_result, null, 2)}</pre>
    </details>
  {/if}
{/if}
