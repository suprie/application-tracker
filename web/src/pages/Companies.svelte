<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { link } from 'svelte-spa-router'

  let items = []
  let loading = true
  let error = null
  let q = ''

  async function load() {
    loading = true
    error = null
    try {
      const r = await api.listCompanies(q)
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
  <h1 class="text-xl font-semibold text-slate-800">Companies</h1>
  <a href="/companies/new" use:link class="rounded bg-slate-800 px-3 py-1.5 text-sm text-white hover:bg-slate-700">+ Add company</a>
</div>

<div class="mb-4 flex items-center gap-2 text-sm">
  <input bind:value={q} placeholder="search…" class="rounded border border-slate-300 px-2 py-1" />
  <button on:click={load} class="text-slate-500 underline">search</button>
</div>

{#if loading}
  <p class="text-slate-500">Loading…</p>
{:else if error}
  <p class="text-red-600">Error: {error}</p>
{:else if items.length === 0}
  <p class="text-slate-500">No companies yet.</p>
{:else}
  <ul class="divide-y divide-slate-100">
    {#each items as c (c.id)}
      <li class="py-2">
        <a href={`/companies/${c.id}`} use:link class="font-medium text-slate-800">{c.name}</a>
        <span class="ml-2 text-sm text-slate-500">{c.industry || ''} {c.size ? '· ' + c.size : ''}</span>
        {#if c.research_summary}
          <p class="text-sm text-slate-500">{c.research_summary}</p>
        {/if}
      </li>
    {/each}
  </ul>
{/if}
