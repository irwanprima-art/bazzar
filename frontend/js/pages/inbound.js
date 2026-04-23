// Inbound Page
Router.register('inbound', async () => {
  let orders = [];
  try {
    const res = await API.get(`/inbound?event_id=${window.currentEventId}`);
    orders = res.data || [];
  } catch(e) {}

  return `
    <div class="toolbar">
      ${Auth.isAdmin() ? `<button class="btn btn-primary" onclick="showInboundImport()"><span class="material-symbols-rounded">upload_file</span> Import PO</button>` : ''}
    </div>
    <div id="inbound-active" class="hidden"></div>
    <div id="inbound-list" class="card">
      <div class="card-header"><span class="card-title">Inbound Purchase Orders</span></div>
      ${renderTable([
        { label: 'Reference', render: r => `<strong style="color:var(--accent-light)">${r.reference_number}</strong>` },
        { label: 'Status', render: r => statusBadge(r.status) },
        { label: 'Created', render: r => new Date(r.created_at).toLocaleDateString('id-ID') },
        { label: 'By', key: 'imported_by_name' },
        { label: '', render: r => `<button class="btn btn-sm btn-primary" onclick="openInbound('${r.id}')">📦 Receive</button>
          <button class="btn btn-sm btn-secondary" onclick="viewInbound('${r.id}')">👁️</button>` }
      ], orders, 'No inbound orders')}
    </div>`;
});

function showInboundImport() {
  Modal.show('Import Inbound PO', `
    <div class="form-group">
      <label class="form-label">Reference Number</label>
      <input type="text" class="form-input" id="inbound-ref" placeholder="e.g. PO-001">
    </div>
    <div class="drop-zone" id="inbound-drop-zone">
      <span class="material-symbols-rounded">cloud_upload</span>
      <p>Upload PO Excel (columns: SKU, Qty)</p>
      <input type="file" id="inbound-file" accept=".xlsx,.xls" style="display:none">
    </div>
    <div id="inbound-import-result" style="margin-top:1rem"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>`);

  const dz = document.getElementById('inbound-drop-zone');
  const fi = document.getElementById('inbound-file');
  dz.onclick = () => fi.click();
  dz.ondragover = (e) => { e.preventDefault(); dz.classList.add('dragover'); };
  dz.ondragleave = () => dz.classList.remove('dragover');
  dz.ondrop = (e) => { e.preventDefault(); dz.classList.remove('dragover'); handleInboundFile(e.dataTransfer.files[0]); };
  fi.onchange = () => { if (fi.files[0]) handleInboundFile(fi.files[0]); };
}

async function handleInboundFile(file) {
  const ref = document.getElementById('inbound-ref').value || 'PO-' + Date.now().toString(36);
  const rd = document.getElementById('inbound-import-result');
  rd.innerHTML = '<div style="text-align:center;color:var(--text-muted)">Importing...</div>';
  try {
    const fd = new FormData();
    fd.append('file', file);
    fd.append('event_id', window.currentEventId);
    fd.append('reference_number', ref);
    const res = await API.upload('/inbound/import', fd);
    rd.innerHTML = `<div class="alert alert-success">PO created with ${res.data.items?.length || 0} items</div>`;
    Toast.success('Inbound PO imported!');
    setTimeout(() => { Modal.hide(); Router.navigate('inbound'); }, 1500);
  } catch(e) { rd.innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

// Track current inbound session
let _currentInboundId = null;
let _currentInboundItems = [];

async function openInbound(inboundId) {
  try {
    const res = await API.get(`/inbound/${inboundId}`);
    const po = res.data;
    _currentInboundId = inboundId;
    _currentInboundItems = po.items || [];

    document.getElementById('inbound-list').classList.add('hidden');
    const active = document.getElementById('inbound-active');
    active.classList.remove('hidden');

    active.innerHTML = `
      <div class="alert alert-info"><span class="material-symbols-rounded">info</span> Receiving PO: <strong>${po.reference_number}</strong> - ${statusBadge(po.status)}</div>
      <div class="scan-container">
        <label style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:0.5rem;display:block">📦 Scan Barcode to Receive</label>
        <input type="text" class="scan-input" id="inbound-scan" placeholder="Scan barcode..." autofocus>
        <div id="inbound-scan-fb" style="margin-top:0.75rem;font-size:0.9rem"></div>
      </div>
      <div class="card" style="margin-top:1rem">
        <table style="width:100%;font-size:0.85rem"><thead><tr><th>SKU</th><th>Name</th><th>Expected</th><th>Received</th><th>Remain</th></tr></thead>
        <tbody id="inbound-items-tbody">${renderInboundRows(po.items)}</tbody></table>
      </div>
      <button class="btn btn-secondary" style="margin-top:1rem" onclick="closeInbound()">← Back</button>`;

    const scanInput = document.getElementById('inbound-scan');
    scanInput.focus();
    scanInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && scanInput.value.trim()) {
        handleInboundScan(scanInput.value.trim());
        scanInput.value = '';
      }
    });
  } catch(e) { Toast.error(e.message); }
}

