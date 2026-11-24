<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'

  const providers = [
    { value: 'openai', label: 'OpenAI-compatible API (LM Studio, OpenAI, etc.)' },
    { value: 'deepseek', label: 'DeepSeek API' },
    { value: 'ollama', label: 'Ollama' },
    { value: 'claude-harness', label: 'Claude CLI (claude -p)' },
    { value: 'codex-harness', label: 'Codex CLI (codex exec)' },
  ]

  let form = { provider: 'openai', model: '', base_url: '' }
  let apiKeyInput = ''
  let apiKeySet = false
  let clearKey = false
  let failedToLoad = false
  let savedProvider = ''
  let savedBaseURL = ''
  let updatedAt = null
  let loading = true
  let busy = false
  let error = null
  let saved = false

  $: isHarness = form.provider.endsWith('-harness')
  $: sameDestination = form.provider === savedProvider && form.base_url === savedBaseURL
  $: canKeepKey = apiKeySet && sameDestination

  onMount(async () => {
    try {
      const s = await api.getLLMSettings()
      form.provider = s.provider || 'openai'
      form.model = s.model || ''
      form.base_url = s.base_url || ''
      savedProvider = s.provider
      savedBaseURL = s.base_url
      apiKeySet = s.api_key_set
      updatedAt = s.updated_at
    } catch (e) {
      error = e.message
      failedToLoad = true
    } finally {
      loading = false
    }
  })

  async function save() {
    busy = true
    error = null
    saved = false
    try {
      const body = { provider: form.provider, model: form.model, base_url: form.base_url }
      if (!isHarness) {
        if (canKeepKey && clearKey) body.api_key = ''
        else if (apiKeyInput) body.api_key = apiKeyInput
      }
      const s = await api.updateLLMSettings(body)
      savedProvider = s.provider
      savedBaseURL = s.base_url
      apiKeySet = s.api_key_set
      updatedAt = s.updated_at
      apiKeyInput = ''
      clearKey = false
      saved = true
    } catch (e) {
      error = e.message
    } finally {
      busy = false
    }
  }
</script>

<h1 class="mb-4 text-xl font-semibold text-slate-800">LLM settings</h1>

{#if loading}
  <p class="text-slate-600" role="status">Loading…</p>
{:else}
  {#if error}
    <p class="mb-3 text-red-700" role="alert">{error}</p>
  {/if}
  {#if saved}
    <p class="mb-3 text-green-700" role="status">Saved.</p>
  {/if}

  {#if !failedToLoad && !updatedAt}
    <p class="mb-4 text-sm text-slate-600">The server is using environment settings. Saving here sets one configuration for all pipeline stages.</p>
  {/if}

  <form on:submit|preventDefault={save} class="max-w-xl">
    <fieldset disabled={busy || failedToLoad} class="space-y-3">
      <div>
        <label for="provider" class="block text-sm text-slate-600">Provider</label>
        <select id="provider" bind:value={form.provider} class="mt-1 w-full rounded border border-slate-500 p-2">
          {#each providers as p}
            <option value={p.value}>{p.label}</option>
          {/each}
        </select>
      </div>

      {#if isHarness}
        <p class="text-sm text-slate-500">
          Install and sign in to {form.provider === 'claude-harness' ? 'Claude Code' : 'Codex'} on the machine running the tracker.
        </p>
      {:else}
        <div>
          <label for="model" class="block text-sm text-slate-600">Model</label>
          <input id="model" bind:value={form.model} placeholder="gemma-4-12b-qat" class="mt-1 w-full rounded border border-slate-500 p-2" />
        </div>
        <div>
          <label for="base_url" class="block text-sm text-slate-600">Base URL</label>
          <input id="base_url" bind:value={form.base_url} placeholder="http://localhost:1234/v1" class="mt-1 w-full rounded border border-slate-500 p-2" />
        </div>
        <div>
          <label for="api_key" class="block text-sm text-slate-600">API key</label>
          <input
            id="api_key"
            type="password"
            bind:value={apiKeyInput}
            autocomplete="new-password"
            disabled={canKeepKey && clearKey}
            placeholder={canKeepKey ? 'Leave blank to keep the saved key' : ''}
            class="mt-1 w-full rounded border border-slate-500 p-2"
          />
          {#if canKeepKey}
            <label class="mt-2 inline-flex items-center gap-2 text-sm text-slate-700">
              <input type="checkbox" bind:checked={clearKey} />
              Clear saved API key
            </label>
          {:else if apiKeySet}
            <p class="mt-2 text-sm text-slate-600">Saving a different provider or URL clears the old key.</p>
          {/if}
        </div>
      {/if}

      {#if updatedAt}
        <p class="text-xs text-slate-600">Last updated {new Date(updatedAt).toLocaleString()}</p>
      {/if}

      <button disabled={busy || failedToLoad} class="rounded bg-slate-800 px-4 py-2 text-sm text-white disabled:opacity-50">
        {busy ? 'Saving…' : 'Save'}
      </button>
    </fieldset>
  </form>
{/if}
