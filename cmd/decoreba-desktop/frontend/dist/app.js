let allContexts = []
let results = []
let selectedIndex = 0
let currentContext = ''
let mode = 'search' // 'search' | 'create'
let searchTimer = null
let contextCycleIndex = -1

const searchInput = document.getElementById('search-input')
const resultsEl = document.getElementById('results')
const createForm = document.getElementById('create-form')
const contextBadge = document.getElementById('context-badge')
const footerHints = document.getElementById('footer-hints')
const createCtx = document.getElementById('create-context')
const createTitle = document.getElementById('create-title')
const createCmd = document.getElementById('create-command')

let backend = null

function getBackend() {
  if (backend) return backend
  if (window.go && window.go.main && window.go.main.App) {
    backend = window.go.main.App
  }
  return backend
}

async function call(method, ...args) {
  const b = getBackend()
  if (!b) return
  return b[method](...args)
}

function setFooter(text) {
  footerHints.textContent = text
}

function updateBadge(ctx) {
  contextBadge.textContent = ctx || 'all'
}

function updateContextCycle() {
  if (currentContext === '' || currentContext === 'all') {
    contextCycleIndex = -1
  } else {
    contextCycleIndex = allContexts.indexOf(currentContext)
  }
}

function renderSearch() {
  resultsEl.style.display = ''
  createForm.classList.add('hidden')

  if (!results || results.length === 0) {
    resultsEl.innerHTML =
      '<div id="empty-state">' +
      '  <span>No commands found</span>' +
      '  <button id="create-btn">+ create new command</button>' +
      '</div>'
    const btn = document.getElementById('create-btn')
    if (btn) btn.onclick = () => switchMode('create')
    setFooter('ctrl+enter new · tab context')
    return
  }

  if (selectedIndex >= results.length) selectedIndex = results.length - 1
  if (selectedIndex < 0) selectedIndex = 0

  let html = ''
  for (let i = 0; i < results.length; i++) {
    const r = results[i]
    const sel = i === selectedIndex ? ' selected' : ''
    const icon = getContextIcon(r.cmd.context)
    html +=
      '<div class="result-item' + sel + '" data-idx="' + i + '">' +
      '  <span class="context-icon">' + icon + '</span>' +
      '  <span class="text-block">' +
      '    <div class="title">' + esc(r.cmd.title) + '</div>' +
      '    <div class="command-line">' + esc(r.cmd.command) + '</div>' +
      '  </span>' +
      '  <span class="context-tag">' + esc(r.cmd.context) + '</span>' +
      '</div>'
  }
  resultsEl.innerHTML = html

  const items = resultsEl.querySelectorAll('.result-item')
  items.forEach((el) => {
    el.onclick = () => {
      const idx = parseInt(el.dataset.idx, 10)
      selectedIndex = idx
      copySelected()
    }
  })

  if (items.length > 0) {
    items[selectedIndex].scrollIntoView({ block: 'nearest' })
  }

  setFooter(getContextsText() + ' · enter copy · tab context · ctrl+enter new')
}

function getContextsText() {
  return results.length + ' result' + (results.length !== 1 ? 's' : '')
}

function switchMode(newMode) {
  mode = newMode
  if (mode === 'create') {
    resultsEl.style.display = 'none'
    createForm.classList.remove('hidden')
    const ctx = currentContext && currentContext !== 'all' ? currentContext : ''
    createCtx.value = ctx
    createTitle.value = ''
    createCmd.value = ''
    if (ctx) {
      createTitle.focus()
    } else {
      createCtx.focus()
    }
    setFooter('enter save · esc cancel')
    return
  }
  resultsEl.style.display = ''
  createForm.classList.add('hidden')
  searchInput.focus()
  doSearch()
}

async function doSearch() {
  if (mode === 'create') return
  const query = searchInput.value
  const ctx = currentContext === 'all' ? '' : currentContext
  try {
    const r = await call('Search', query, ctx)
    results = r || []
    selectedIndex = 0
    renderSearch()
  } catch (e) {
    console.error('search error', e)
  }
}

async function loadContexts() {
  try {
    allContexts = await call('GetContexts') || []
  } catch (e) {
    allContexts = []
  }
}

function cycleContext() {
  if (allContexts.length === 0) return
  contextCycleIndex++
  if (contextCycleIndex >= allContexts.length) {
    contextCycleIndex = -1
    currentContext = ''
  } else {
    currentContext = allContexts[contextCycleIndex]
  }
  updateBadge(currentContext || 'all')
  doSearch()
}

