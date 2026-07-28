<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { link, push } from 'svelte-spa-router'

  export let params = {}
  const isNew = !params.id
  const id = params.id

  let form = {
    name: '',
    website_url: '',
    industry: '',
    size: '',
    country: '',
    notes: '',
    source: '',
    research_summary: '',
  }
  let busy = false
  let error = null

  const fields = [
    ['name', 'Name'],
    ['website_url', 'Website URL'],
    ['industry', 'Industry'],
    ['size', 'Size'],
    ['country', 'Country'],
    ['source', 'Source'],
    ['research_summary', 'Research summary'],
    ['notes', 'Notes'],
  ]

  onMount(async () => {
    if (isNew) return
    try {
      const c = await api.getCompany(id)
      for (const k of Object.keys(form)) form[k] = c[k] || ''
    } catch (e) {
      error = e.message
    }
  })

  async function save() {
    busy = true
    error = null
    try {
      if (isNew) {
        await api.createCompany(form)
      } else {
        await api.updateCompany(id, form)
      }
      push('/companies')
    } catch (e) {
      error = e.message
    } finally {
      busy = false
    }
  }
</script>

<p class="mb-3"><a href="/companies" use:link class="text-sm text-slate-500 underline">← back to companies</a></p>

<h1 class="mb-4 text-xl font-semibold text-slate-800">{isNew ? 'Add company' : 'Edit company'}</h1>

{#if error}
  <p class="mb-3 text-red-600">{error}</p>
{/if}

<form on:submit|preventDefault={save} class="max-w-xl space-y-3">
  {#each fields as [key, label]}
    <div>
      <label for={key} class="block text-sm text-slate-600">{label}{key === 'name' ? ' *' : ''}</label>
      {#if key === 'research_summary' || key === 'notes'}
        <textarea id={key} bind:value={form[key]} rows="3" class="mt-1 w-full rounded border border-slate-300 p-2 text-sm"></textarea>
      {:else}
        <input id={key} bind:value={form[key]} class="mt-1 w-full rounded border border-slate-300 p-2" />
      {/if}
    </div>
  {/each}
  <button disabled={busy || (isNew && !form.name)} class="rounded bg-slate-800 px-4 py-2 text-sm text-white disabled:opacity-50">
    {busy ? 'Saving…' : 'Save'}
  </button>
</form>