function renderInboundRows(items) {
  return (items || []).map(it => `
    <tr id="inb-row-${it.id}">
      <td>${it.sku_code}</td><td>${it.sku_name || '-'}</td>
      <td>${it.qty_expected}</td>
      <td class="inb-recv" data-id="${it.id}">${it.qty_received}</td>
      <td>${it.qty_remaining > 0 ? `<span style="color:var(--warning)">${it.qty_remaining}</span>` : '<span style="color:var(--success)">✓</span>'}</td>
    </tr>
  `).join('');
}

async function handleInboundScan(barcode) {
  const fb = document.getElementById('inbound-scan-fb');

  // Find matching item in the PO by barcode or SKU code
  const matchedItem = _currentInboundItems.find(it =>
    it.barcode === barcode || it.sku_code === barcode
  );

  if (!matchedItem) {
    fb.innerHTML = `<span style="color:var(--danger)">✗ Barcode "${barcode}" not found in this PO</span>`;
    return;
  }

  const remaining = matchedItem.qty_expected - matchedItem.qty_received;
  const skuLabel = `${matchedItem.sku_code} - ${matchedItem.sku_name || ''}`;

  // Show qty input modal
  Modal.show('📦 Receive Item', `
    <div style="margin-bottom:1rem">
      <div style="font-size:0.8rem;color:var(--text-muted);margin-bottom:0.25rem">SKU</div>
      <div style="font-size:1.1rem;font-weight:700;color:var(--accent-light)">${skuLabel}</div>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:0.75rem;margin-bottom:1.25rem">
      <div class="stat-card info" style="padding:0.75rem;text-align:center">
        <div class="stat-label">Expected</div>
        <div class="stat-value" style="font-size:1.3rem">${matchedItem.qty_expected}</div>
      </div>
      <div class="stat-card success" style="padding:0.75rem;text-align:center">
        <div class="stat-label">Received</div>
        <div class="stat-value" style="font-size:1.3rem">${matchedItem.qty_received}</div>
      </div>
      <div class="stat-card warning" style="padding:0.75rem;text-align:center">
        <div class="stat-label">Remaining</div>
        <div class="stat-value" style="font-size:1.3rem">${remaining}</div>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">Qty to Receive</label>
      <input type="number" class="form-input" id="inbound-qty-input" value="1" min="1" style="text-align:center;font-size:1.2rem;font-weight:700" autofocus>
    </div>
    <div id="inbound-qty-warn" class="hidden" style="margin-top:0.5rem"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide();document.getElementById('inbound-scan').focus()">Cancel</button>
     <button class="btn btn-primary" id="inbound-confirm-btn" onclick="confirmInboundReceive('${barcode}', ${remaining})">Confirm Receive</button>`);

  // Focus qty input after modal opens
  setTimeout(() => {
    const qtyInput = document.getElementById('inbound-qty-input');
    if (qtyInput) {
      qtyInput.focus();
      qtyInput.select();
      qtyInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          document.getElementById('inbound-confirm-btn').click();
        }
      });
    }
  }, 100);
}

