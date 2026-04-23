// Dashboard Page
Router.register('dashboard', async () => {
  const eventId = window.currentEventId;
  if (!eventId) return '<div class="empty-state"><span class="material-symbols-rounded">event</span><h3>No Active Event</h3></div>';

  let counts = {}, alerts = [];
  try {
    const [cRes, aRes] = await Promise.all([
      API.get(`/orders/status-counts?event_id=${eventId}`),
      API.get(`/inventory/alerts?event_id=${eventId}`)
    ]);
    counts = cRes.data || {};
    alerts = aRes.data || [];
  } catch(e) { console.error(e); }

  const total = Object.values(counts).reduce((a,b) => a+b, 0);
  const alertsHtml = alerts.length > 0 ? `
    <div class="card" style="margin-bottom:1rem">
      <div class="card-header">
        <span class="card-title">⚠️ Replenishment Alerts</span>
        <span class="badge badge-warning">${alerts.length}</span>
      </div>
      ${renderTable([
        { label: 'SKU', key: 'sku_code' },
        { label: 'Name', key: 'sku_name' },
        { label: 'Event Available', render: r => `<strong style="color:${r.event_available <= 0 ? 'var(--danger)' : 'var(--warning)'}">${r.event_available}</strong>` },
        { label: 'Storage', render: r => r.storage_depleted ? '<span style="color:var(--danger)">EMPTY</span>' : r.storage_onhand },
        { label: 'Action', render: r => r.storage_depleted
          ? '<span class="badge badge-issue">No Stock</span>'
          : `<button class="btn btn-sm btn-warning" onclick="replenishFromDashboard('${r.sku_id}','${r.sku_name}',${r.storage_onhand})">Replenish</button>` }
      ], alerts)}
    </div>` : '';

  return `
    <div class="stats-grid">
      <div class="stat-card accent"><div class="stat-label">Total Orders</div><div class="stat-value">${total}</div></div>
      <div class="stat-card info"><div class="stat-label">Imported</div><div class="stat-value">${counts.imported || 0}</div></div>
      <div class="stat-card" style="--c:var(--accent-light)"><div class="stat-label">Allocated</div><div class="stat-value" style="color:var(--accent-light)">${counts.allocated || 0}</div></div>
      <div class="stat-card warning"><div class="stat-label">Printed</div><div class="stat-value">${counts.printed || 0}</div></div>
      <div class="stat-card" style="border-color:orange"><div class="stat-label">Picking</div><div class="stat-value" style="color:orange">${counts.picking || 0}</div></div>
      <div class="stat-card success"><div class="stat-label">Picked</div><div class="stat-value">${counts.picked || 0}</div></div>
      <div class="stat-card" style="border-color:var(--success)"><div class="stat-label">Shipped</div><div class="stat-value" style="color:#00e6e0">${counts.shipped || 0}</div></div>
      <div class="stat-card danger"><div class="stat-label">Issues</div><div class="stat-value">${counts.issue || 0}</div></div>
    </div>
    ${alertsHtml}
    <div class="card">
      <div class="card-header">
        <span class="card-title">Quick Actions</span>
      </div>
      <div style="display:flex;gap:0.75rem;flex-wrap:wrap">
        <button class="btn btn-primary" onclick="Router.navigate('orders')"><span class="material-symbols-rounded">shopping_cart</span> View Orders</button>
        <button class="btn btn-secondary" onclick="Router.navigate('picking')"><span class="material-symbols-rounded">inventory_2</span> Start Picking</button>
        <button class="btn btn-secondary" onclick="Router.navigate('handover')"><span class="material-symbols-rounded">local_shipping</span> Handover</button>
        <button class="btn btn-secondary" onclick="Router.navigate('inbound')"><span class="material-symbols-rounded">move_to_inbox</span> Inbound</button>
      </div>
    </div>`;
});

async function replenishFromDashboard(skuId, skuName, maxQty) {
  Modal.show('Replenish Stock', `
    <p style="color:var(--text-secondary);margin-bottom:1rem">Transfer <strong>${skuName}</strong> from Storage to Event</p>
    <div class="form-group">
      <label class="form-label">Quantity (max: ${maxQty})</label>
      <input type="number" class="form-input" id="replenish-qty" value="${Math.min(maxQty, 10)}" min="1" max="${maxQty}">
    </div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-warning" id="do-replenish-btn">Replenish</button>`);

  document.getElementById('do-replenish-btn').onclick = async () => {
    const qty = parseInt(document.getElementById('replenish-qty').value);
    try {
      await API.post('/inventory/replenish', { event_id: window.currentEventId, sku_id: skuId, qty });
      Toast.success('Stock replenished!');
      Modal.hide();
      Router.navigate('dashboard');
    } catch(e) { Toast.error(e.message); }
  };
}
