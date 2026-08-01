<script>
  import Router, { link } from 'svelte-spa-router'
  import List from './pages/List.svelte'
  import New from './pages/New.svelte'
  import Detail from './pages/Detail.svelte'
  import Companies from './pages/Companies.svelte'
  import CompanyDetail from './pages/CompanyDetail.svelte'
  import Settings from './pages/Settings.svelte'

  // Static routes are listed before their param siblings so they win.
  const routes = {
    '/': List,
    '/jds/new': New,
    '/jds/:id': Detail,
    '/companies/new': CompanyDetail,
    '/companies/:id': CompanyDetail,
    '/companies': Companies,
    '/settings': Settings,
  }
</script>

<header class="border-b border-slate-200 bg-white">
  <nav class="mx-auto flex max-w-5xl items-center gap-6 px-4 py-3 text-sm">
    <a href="/" use:link class="font-semibold text-slate-800">Application Tracker</a>
    <a href="/companies" use:link class="text-slate-600 hover:text-slate-900">Companies</a>
    <a href="/settings" use:link class="text-slate-600 hover:text-slate-900">Settings</a>
  </nav>
</header>

<main class="mx-auto max-w-5xl px-4 py-6">
  <Router {routes} />
</main>
