// Inventory Page
Router.register('inventory', async () => {
  return `
    <div class="toolbar">
      <select class="filter-select" id="inv-location-filter">
        <option value="">All Locations</option>
        <option value="EVENT">Event Floor</option>
        <option value="STORAGE">Storage</option>
      </select>
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
    ], res.data || [], 'No inventory data');
  } catch(e) { document.getElementById('inventory-table').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
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
