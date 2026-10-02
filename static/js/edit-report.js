// edit-report.js — modal form edit laporan lengkap (dipakai my-reports & admin-reports)
// Butuh: api(), toast() dari app.js. Muat SETELAH app.js.
function escHtml(s) {
  return String(s == null ? '' : s)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

function closeEditModal() {
  var el = document.getElementById('editModalBg');
  if (el) el.remove();
}

async function openEditModal(reportId, onSaved) {
  closeEditModal();
  toast('Memuat data laporan...');
  var repRes, catRes;
  try {
    var res = await Promise.all([api('/api/reports/' + reportId), api('/api/categories')]);
    repRes = res[0]; catRes = res[1];
  } catch (e) { toast('Gagal memuat data'); return; }
  if (!repRes.success) { toast(repRes.message || 'Laporan tidak ditemukan'); return; }
  var r = repRes.data;
  var cats = catRes.success ? catRes.data : [];
  var catOpts = '<option value="">-- Pilih --</option>' + cats.map(function (c) {
    return '<option value="' + c.id + '"' + (String(c.id) === String(r.category_id) ? ' selected' : '') + '>' + escHtml(c.name) + '</option>';
  }).join('');

  var bg = document.createElement('div');
  bg.className = 'modal-bg';
  bg.id = 'editModalBg';
  bg.style.display = 'flex';
  bg.innerHTML =
    '<div class="modal" style="max-width:660px;width:94%;max-height:92vh;overflow:auto">'
    + '<span class="kicker"><span class="dot"></span> Edit Laporan #' + r.id + '</span>'
    + '<h3 style="margin:14px 0 4px;font-size:22px">Edit Laporan</h3>'
    + '<p style="color:var(--text2);font-size:13px;margin:0 0 6px">Perbarui data barang, lalu simpan.</p>'
    + '<form id="editReportForm">'
    + '<div class="form-grid">'
    + '<div><label>Nama barang *</label><input name="item_name" required value="' + escHtml(r.item_name) + '"></div>'
    + '<div><label>Kategori</label><select name="category_id">' + catOpts + '</select></div>'
    + '<div><label>Merek</label><input name="brand" value="' + escHtml(r.brand) + '"></div>'
    + '<div><label>Warna</label><input name="color" value="' + escHtml(r.color) + '"></div>'
    + '<div><label>Lokasi</label><input name="location" value="' + escHtml(r.location) + '"></div>'
    + '<div><label>Tanggal</label><input name="event_date" type="date" value="' + escHtml(r.event_date) + '"></div>'
    + '<div><label>Waktu</label><input name="event_time" type="time" value="' + escHtml(r.event_time) + '"></div>'
    + '<div><label>Disimpan di (barang ditemukan)</label><input name="storage_location" value="' + escHtml(r.storage_location) + '"></div>'
    + '<div class="full"><label>Deskripsi</label><textarea name="description" rows="2">' + escHtml(r.description) + '</textarea></div>'
    + '<div class="full"><label>Ciri-ciri khusus</label><textarea name="characteristics" rows="2">' + escHtml(r.characteristics) + '</textarea></div>'
    + '<div class="full"><label>Catatan tambahan</label><textarea name="additional_note" rows="2">' + escHtml(r.additional_note) + '</textarea></div>'
    + '<div class="full"><label>Foto (kosongkan jika tidak diganti — JPG/PNG/WEBP, maks 5MB)</label>'
    + (r.photo ? '<img src="' + escHtml(r.photo) + '" alt="" style="display:block;max-width:160px;border-radius:12px;border:1px solid var(--line);margin-bottom:8px">' : '')
    + '<input type="file" name="photo" accept=".jpg,.jpeg,.png,.webp"></div>'
    + '</div>'
    + '<div style="display:flex;gap:10px;justify-content:flex-end;margin-top:18px">'
    + '<button type="button" class="btn" id="editCancel">Batal</button>'
    + '<button type="submit" class="btn btn-primary" id="editSave">Simpan Perubahan</button>'
    + '</div></form></div>';
  document.body.appendChild(bg);
  bg.addEventListener('click', function (e) { if (e.target === bg) closeEditModal(); });
  document.getElementById('editCancel').onclick = closeEditModal;
  document.getElementById('editReportForm').addEventListener('submit', async function (e) {
    e.preventDefault();
    var btn = document.getElementById('editSave');
    btn.textContent = 'Menyimpan...'; btn.disabled = true;
    var fd = new FormData(e.target);
    try {
      var resp = await fetch('/api/reports/' + reportId, { method: 'PUT', body: fd, credentials: 'include' });
      var data = await resp.json();
      if (data.success) {
        toast('Laporan berhasil diperbarui');
        closeEditModal();
        if (typeof onSaved === 'function') onSaved();
      } else {
        toast(data.message || 'Gagal menyimpan perubahan');
        btn.textContent = 'Simpan Perubahan'; btn.disabled = false;
      }
    } catch (err) {
      toast('Terjadi kesalahan jaringan');
      btn.textContent = 'Simpan Perubahan'; btn.disabled = false;
    }
  });
}
