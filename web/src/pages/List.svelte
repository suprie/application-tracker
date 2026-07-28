<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { link } from 'svelte-spa-router'

  let items = []
  let loading = true
  let error = null
  let status = ''

  const statuses = [
    { value: '', label: 'All' },
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
      const r = await api.listJds(status)
      items = r.items
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }

  onMount(load)
</script>

<div class="mb-4 flex items-center justify-between">
  <h1 class="text-xl font-semibold text-slate-800">Applications</h1>
  <a href="/jds/new" use:link class="rounded bg-slate-800 px-3 py-1.5 text-sm text-white hover:bg-slate-700">+ Add JD</a>
</div>

<div class="mb-4 flex items-center gap-3 text-sm">
  <label for="status" class="text-slate-600">Status:</label>
  <select id="status" bind:value={status} onchange={load} class="rounded border border-slate-300 px-2 py-1">
    {#each statuses as s}
      <option value={s.value}>{s.label}</option>
    {/each}
  </select>
  <button on:click={load} class="text-slate-500 underline">refresh</button>
</div>

{#if loading}
  <p class="text-slate-500">Loading…</p>
{:else if error}
  <p class="text-red-600">Error: {error}</p>
{:else if items.length === 0}
  <p class="text-slate-500">No applications yet. Click <strong>+ Add JD</strong> to paste a job description.</p>
{:else}
  <table class="w-full text-sm">
    <thead class="border-b text-left text-slate-500">
      <tr>
        <th class="py-2">Company</th>
        <th>Role</th>
        <th>Status</th>
        <th>Fit</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#each items as jd (jd.id)}
        <tr class="border-b border-slate-100">
          <td class="py-2">{jd.company || '—'}</td>
          <td>{jd.role_title || '—'}</td>
          <td>{jd.status}</td>
          <td>{jd.fit_score != null ? jd.fit_score + '%' : '—'}</td>
          <td class="text-right">
            <a href={`/jds/${jd.id}`} use:link class="text-slate-600 underline">open</a>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}
