// Orders Page
Router.register('orders', async () => {
  return `
    <div class="toolbar">
      <div class="toolbar-search">
        <span class="material-symbols-rounded">search</span>
        <input class="form-input" id="order-search" placeholder="Search order # or buyer...">
      </div>
      <select class="filter-select" id="order-status-filter">
        <option value="">All Status</option>
        <option value="imported">Imported</option>
        <option value="allocated">Allocated</option>
        <option value="printed">Printed</option>
        <option value="picking">Picking</option>
        <option value="picked">Picked</option>
        <option value="shipped">Shipped</option>
        <option value="issue">Issue</option>
      </select>
      ${Auth.isAdmin() ? `
        <button class="btn btn-primary" id="import-orders-btn"><span class="material-symbols-rounded">upload_file</span> Import</button>
        <button class="btn btn-secondary" id="allocate-all-btn"><span class="material-symbols-rounded">assignment_turned_in</span> Allocate All</button>
      ` : ''}
    </div>
    <div id="orders-table">Loading...</div>
    <div id="orders-pagination" style="display:flex;justify-content:center;gap:0.5rem;margin-top:1rem"></div>`;
});

let ordersPage = 1;
function init_orders() {
  loadOrders();
  document.getElementById('order-search')?.addEventListener('input', debounce(() => { ordersPage = 1; loadOrders(); }, 400));
  document.getElementById('order-status-filter')?.addEventListener('change', () => { ordersPage = 1; loadOrders(); });
  document.getElementById('import-orders-btn')?.addEventListener('click', showImportModal);
  document.getElementById('allocate-all-btn')?.addEventListener('click', allocateAllOrders);
}