async function confirmInboundReceive(barcode, remaining) {
  const qtyInput = document.getElementById('inbound-qty-input');
  const qty = parseInt(qtyInput.value) || 1;

  // If qty exceeds remaining, show warning confirmation
  if (qty > remaining && remaining > 0) {
    const warnDiv = document.getElementById('inbound-qty-warn');
    if (!warnDiv.dataset.confirmed) {
      warnDiv.classList.remove('hidden');
      warnDiv.innerHTML = `
        <div class="alert alert-warning" style="margin-bottom:0">
          <span class="material-symbols-rounded">warning</span>
          <div>
            <strong>Qty melebihi sisa (${remaining})!</strong><br>
            <span style="font-size:0.8rem">Klik "Confirm Receive" lagi untuk konfirmasi</span>
          </div>
        </div>`;
      warnDiv.dataset.confirmed = 'pending';

      // Change button to danger style
      const btn = document.getElementById('inbound-confirm-btn');
      btn.className = 'btn btn-danger';
      btn.innerHTML = '⚠️ Yes, Receive ' + qty;
      return;
    }
  }

  // Submit to backend
  const btn = document.getElementById('inbound-confirm-btn');
  btn.disabled = true;
  btn.textContent = 'Saving...';

  try {
    const r = await API.post(`/inbound/${_currentInboundId}/scan?event_id=${window.currentEventId}`, { barcode, qty });
    const d = r.data;
    Modal.hide();

    // Update table row
    const cell = document.querySelector(`.inb-recv[data-id="${d.item_id}"]`);
    if (cell) cell.textContent = d.qty_received;

    // Update local item data
    const item = _currentInboundItems.find(it => it.id === d.item_id);
    if (item) {
      item.qty_received = d.qty_received;
      item.qty_remaining = d.qty_expected - d.qty_received;
    }

    // Refresh remaining column
    document.getElementById('inbound-items-tbody').innerHTML = renderInboundRows(_currentInboundItems);

    // Show success feedback
    const fb = document.getElementById('inbound-scan-fb');
    fb.innerHTML = `<span style="color:var(--success)">✓ ${d.sku_code} received +${qty} (${d.qty_received}/${d.qty_expected})</span>`;
    Toast.success(`${d.sku_code} received: +${qty}`);

    // Refocus scan input
    document.getElementById('inbound-scan').focus();
  } catch(err) {
    Modal.hide();
    const fb = document.getElementById('inbound-scan-fb');
    fb.innerHTML = `<span style="color:var(--danger)">✗ ${err.message}</span>`;
    Toast.error(err.message);
    document.getElementById('inbound-scan').focus();
  }
}

async function viewInbound(id) {
  try {
    const res = await API.get(`/inbound/${id}`);
    const po = res.data;
    const items = (po.items || []).map(it => `<tr><td>${it.sku_code}</td><td>${it.sku_name}</td><td>${it.qty_expected}</td><td>${it.qty_received}</td></tr>`).join('');
    Modal.show(`PO: ${po.reference_number}`, `
      <div style="margin-bottom:1rem">${statusBadge(po.status)}</div>
      <table style="width:100%;font-size:0.85rem"><thead><tr><th>SKU</th><th>Name</th><th>Expected</th><th>Received</th></tr></thead><tbody>${items}</tbody></table>`);
  } catch(e) { Toast.error(e.message); }
}

function closeInbound() {
  _currentInboundId = null;
  _currentInboundItems = [];
  document.getElementById('inbound-active').classList.add('hidden');
  document.getElementById('inbound-list').classList.remove('hidden');
}
