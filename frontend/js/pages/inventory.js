// Inventory Page
Router.register('inventory', async () => {
  return `
    <div class="toolbar">
      <select class="filter-select" id="inv-location-filter">
        <option value="">All Locations</option>
        <option value="EVENT">Event Floor</option>
        <option value="STORAGE">Storage</option>
      </select>
      ${Auth.isAdmin() ? `
        <button class="btn btn-primary" onclick="showReplenishModal()"><span class="material-symbols-rounded">sync_alt</span> Replenish</button>
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
      { label: '', render: r => Auth.isAdmin() ? `<button class="btn btn-sm btn-primary" onclick="quickReplenish('${r.sku_id}','${r.sku_code}','${r.sku_name || ''}')">↗ Replenish</button>` : '' },
    ], res.data || [], 'No inventory data');
  } catch(e) { document.getElementById('inventory-table').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

async function showReplenishModal() {
  // Load SKU list from inventory (storage side)
  let skus = [];
  try {
    const res = await API.get(`/inventory?event_id=${window.currentEventId}&location=STORAGE`);
    skus = res.data || [];
  } catch(e) {}

  const skuOptions = skus.map(s => `<option value="${s.sku_id}" data-code="${s.sku_code}" data-stock="${s.qty_onhand}">${s.sku_code} - ${s.sku_name || ''} (stok: ${s.qty_onhand})</option>`).join('');

  Modal.show('🔄 Replenish Stock (Storage → Event)', `
    <p style="font-size:0.8rem;color:var(--text-muted);margin-bottom:1rem">Transfer stok dari Storage ke Event Floor</p>
    <div class="form-group">
      <label class="form-label">SKU</label>
      <select class="form-select" id="replenish-sku">
        <option value="">Pilih SKU...</option>
        ${skuOptions}
      </select>
    </div>
    <div class="form-group">
      <label class="form-label">Qty Transfer</label>
      <input type="number" class="form-input" id="replenish-qty" min="1" value="1" style="text-align:center">
    </div>
    <div class="form-group">
      <label class="form-label">Notes (optional)</label>
      <input type="text" class="form-input" id="replenish-notes" placeholder="e.g. Refill event floor">
    </div>
    <div id="replenish-info" style="margin-top:0.5rem"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="doReplenish()">Transfer</button>`);

  // Show storage info when SKU selected
  document.getElementById('replenish-sku')?.addEventListener('change', function() {
    const opt = this.selectedOptions[0];
    const stock = opt?.dataset?.stock || 0;
    const info = document.getElementById('replenish-info');
    if (this.value) {
      info.innerHTML = `<div class="alert alert-info"><span class="material-symbols-rounded">info</span> Stok di Storage: <strong>${stock}</strong></div>`;
    } else {
      info.innerHTML = '';
    }
  });
}

function quickReplenish(skuId, skuCode, skuName) {
  Modal.show(`↗ Replenish: ${skuCode}`, `
    <p style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:1rem">${skuName}</p>
    <div class="form-group">
      <label class="form-label">Qty Transfer (Storage → Event)</label>
      <input type="number" class="form-input" id="replenish-qty" min="1" value="1" style="text-align:center;font-size:1.2rem;font-weight:700" autofocus>
    </div>
    <div class="form-group">
      <label class="form-label">Notes</label>
      <input type="text" class="form-input" id="replenish-notes" placeholder="e.g. Refill event floor">
    </div>
    <input type="hidden" id="replenish-sku-id" value="${skuId}">`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="doQuickReplenish()">Transfer</button>`);

  setTimeout(() => document.getElementById('replenish-qty')?.focus(), 100);
}

async function doReplenish() {
  const skuId = document.getElementById('replenish-sku').value;
  const qty = parseInt(document.getElementById('replenish-qty').value) || 0;
  const notes = document.getElementById('replenish-notes').value;

  if (!skuId) { Toast.error('Pilih SKU dulu'); return; }
  if (qty <= 0) { Toast.error('Qty harus > 0'); return; }

  try {
    await API.post('/inventory/replenish', {
      event_id: window.currentEventId,
      sku_id: skuId,
      qty: qty,
      notes: notes || 'Manual replenish'
    });
    Modal.hide();
    Toast.success(`Berhasil transfer ${qty} unit ke Event Floor`);
    loadInventory();
  } catch(e) { Toast.error(e.message); }
}

async function doQuickReplenish() {
  const skuId = document.getElementById('replenish-sku-id').value;
  const qty = parseInt(document.getElementById('replenish-qty').value) || 0;
  const notes = document.getElementById('replenish-notes').value;

  if (qty <= 0) { Toast.error('Qty harus > 0'); return; }

  try {
    await API.post('/inventory/replenish', {
      event_id: window.currentEventId,
      sku_id: skuId,
      qty: qty,
      notes: notes || 'Quick replenish'
    });
    Modal.hide();
    Toast.success(`Berhasil transfer ${qty} unit ke Event Floor`);
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
