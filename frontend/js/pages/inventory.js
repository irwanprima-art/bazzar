// Inventory Page
Router.register('inventory', async () => {
  return `
    <div class="toolbar">
      <select class="filter-select" id="inv-location-filter">
        <option value="">All Locations</option>
        <option value="EVENT">Event Floor</option>
        <option value="STORAGE">Storage</option>
      </select>
      <button class="btn btn-primary" onclick="showTransferModal()"><span class="material-symbols-rounded">sync_alt</span> Transfer</button>
      ${Auth.isAdmin() ? `
        <button class="btn btn-secondary" onclick="loadReplenishAlerts()"><span class="material-symbols-rounded">notification_important</span> Alerts</button>
      ` : ''}
      <button class="btn btn-secondary" onclick="loadSalesReport()"><span class="material-symbols-rounded">assessment</span> Sales Report</button>
    </div>
    <div id="inventory-table" class="card">Loading...</div>`;
});

function init_inventory() {
  loadInventory();
  document.getElementById('inv-location-filter')?.addEventListener('change', loadInventory);
}

async function loadInventory() {
  const loc = document.getElementById('inv-location-filter')?.value || '';
  try {
    const res = await API.get(`/inventory?event_id=${window.currentEventId}&location=${loc}`);
    document.getElementById('inventory-table').innerHTML = renderTable([
      { label: 'SKU', key: 'sku_code' },
      { label: 'Name', key: 'sku_name' },
      { label: 'Location', key: 'location_code' },
      { label: 'On Hand', render: r => `<strong>${r.qty_onhand}</strong>` },
      { label: 'Allocated', render: r => `<span style="color:var(--warning)">${r.qty_allocated}</span>` },
      { label: 'Available', render: r => `<strong style="color:${r.available <= 0 ? 'var(--danger)' : 'var(--success)'}">${r.available}</strong>` },
      { label: '', render: r => `<button class="btn btn-sm btn-primary" onclick="quickReplenish('${r.sku_id}','${r.sku_code}','${r.sku_name || ''}')">🔄 Transfer</button>` },
    ], res.data || [], 'No inventory data');
  } catch(e) { document.getElementById('inventory-table').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

// Transfer from top button: scan first, then show form
function showTransferModal() {
  Modal.show('🔄 Transfer Stock', `
    <div class="form-group">
      <label class="form-label">Scan / Ketik Barcode atau SKU Code</label>
      <div style="display:flex;gap:0.5rem;align-items:center">
        <input type="text" class="form-input" id="transfer-scan-input" placeholder="Scan barcode atau ketik SKU..." autofocus style="flex:1">
        <button class="scan-with-camera-btn" onclick="openTransferCamera()" title="Scan pakai Kamera">
          <span class="material-symbols-rounded">photo_camera</span>
        </button>
      </div>
    </div>
    <div id="transfer-scan-result" style="margin-top:0.5rem"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="doTransferScanLookup()">Cari</button>`);

  setTimeout(() => {
    const inp = document.getElementById('transfer-scan-input');
    inp?.focus();
    inp?.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') { e.preventDefault(); doTransferScanLookup(); }
    });
  }, 100);
}

function openTransferCamera() {
  CameraScanner.open((decodedText) => {
    const inp = document.getElementById('transfer-scan-input');
    if (inp) inp.value = decodedText;
    doTransferScanLookup();
  });
}

async function doTransferScanLookup() {
  const input = document.getElementById('transfer-scan-input')?.value?.trim();
  if (!input) { Toast.error('Masukkan barcode / SKU code'); return; }

  // Try to find SKU via barcode endpoint
  try {
    const res = await API.get(`/skus/barcode/${encodeURIComponent(input)}`);
    const sku = res.data;
    Modal.hide();
    showTransferFormForSku(sku.id, sku.sku_code, sku.name || '');
    return;
  } catch(e) {}

  // Try by sku_code via search
  try {
    const res = await API.get(`/skus?search=${encodeURIComponent(input)}&page_size=5`);
    const skus = res.data || [];
    const exact = skus.find(s => s.sku_code === input);
    if (exact) {
      Modal.hide();
      showTransferFormForSku(exact.id, exact.sku_code, exact.name || '');
      return;
    }
  } catch(e) {}

  document.getElementById('transfer-scan-result').innerHTML =
    `<div class="alert alert-danger"><span class="material-symbols-rounded">error</span> SKU/barcode "${input}" tidak ditemukan</div>`;
}

