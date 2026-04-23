// Picking Page
Router.register('picking', async () => {
  const eventId = window.currentEventId;
  let orders = [];
  try {
    const res = await API.get(`/orders?event_id=${eventId}&status=printed&page_size=100`);
    orders = res.data || [];
  } catch(e) {}

  return `
    <div id="picking-active" class="hidden"></div>
    <div id="picking-list">
      <div class="card" style="margin-bottom:1rem">
        <div class="card-header"><span class="card-title">Orders Ready for Picking</span></div>
        ${renderTable([
          { label: 'Order #', render: r => `<strong style="color:var(--accent-light)">${r.order_number}</strong>` },
          { label: 'Buyer', key: 'buyer_name' },
          { label: 'Status', render: r => statusBadge(r.status) },
          { label: '', render: r => `<button class="btn btn-sm btn-primary" onclick="startPickOrder('${r.id}')">▶ Start</button>` }
        ], orders, 'No orders ready for picking')}
      </div>
    </div>`;
});

async function startPickOrder(orderId) {
  try {
    await API.post(`/picking/${orderId}/start`);
    const res = await API.get(`/orders/${orderId}`);
    const order = res.data;

    document.getElementById('picking-list').classList.add('hidden');
    const active = document.getElementById('picking-active');
    active.classList.remove('hidden');

    const itemRows = (order.items || []).map(it => `
      <tr id="pick-item-${it.id}">
        <td>${it.sku_code}</td>
        <td>${it.variation_name || it.sku_name || '-'}</td>
        <td><strong>${it.qty_ordered}</strong></td>
        <td class="pick-qty" data-id="${it.id}">${it.qty_picked}</td>
        <td>${it.qty_ordered - it.qty_picked > 0 ? `<span style="color:var(--warning)">${it.qty_ordered - it.qty_picked}</span>` : '<span style="color:var(--success)">✓</span>'}</td>
      </tr>`).join('');

    active.innerHTML = `
      <div class="alert alert-info">
        <span class="material-symbols-rounded">info</span>
        Picking order <strong>#${order.order_number}</strong> - ${order.buyer_name || ''}
      </div>
      <div class="scan-container">
        <label style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:0.5rem;display:block">Scan Barcode / SKU Code</label>
        <div style="display:flex;gap:0.5rem;align-items:center">
          <input type="text" class="scan-input" id="pick-scan-input" placeholder="Scan atau ketik SKU..." autofocus style="flex:1">
          <button class="scan-with-camera-btn" onclick="openPickCamera('${orderId}')" title="Scan pakai Kamera">
            <span class="material-symbols-rounded">photo_camera</span> Kamera
          </button>
        </div>
        <div id="pick-scan-feedback" style="margin-top:0.75rem;font-size:0.9rem"></div>
      </div>
      <div class="card">
        <table style="width:100%;font-size:0.85rem">
          <thead><tr><th>SKU</th><th>Item</th><th>Ordered</th><th>Picked</th><th>Remain</th></tr></thead>
          <tbody>${itemRows}</tbody>
        </table>
      </div>
      <div style="display:flex;gap:0.75rem;margin-top:1rem">
        <button class="btn btn-secondary" onclick="cancelPicking()">Cancel</button>
        <button class="btn btn-success" id="complete-pick-btn" onclick="completePicking('${orderId}')">✓ Complete Picking</button>
      </div>`;

    const scanInput = document.getElementById('pick-scan-input');
    scanInput.focus();
    scanInput.addEventListener('keydown', async (e) => {
      if (e.key === 'Enter' && scanInput.value.trim()) {
        const barcode = scanInput.value.trim();
        scanInput.value = '';
        await processPickScan(orderId, barcode);
        scanInput.focus();
      }
    });
  } catch(e) { Toast.error(e.message); }
}

async function processPickScan(orderId, barcode) {
  const fb = document.getElementById('pick-scan-feedback');
  try {
    const res = await API.post(`/picking/${orderId}/scan`, { barcode, qty: 1 });
    const r = res.data;
    if (r.message === 'OK') {
      Sound.success();
      fb.innerHTML = `<span style="color:var(--success)">✓ ${r.sku_code} - ${r.sku_name} (${r.qty_picked}/${r.qty_ordered})</span>`;
      const cell = document.querySelector(`.pick-qty[data-id="${r.item_id}"]`);
      if (cell) cell.textContent = r.qty_picked;
    } else if (r.message && r.message.includes('Cannot exceed')) {
      Sound.warning();
      fb.innerHTML = `<span style="color:var(--warning)">⚠ ${r.message}</span>`;
    } else {
      Sound.error();
      fb.innerHTML = `<span style="color:var(--danger)">⚠ ${r.message}</span>`;
    }
  } catch(err) {
    Sound.error();
    fb.innerHTML = `<span style="color:var(--danger)">✗ ${err.message}</span>`;
  }
}

function openPickCamera(orderId) {
  CameraScanner.open((decodedText) => {
    processPickScan(orderId, decodedText);
    document.getElementById('pick-scan-input')?.focus();
  });
}

async function completePicking(orderId) {
  try {
    await API.post(`/picking/${orderId}/complete`);
    Toast.success('Picking completed!');
    Router.navigate('picking');
  } catch(e) { Toast.error(e.message); }
}

function cancelPicking() {
  document.getElementById('picking-active').classList.add('hidden');
  document.getElementById('picking-list').classList.remove('hidden');
}
