// Handover / Shipping Page
Router.register('handover', async () => {
  return `
    <div class="scan-container">
      <label style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:0.5rem;display:block">
        Scan Order Label to Ship
      </label>
      <div style="display:flex;gap:0.5rem;align-items:center">
        <input type="text" class="scan-input" id="handover-scan" placeholder="Scan order label..." autofocus style="flex:1">
        <button class="scan-with-camera-btn" onclick="openHandoverCamera()" title="Scan pakai Kamera">
          <span class="material-symbols-rounded">photo_camera</span> Kamera
        </button>
      </div>
      <div id="handover-feedback" style="margin-top:0.75rem;font-size:0.9rem"></div>
    </div>
    <div class="card" style="margin-bottom:1rem">
      <div class="card-header"><span class="card-title">📦 Ready to Ship (Picked)</span></div>
      <div id="picked-list">Loading...</div>
    </div>
    <div class="card">
      <div class="card-header"><span class="card-title">✅ Recently Shipped</span></div>
      <div id="shipped-list">Loading...</div>
    </div>`;
});

function init_handover() {
  loadPickedOrders();
  loadShippedOrders();
  const input = document.getElementById('handover-scan');
  input?.focus();
  input?.addEventListener('keydown', async (e) => {
    if (e.key === 'Enter' && input.value.trim()) {
      const orderNum = input.value.trim();
      input.value = '';
      await processHandoverScan(orderNum);
      input.focus();
    }
  });
}

function openHandoverCamera() {
  if (typeof CameraScanner !== 'undefined') {
    CameraScanner.open((decodedText) => {
      processHandoverScan(decodedText);
      document.getElementById('handover-scan')?.focus();
    });
  }
}

async function processHandoverScan(orderNum) {
  const fb = document.getElementById('handover-feedback');
  fb.innerHTML = '<span style="color:var(--text-muted)">Processing...</span>';
  try {
    await API.post('/handover/scan', { order_number: orderNum, event_id: window.currentEventId });
    Sound.success();
    fb.innerHTML = `<span style="color:var(--success)">✓ Order #${orderNum} shipped successfully!</span>`;
    Toast.success(`Order #${orderNum} shipped!`);
    loadPickedOrders();
    loadShippedOrders();
  } catch(err) {
    Sound.error();
    fb.innerHTML = `<span style="color:var(--danger)">✗ ${err.message}</span>`;
  }
}

async function shipOrderManual(orderId, orderNum) {
  if (!confirm(`Ship order #${orderNum}?`)) return;
  try {
    await API.post('/handover/scan', { order_number: orderNum, event_id: window.currentEventId });
    Toast.success(`Order #${orderNum} shipped!`);
    loadPickedOrders();
    loadShippedOrders();
  } catch(e) { Toast.error(e.message); }
}

async function loadPickedOrders() {
  try {
    const res = await API.get(`/orders?event_id=${window.currentEventId}&status=picked&page_size=100`);
    const orders = res.data || [];
    const el = document.getElementById('picked-list');
    if (orders.length === 0) {
      el.innerHTML = '<div style="text-align:center;padding:1.5rem;color:var(--text-muted)">Tidak ada order siap kirim</div>';
      return;
    }
    el.innerHTML = renderTable([
      { label: 'Order #', render: r => `<strong style="color:var(--accent-light)">${r.order_number}</strong>` },
      { label: 'Buyer', key: 'buyer_name' },
      { label: 'Items', render: r => {
          const items = r.items || [];
          if (items.length === 0) return '-';
          return items.map(it => `<div style="font-size:0.75rem">${it.sku_code} × ${it.qty_ordered}</div>`).join('');
        }
      },
      { label: 'Status', render: r => statusBadge(r.status) },
      { label: '', render: r => `<button class="btn btn-sm btn-success" onclick="shipOrderManual('${r.id}','${r.order_number}')">🚚 Ship</button>` }
    ], orders, '');
  } catch(e) { document.getElementById('picked-list').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

async function loadShippedOrders() {
  try {
    const res = await API.get(`/orders?event_id=${window.currentEventId}&status=shipped&page_size=20`);
    document.getElementById('shipped-list').innerHTML = renderTable([
      { label: 'Order #', key: 'order_number' },
      { label: 'Buyer', key: 'buyer_name' },
      { label: 'Shipped At', render: r => r.shipped_at ? new Date(r.shipped_at).toLocaleString('id-ID') : '-' },
      { label: 'Status', render: r => statusBadge(r.status) }
    ], res.data || [], 'No shipped orders yet');
  } catch(e) { document.getElementById('shipped-list').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}
