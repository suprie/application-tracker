// Thin fetch wrapper for the tracker API + a task-polling helper.

async function request(path, options = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`
    try {
      const j = await res.json()
      if (j.error) msg = j.error
    } catch (_) {
      /* non-JSON error body */
    }
    throw new Error(msg)
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  // JDs (applications)
  listJds: (status) => request('/api/jds' + (status ? `?status=${encodeURIComponent(status)}` : '')),
  getJd: (id) => request(`/api/jds/${id}`),
  patchJd: (id, body) => request(`/api/jds/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  applyJd: (id) => request(`/api/jds/${id}/apply`, { method: 'POST' }),
  createJdTask: (text, apply_url) => request('/api/jds', { method: 'POST', body: JSON.stringify({ text, apply_url }) }),
  matchTask: (id) => request(`/api/jds/${id}/match`, { method: 'POST' }),
  rankTask: (id) => request(`/api/jds/${id}/rank`, { method: 'POST' }),
  coverLetterTask: (id) => request(`/api/jds/${id}/cover-letter`, { method: 'POST' }),

  // Companies
  listCompanies: (q) => request('/api/companies' + (q ? `?q=${encodeURIComponent(q)}` : '')),
  getCompany: (id) => request(`/api/companies/${id}`),
  createCompany: (body) => request('/api/companies', { method: 'POST', body: JSON.stringify(body) }),
  updateCompany: (id, body) => request(`/api/companies/${id}`, { method: 'PUT', body: JSON.stringify(body) }),

  // Tasks
  getTask: (id) => request(`/api/tasks/${id}`),
  cancelTask: (id) => request(`/api/tasks/${id}`, { method: 'DELETE' }),

  // LLM settings
  getLLMSettings: () => request('/api/settings/llm'),
  updateLLMSettings: (body) => request('/api/settings/llm', { method: 'PUT', body: JSON.stringify(body) }),
}

// pollTask calls onUpdate with each task snapshot, resolving on done/failed
// or rejecting on timeout. LLM actions can take a while, so default is 3 min.
export async function pollTask(taskId, { onUpdate, interval = 1000, timeoutMs = 180000 } = {}) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    const t = await api.getTask(taskId)
    if (onUpdate) onUpdate(t)
    if (t.state === 'done' || t.state === 'failed') return t
    await new Promise((r) => setTimeout(r, interval))
  }
  throw new Error('task timed out')
}