async function loadOrders() {
  const search = document.getElementById('order-search')?.value || '';
  const status = document.getElementById('order-status-filter')?.value || '';
  const eventId = window.currentEventId;
  try {
    const res = await API.get(`/orders?event_id=${eventId}&page=${ordersPage}&page_size=20&status=${status}&search=${search}`);
    const orders = res.data || [];
    const meta = res.meta || {};
    document.getElementById('orders-table').innerHTML = renderTable([
      { label: 'Order #', render: r => `<strong style="color:var(--accent-light);cursor:pointer" onclick="viewOrder('${r.id}')">${r.order_number}</strong>` },
      { label: 'Platform', render: r => `<span style="font-size:0.75rem;color:var(--text-muted)">${r.platform_status}</span>` },
      { label: 'Status', render: r => statusBadge(r.status) },
      { label: 'Buyer', key: 'buyer_name' },
      { label: 'SKU', render: r => r.product_name ? r.product_name.substring(0, 40) + '...' : '-' },
      { label: 'Actions', render: r => orderActions(r) }
    ], orders, 'No orders found');

    // Pagination
    const pg = document.getElementById('orders-pagination');
    if (meta.total_page > 1) {
      let btns = '';
      for (let i = 1; i <= meta.total_page; i++) {
        btns += `<button class="btn btn-sm ${i === ordersPage ? 'btn-primary' : 'btn-secondary'}" onclick="ordersPage=${i};loadOrders()">${i}</button>`;
      }
      pg.innerHTML = btns;
    } else pg.innerHTML = '';
  } catch(e) { document.getElementById('orders-table').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

function orderActions(order) {
  let btns = '';
  if (order.status === 'allocated' && Auth.isAdmin()) {
    btns += `<button class="btn btn-sm btn-warning" onclick="printLabel('${order.id}')">🏷️ Print</button> `;
  }
  if (order.status === 'printed') {
    btns += `<button class="btn btn-sm btn-primary" onclick="startPickingFromOrder('${order.id}')">▶ Pick</button> `;
  }
  btns += `<button class="btn btn-sm btn-secondary" onclick="viewOrder('${order.id}')">👁️</button>`;
  return btns;
}

async function viewOrder(orderId) {
  try {
    const res = await API.get(`/orders/${orderId}`);
    const o = res.data;
    const items = (o.items || []).map(it => `
      <tr><td>${it.sku_code}</td><td>${it.variation_name || it.sku_name || '-'}</td><td>${it.qty_ordered}</td><td>${it.qty_picked}</td></tr>
    `).join('');
    Modal.show(`Order #${o.order_number}`, `
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.75rem;margin-bottom:1rem;font-size:0.85rem">
        <div><span style="color:var(--text-muted)">Status:</span> ${statusBadge(o.status)}</div>
        <div><span style="color:var(--text-muted)">Platform:</span> ${o.platform_status}</div>
        <div><span style="color:var(--text-muted)">Buyer:</span> ${o.buyer_name || '-'}</div>
        <div><span style="color:var(--text-muted)">Shipping:</span> ${o.shipping_option}</div>
      </div>
      <h4 style="margin-bottom:0.5rem">Items</h4>
      <table style="width:100%;font-size:0.85rem"><thead><tr><th>SKU</th><th>Variant</th><th>Ordered</th><th>Picked</th></tr></thead><tbody>${items}</tbody></table>
      ${o.status === 'allocated' ? `<div style="margin-top:1rem">${Label.render(o)}</div>` : ''}
    `, o.status === 'allocated' ? `<button class="btn btn-warning" onclick="printLabel('${o.id}')">🏷️ Print Label & Mark Printed</button>` : '');
  } catch(e) { Toast.error(e.message); }
}

async function printLabel(orderId) {
  try {
    const res = await API.get(`/orders/${orderId}`);
    const order = res.data;
    Label.print(order);
    if (Auth.isAdmin()) {
      await API.post(`/orders/${orderId}/print`);
      Toast.success('Label printed & order marked as printed');
      Modal.hide();
      loadOrders();
    }
  } catch(e) { Toast.error(e.message); }
}

function showImportModal() {
  Modal.show('Import Shopee Orders', `
    <p style="color:var(--text-secondary);margin-bottom:1rem;font-size:0.85rem">
      Upload Shopee order export (.xlsx). Only <strong>Jasa Kirim Toko</strong> orders will be imported.
      Duplicates are automatically skipped.
    </p>
    <div class="drop-zone" id="import-drop-zone">
      <span class="material-symbols-rounded">cloud_upload</span>
      <p>Drag & drop Excel file here</p>
      <p style="font-size:0.75rem;margin-top:0.5rem">or click to browse</p>
      <input type="file" id="import-file" accept=".xlsx,.xls" style="display:none">
    </div>
    <div id="import-result" style="margin-top:1rem"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Close</button>`);

  const dropZone = document.getElementById('import-drop-zone');
  const fileInput = document.getElementById('import-file');

  dropZone.onclick = () => fileInput.click();
  dropZone.ondragover = (e) => { e.preventDefault(); dropZone.classList.add('dragover'); };
  dropZone.ondragleave = () => dropZone.classList.remove('dragover');
  dropZone.ondrop = (e) => { e.preventDefault(); dropZone.classList.remove('dragover'); handleImportFile(e.dataTransfer.files[0]); };
  fileInput.onchange = () => { if (fileInput.files[0]) handleImportFile(fileInput.files[0]); };
}

async function handleImportFile(file) {
  const resultDiv = document.getElementById('import-result');
  resultDiv.innerHTML = '<div style="text-align:center;color:var(--text-muted)"><div class="btn-loader" style="margin:0 auto"></div><p>Importing...</p></div>';
  try {
    const fd = new FormData();
    fd.append('file', file);
    fd.append('event_id', window.currentEventId);
    const res = await API.upload('/orders/import', fd);
    const r = res.data;
    resultDiv.innerHTML = `
      <div class="alert alert-success">Import complete!</div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;font-size:0.85rem">
        <div>Total Rows: <strong>${r.total_rows}</strong></div>
        <div>Imported: <strong style="color:var(--success)">${r.imported}</strong></div>
        <div>Updated: <strong style="color:var(--accent-light)">${r.updated || 0}</strong></div>
        <div>Duplicates: <strong style="color:var(--warning)">${r.duplicates}</strong></div>
        <div>Skipped: <strong>${r.skipped}</strong></div>
        <div>Errors: <strong style="color:var(--danger)">${r.errors}</strong></div>
      </div>
      ${r.skipped_details?.length ? `<details style="margin-top:0.75rem;font-size:0.8rem"><summary style="cursor:pointer;color:var(--text-muted)">Skipped details</summary><pre style="max-height:200px;overflow:auto;font-size:0.75rem;color:var(--text-muted);margin-top:0.5rem">${r.skipped_details.join('\n')}</pre></details>` : ''}`;
    const updText = r.updated ? `, ${r.updated} updated` : '';
    Toast.success(`Imported ${r.imported}${updText} orders`);
    loadOrders();
  } catch(e) { resultDiv.innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

async function allocateAllOrders() {
  try {
    const res = await API.post(`/orders/allocate-all?event_id=${window.currentEventId}`);
    const r = res.data;
    const failedCount = r.failed?.length || 0;

    if (failedCount === 0) {
      Toast.success(`✓ Berhasil allocate ${r.allocated} dari ${r.total_orders} orders`);
    } else {
      // Show detailed modal with failures
      const failRows = r.failed.map(f => `
        <tr>
          <td><strong style="color:var(--accent-light)">${f.order_number}</strong></td>
          <td>${f.sku_code}</td>
          <td style="color:var(--danger);font-size:0.8rem">${f.reason}</td>
        </tr>`).join('');

      Modal.show('⚠️ Hasil Alokasi', `
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.75rem;margin-bottom:1rem">
          <div class="stat-card success" style="padding:0.75rem;text-align:center">
            <div class="stat-label">Allocated</div>
            <div class="stat-value" style="font-size:1.5rem">${r.allocated}</div>
          </div>
          <div class="stat-card danger" style="padding:0.75rem;text-align:center">
            <div class="stat-label">Gagal</div>
            <div class="stat-value" style="font-size:1.5rem">${failedCount}</div>
          </div>
        </div>
        <div style="max-height:300px;overflow-y:auto">
          <table style="width:100%;font-size:0.8rem"><thead><tr><th>Order</th><th>SKU</th><th>Alasan</th></tr></thead>
          <tbody>${failRows}</tbody></table>
        </div>
        <div class="alert alert-warning" style="margin-top:1rem">
          <span class="material-symbols-rounded">info</span>
          <span>Order yang gagal perlu <strong>Replenish</strong> stok dulu dari halaman Inventory</span>
        </div>`,
        `<button class="btn btn-secondary" onclick="Modal.hide()">Tutup</button>`);

      Toast.warning(`${r.allocated} allocated, ${failedCount} gagal (stok kurang)`);
    }
    loadOrders();
  } catch(e) { Toast.error(e.message); }
}

function startPickingFromOrder(orderId) { Modal.hide(); Router.navigate('picking'); setTimeout(() => startPickOrder(orderId), 500); }
function debounce(fn, ms) { let t; return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), ms); }; }
