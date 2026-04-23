// Handover / Shipping Page
Router.register('handover', async () => {
  return `
    <div class="scan-container">
      <label style="font-size:0.85rem;color:var(--text-secondary);margin-bottom:0.5rem;display:block">
        Scan Order Label to Ship
      </label>
      <input type="text" class="scan-input" id="handover-scan" placeholder="Scan order label..." autofocus>
      <div id="handover-feedback" style="margin-top:0.75rem;font-size:0.9rem"></div>
    </div>
    <div class="card">
      <div class="card-header"><span class="card-title">Recently Shipped</span></div>
      <div id="shipped-list">Loading...</div>
    </div>`;
});

function init_handover() {
  loadShippedOrders();
  const input = document.getElementById('handover-scan');
  input?.focus();
  input?.addEventListener('keydown', async (e) => {
    if (e.key === 'Enter' && input.value.trim()) {
      const orderNum = input.value.trim();
      input.value = '';
      const fb = document.getElementById('handover-feedback');
      fb.innerHTML = '<span style="color:var(--text-muted)">Processing...</span>';
      try {
        await API.post('/handover/scan', { order_number: orderNum, event_id: window.currentEventId });
        fb.innerHTML = `<span style="color:var(--success)">✓ Order #${orderNum} shipped successfully!</span>`;
        Toast.success(`Order #${orderNum} shipped!`);
        loadShippedOrders();
      } catch(err) {
        fb.innerHTML = `<span style="color:var(--danger)">✗ ${err.message}</span>`;
      }
      input.focus();
    }
  });
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
