// app.js — helpers global
async function api(path, opts = {}) {
  const res = await fetch(path, { credentials: 'include', headers: { 'Content-Type': 'application/json' }, ...opts,
    headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) } });
  const data = await res.json().catch(() => ({ success: false, message: 'Response tidak valid' }));
  if (!res.ok && res.status === 401 && !location.pathname.startsWith('/login')) {
    // hanya redirect untuk halaman protected
    if (['/dashboard','/my-reports','/claims','/notifications','/profile','/admin'].some(p => location.pathname.startsWith(p))) {
      location.href = '/login';
    }
  }
  return { status: res.status, ...data };
}
function toast(msg, type = 'info') {
  let wrap = document.querySelector('.toast');
  if (!wrap) { wrap = document.createElement('div'); wrap.className = 'toast'; document.body.appendChild(wrap); }
  const el = document.createElement('div');
  el.className = 'toast-item';
  el.textContent = msg;
  wrap.appendChild(el);
  setTimeout(() => el.remove(), 3200);
}
function badge(status) { return `<span class="badge b-${status}">${status}</span>`; }
function fmtDate(s) { if (!s) return '-'; try { return new Date(s).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }); } catch { return s; } }
function imgSrc(p) { return p || '/static/images/placeholder.svg'; }
async function refreshNotifBadge() {
  const el = document.getElementById('notifBadge');
  if (!el) return;
  try {
    const r = await api('/api/notifications/unread-count');
    const n = r.data?.unread ?? r.unread ?? 0;
    el.style.display = n > 0 ? 'inline-flex' : 'none';
    el.textContent = n;
  } catch { /* abaikan */ }
}
async function logout() {
  await api('/api/logout', { method: 'POST' });
  location.href = '/';
}
// Sinkronkan navbar publik (Login/Daftar -> Bel/Dasbor/Logout) sesuai status login.
// Dipakai di halaman publik: index, reports, report-detail, about.
async function initPublicNav() {
  const nav = document.getElementById('navAuth');
  if (!nav || nav.dataset.done) return;
  nav.dataset.done = '1';
  try {
    const r = await api('/api/profile');
    if (!r.success) return; // tamu: biarkan tombol Login/Daftar
    const isAdmin = r.data.role === 'ADMIN';
    const bell = (typeof icon === 'function' ? icon('bell') : '');
    nav.innerHTML =
      `<a class="icon-btn" href="/notifications" title="Notifikasi">${bell}<span id="notifBadge" style="display:none;position:absolute;top:-4px;right:-4px;background:#f4f4f5;color:#0a0a0b;font-size:10px;font-weight:700;border-radius:12px;padding:1px 6px">0</span></a>`
      + `<a class="btn btn-sm" href="${isAdmin ? '/admin' : '/dashboard'}">Dashboard</a>`
      + `<button class="btn btn-sm" onclick="logout()">Logout</button>`;
    refreshNotifBadge();
  } catch { /* abaikan, tetap mode tamu */ }
}
document.addEventListener('DOMContentLoaded', () => {
  const nav = document.querySelector('.navbar');
  if (nav) addEventListener('scroll', () => nav.classList.toggle('scrolled', scrollY > 10));
  refreshNotifBadge();
  const ham = document.getElementById('hamburger');
  if (ham) ham.addEventListener('click', () => {
    const sb = document.querySelector('.sidebar');
    if (sb) sb.classList.toggle('open');
    const nl = document.querySelector('.nav-links');
    if (nl) nl.classList.toggle('open');
  });
});
function confirmModal(title, onYes) {
  let bg = document.getElementById('confirmBg');
  if (!bg) {
    bg = document.createElement('div');
    bg.id = 'confirmBg'; bg.className = 'modal-bg';
    bg.innerHTML = `<div class="modal"><h3 id="cfTitle"></h3><p style="color:var(--text2);font-size:14px">Tindakan ini tidak dapat dibatalkan.</p><div style="display:flex;gap:10px;justify-content:flex-end;margin-top:18px"><button class="btn" id="cfNo">Batal</button><button class="btn btn-primary" id="cfYes">Hapus</button></div></div>`;
    document.body.appendChild(bg);
  }
  document.getElementById('cfTitle').textContent = title;
  bg.style.display = 'flex';
  document.getElementById('cfNo').onclick = () => bg.style.display = 'none';
  document.getElementById('cfYes').onclick = () => { bg.style.display = 'none'; onYes(); };
}