async function showTransferFormForSku(skuId, skuCode, skuName) {
  // Get stock info for this SKU
  let storageStock = 0, eventStock = 0;
  try {
    const res = await API.get(`/inventory?event_id=${window.currentEventId}`);
    (res.data || []).forEach(inv => {
      if (inv.sku_id === skuId) {
        if (inv.location_code === 'STORAGE') storageStock = inv.qty_onhand;
        if (inv.location_code === 'EVENT') eventStock = inv.qty_onhand;
      }
    });
  } catch(e) {}

  Modal.show(`🔄 Transfer: ${skuCode}`, `
    <p style="font-size:0.95rem;font-weight:700;color:var(--accent-light);margin-bottom:0.5rem">${skuName}</p>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;margin-bottom:1rem">
      <div class="stat-card info" style="padding:0.5rem;text-align:center">
        <div style="font-size:0.7rem;color:var(--text-muted)">Storage</div>
        <div style="font-size:1.2rem;font-weight:700">${storageStock}</div>
      </div>
      <div class="stat-card success" style="padding:0.5rem;text-align:center">
        <div style="font-size:0.7rem;color:var(--text-muted)">Event</div>
        <div style="font-size:1.2rem;font-weight:700">${eventStock}</div>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">Direction</label>
      <select class="form-select" id="transfer-direction">
        <option value="storage_to_event">Storage → Event Floor</option>
        <option value="event_to_storage">Event Floor → Storage</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label">Qty Transfer</label>
      <input type="number" class="form-input" id="transfer-qty" min="1" value="1" style="text-align:center;font-size:1.2rem;font-weight:700">
    </div>
    <div class="form-group">
      <label class="form-label">Notes (optional)</label>
      <input type="text" class="form-input" id="transfer-notes" placeholder="e.g. Refill event floor">
    </div>
    <input type="hidden" id="transfer-sku-id" value="${skuId}">`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="doQuickTransfer()">Transfer</button>`);

  setTimeout(() => document.getElementById('transfer-qty')?.focus(), 100);
}

function quickReplenish(skuId, skuCode, skuName) {
  Modal.show(`🔄 Transfer: ${skuCode}`, `
    <p style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:1rem">${skuName}</p>
    <div class="form-group">
      <label class="form-label">Direction</label>
      <select class="form-select" id="transfer-direction">
        <option value="storage_to_event">Storage → Event Floor</option>
        <option value="event_to_storage">Event Floor → Storage</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label">Qty Transfer</label>
      <input type="number" class="form-input" id="transfer-qty" min="1" value="1" style="text-align:center;font-size:1.2rem;font-weight:700" autofocus>
    </div>
    <div class="form-group">
      <label class="form-label">Notes</label>
      <input type="text" class="form-input" id="transfer-notes" placeholder="e.g. Refill event floor">
    </div>
    <input type="hidden" id="transfer-sku-id" value="${skuId}">`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="doQuickTransfer()">Transfer</button>`);

  setTimeout(() => document.getElementById('transfer-qty')?.focus(), 100);
}


async function doQuickTransfer() {
  const skuId = document.getElementById('transfer-sku-id').value;
  const qty = parseInt(document.getElementById('transfer-qty').value) || 0;
  const direction = document.getElementById('transfer-direction').value;
  const notes = document.getElementById('transfer-notes').value;

  if (qty <= 0) { Toast.error('Qty harus > 0'); return; }

  try {
    const res = await API.post('/inventory/transfer', {
      event_id: window.currentEventId,
      sku_id: skuId,
      qty: qty,
      direction: direction,
      notes: notes || 'Quick transfer'
    });
    Modal.hide();
    Toast.success(res.message || `Berhasil transfer ${qty} unit`);
    loadInventory();
  } catch(e) { Toast.error(e.message); }
}

async function loadReplenishAlerts() {
  try {
    const res = await API.get(`/inventory/alerts?event_id=${window.currentEventId}`);
    const alerts = (res.data || []).filter(a => a.needs_replenish);

    if (alerts.length === 0) {
      Modal.show('✅ Replenish Alerts', '<div class="empty-state"><span class="material-symbols-rounded">check_circle</span><h3>Semua stok aman</h3><p>Tidak ada SKU yang perlu replenish</p></div>');
      return;
    }

    const rows = alerts.map(a => `
      <tr>
        <td>${a.sku_code}</td>
        <td>${a.sku_name}</td>
        <td><strong style="color:${a.event_available <= 0 ? 'var(--danger)' : 'var(--warning)'}">${a.event_available}</strong></td>
        <td>${a.storage_onhand}</td>
        <td>${a.storage_depleted ? '<span style="color:var(--danger)">Habis!</span>' :
          `<button class="btn btn-sm btn-primary" onclick="Modal.hide();quickReplenish('${a.sku_id}','${a.sku_code}','${a.sku_name}')">↗ Transfer</button>`}</td>
      </tr>`).join('');

    Modal.show(`⚠️ Replenish Alerts (${alerts.length})`, `
      <div style="max-height:400px;overflow-y:auto">
        <table style="width:100%;font-size:0.85rem"><thead><tr><th>SKU</th><th>Name</th><th>Event Stok</th><th>Storage</th><th>Action</th></tr></thead>
        <tbody>${rows}</tbody></table>
      </div>`,
      `<button class="btn btn-secondary" onclick="Modal.hide()">Tutup</button>`);
  } catch(e) { Toast.error(e.message); }
}

async function loadSalesReport() {
  try {
    const res = await API.get(`/inventory/sales-report?event_id=${window.currentEventId}`);
    const items = res.data || [];
    Modal.show('Sales Report', renderTable([
      { label: 'SKU', key: 'sku_code' },
      { label: 'Name', key: 'sku_name' },
      { label: 'Sold', render: r => `<strong style="color:var(--success)">${r.qty_sold}</strong>` },
      { label: 'Remaining', key: 'qty_onhand' },
    ], items, 'No sales data'));
  } catch(e) { Toast.error(e.message); }
}
