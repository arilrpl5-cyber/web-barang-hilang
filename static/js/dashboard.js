// dashboard.js — user dashboard: recent reports + possible matches
document.addEventListener('DOMContentLoaded', async () => {
  const recentEl = document.getElementById('recentTable');
  if (recentEl) {
    const r = await api('/api/my-reports?limit=5');
    if (r.success && r.data.length) {
      recentEl.innerHTML = r.data.map(it => `<tr><td>${it.item_name}</td><td>${badge(it.type)}</td><td>${fmtDate(it.created_at)}</td><td>${badge(it.status)}</td><td><a class="btn btn-sm" href="/reports/${it.id}">Detail</a></td></tr>`).join('');
    } else recentEl.innerHTML = '<tr><td colspan="5" style="color:var(--text2)">Belum ada laporan. <a href="/reports/create" style="text-decoration:underline">Buat Laporan</a></td></tr>';
  }
  const matchEl = document.getElementById('matchList');
  if (matchEl) {
    const r = await api('/api/my-matches');
    if (r.success && r.data.length) {
      matchEl.innerHTML = r.data.map(m => {
        const f = m.found_report || {};
        return `<div class="hv-card"><div style="flex:1"><b>${f.item_name || '-'}</b><div style="color:var(--text2);font-size:13px">Score: ${Math.round(m.score)}% • ${f.location || '-'}</div></div><a class="btn btn-sm" href="/reports/${f.id || m.found_report_id}">Lihat</a></div>`;
      }).join('') + '<p style="color:var(--muted);font-size:12px">Skor hanya merupakan rekomendasi dan bukan bukti kepemilikan.</p>';
    } else matchEl.innerHTML = '<p style="color:var(--text2)">Belum ada barang yang kemungkinan cocok.</p>';
  }
  // admin dashboard widgets
  if (document.getElementById('adminStats')) {
    const r = await api('/api/admin/dashboard');
    if (r.success) {
      const s = r.data.stats;
      document.getElementById('adminStats').innerHTML = [
        ['Total User', s.total_users], ['Total Laporan', s.total_reports],
        ['Hilang', s.lost], ['Ditemukan', s.found],
        ['Dikembalikan', s.returned], ['Pending', s.pending], ['Klaim Aktif', s.claims],
      ].map(([l, v]) => `<div class="stat"><div class="num">${v}</div><div class="lbl">${l}</div></div>`).join('');
      const pm = r.data.per_month || {};
      const months = Object.keys(pm).sort();
      if (months.length && document.getElementById('chartBars')) {
        const max = Math.max(...Object.values(pm), 1);
        document.getElementById('chartBars').innerHTML = months.map(mo =>
          `<div style="flex:1;text-align:center"><div title="${pm[mo]}" style="background:#3a3a3a;border-radius:6px 6px 0 0;height:${Math.max(8, Math.round(pm[mo] / max * 120))}px"></div><div style="font-size:11px;color:var(--muted);margin-top:6px">${mo.slice(5)}</div></div>`
        ).join('');
      }
      if (document.getElementById('adminRecent')) {
        document.getElementById('adminRecent').innerHTML =
          (r.data.recent_reports || []).map(it => `<tr><td>${it.item_name}</td><td>${badge(it.type)}</td><td>${badge(it.status)}</td><td>${fmtDate(it.created_at)}</td></tr>`).join('') || '<tr><td colspan="4">Belum ada data</td></tr>';
      }
      if (document.getElementById('adminActivity')) {
        document.getElementById('adminActivity').innerHTML =
          (r.data.recent_activity || []).map(a => `<div class="hv-card"><div><b>${a.action}</b><div style="color:var(--text2);font-size:13px">${a.description || ''} • ${fmtDate(a.created_at)}</div></div></div>`).join('') || '<p style="color:var(--text2)">Belum ada aktivitas.</p>';
      }
    }
  }
});
