// search.js — halaman /reports
let curPage = 1;
const LIMIT = 9;
async function loadCategories() {
  const sel = document.getElementById('fCategory');
  if (!sel) return;
  const r = await api('/api/categories');
  if (r.success) r.data.forEach(c => {
    const o = document.createElement('option');
    o.value = c.id; o.textContent = c.name; sel.appendChild(o);
  });
}
function baseType() {
  if (location.pathname.includes('/lost')) return 'LOST';
  if (location.pathname.includes('/found')) return 'FOUND';
  return '';
}
async function doSearch(page = 1) {
  curPage = page;
  const grid = document.getElementById('resultGrid');
  const empty = document.getElementById('emptyState');
  const pag = document.getElementById('pagination');
  grid.innerHTML = '<p style="color:var(--text2)">Memuat...</p>';
  const p = new URLSearchParams({
    page, limit: LIMIT,
    search: document.getElementById('fSearch')?.value || '',
    status: document.getElementById('fStatus')?.value || '',
    category_id: document.getElementById('fCategory')?.value || '',
    location: document.getElementById('fLocation')?.value || '',
    sort: document.getElementById('fSort')?.value || 'newest',
  });
  const t = baseType() || document.getElementById('fType')?.value || '';
  if (t) p.set('type', t);
  const r = await api('/api/reports?' + p.toString());
  if (!r.success) { grid.innerHTML = ''; empty.style.display = 'block'; pag.innerHTML = ''; return; }
  const items = r.data || [];
  if (!items.length) { grid.innerHTML = ''; empty.style.display = 'block'; }
  else {
    empty.style.display = 'none';
    grid.innerHTML = items.map(it => `
      <div class="card">
        <img src="${imgSrc(it.photo)}" alt="${it.item_name}" loading="lazy">
        <div class="body">
          <div class="badges">${badge(it.type)} ${badge(it.status)} <span class="badge">${it.category?.name || '-'}</span></div>
          <div class="title">${it.item_name}</div>
          <div class="meta"><span>${icon('pin')}${it.location || '-'}</span><span>${icon('calendar')}${it.event_date || fmtDate(it.created_at)}</span></div>
          <a class="btn btn-sm" href="/reports/${it.id}">Detail</a>
        </div>
      </div>`).join('');
  }
  const meta = r.meta || { pages: 1, page: 1 };
  pag.innerHTML = '';
  const btn = (label, pg, active = false, dis = false) => {
    const b = document.createElement('button');
    b.textContent = label;
    if (active) b.className = 'active';
    b.disabled = dis;
    b.onclick = () => doSearch(pg);
    pag.appendChild(b);
  };
  btn('Previous', Math.max(1, meta.page - 1), false, meta.page <= 1);
  for (let i = 1; i <= (meta.pages || 1); i++) btn(String(i), i, i === meta.page);
  btn('Next', (meta.page || 1) + 1, false, meta.page >= meta.pages);
}
document.addEventListener('DOMContentLoaded', () => {
  if (!document.getElementById('resultGrid')) return;
  loadCategories();
  doSearch(1);
  ['fSearch', 'fStatus', 'fCategory', 'fLocation', 'fSort', 'fType'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener(id === 'fSearch' ? 'input' : 'change', () => doSearch(1));
  });
});