async function copySelected() {
  if (!results || results.length === 0 || selectedIndex < 0) return
  const cmd = results[selectedIndex].cmd
  try {
    await call('CopyCommand', cmd.id)
    showCopiedFeedback(selectedIndex)
  } catch (e) {
    console.error('copy error', e)
  }
}

function showCopiedFeedback(idx) {
  const items = resultsEl.querySelectorAll('.result-item')
  if (idx >= items.length) return
  const tag = items[idx].querySelector('.context-tag')
  if (!tag) return
  const orig = tag.textContent
  tag.textContent = 'copied'
  tag.style.color = 'var(--success)'
  setTimeout(() => {
    tag.textContent = orig
    tag.style.color = ''
  }, 800)
}

async function saveNewCommand() {
  const ctx = createCtx.value.trim()
  const title = createTitle.value.trim()
  const cmd = createCmd.value.trim()
  if (!ctx || !cmd) return
  try {
    await call('AddCommand', ctx, title, cmd)
    await loadContexts()
    searchInput.value = ''
    switchMode('search')
  } catch (e) {
    console.error('save error', e)
  }
}

function esc() {
  if (mode === 'create') {
    switchMode('search')
    return
  }
  if (searchInput.value) {
    searchInput.value = ''
    doSearch()
    return
  }
  try {
    window.runtime && window.runtime.Window && window.runtime.Window.Hide()
  } catch (_) {}
}

function getContextIcon(ctx) {
  const c = (ctx || '').toLowerCase()
  if (c.includes('git')) {
    return '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3v12"/><path d="M18 9a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/><path d="M6 21a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/><path d="M18 9a3 3 0 0 0-3 3v3"/></svg>'
  }
  if (c.includes('tmux') || c.includes('terminal') || c.includes('shell')) {
    return '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>'
  }
  if (c.includes('docker') || c.includes('container')) {
    return '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 6a2 2 0 1 0 0-4 2 2 0 0 0 0 4z"/><path d="M18 6a2 2 0 1 0 0-4 2 2 0 0 0 0 4z"/><path d="M4 12h16"/><path d="M4 16h16"/><path d="M4 20h16"/></svg>'
  }
  return '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>'
}

function escHtml(s) { return esc(s) }
function esc(s) {
  if (!s) return ''
  const d = document.createElement('div')
  d.textContent = s
  return d.innerHTML
}

searchInput.addEventListener('keydown', (e) => {
  if (mode === 'create') return

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (selectedIndex < results.length - 1) selectedIndex++
    else selectedIndex = 0
    renderSearch()
    return
  }

  if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (selectedIndex > 0) selectedIndex--
    else selectedIndex = results.length - 1
    renderSearch()
    return
  }

  if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
    e.preventDefault()
    switchMode('create')
    return
  }

  if (e.key === 'Enter') {
    e.preventDefault()
    if (results.length > 0 && selectedIndex >= 0) {
      copySelected()
    }
    return
  }

  if (e.key === 'Tab') {
    e.preventDefault()
    cycleContext()
    return
  }

  if (e.key === 'Escape') {
    e.preventDefault()
    esc()
    return
  }
})

searchInput.addEventListener('input', () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(doSearch, 40)
})

createCtx.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { e.preventDefault(); createTitle.focus() }
  if (e.key === 'Escape') { e.preventDefault(); esc() }
})

createTitle.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { e.preventDefault(); createCmd.focus() }
  if (e.key === 'Escape') { e.preventDefault(); esc() }
})

createCmd.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { e.preventDefault(); saveNewCommand() }
  if (e.key === 'Escape') { e.preventDefault(); esc() }
})

async function init() {
  await loadContexts()
  updateBadge('all')
  setFooter('loading...')
  await doSearch()
  searchInput.focus()
}

document.addEventListener('DOMContentLoaded', () => {
  if (getBackend()) {
    init()
  } else {
    setFooter('Wails runtime not ready')
    searchInput.focus()
    searchInput.addEventListener('input', () => {
      const q = searchInput.value.toLowerCase()
      resultsEl.innerHTML = '<div id="empty-state"><span>Wails backend not connected. Running in browser fallback.</span></div>'
    })
  }
})

window.decorebaShow = function() {
  searchInput.value = ''
  currentContext = ''
  updateBadge('all')
  doSearch()
  searchInput.focus()
}
