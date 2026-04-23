// SKU Master Page
Router.register('skus', async () => {
  return `
    <div class="toolbar">
      <div class="toolbar-search">
        <span class="material-symbols-rounded">search</span>
        <input class="form-input" id="sku-search" placeholder="Search SKU code, barcode, name...">
      </div>
      <button class="btn btn-primary" onclick="showCreateSKU()"><span class="material-symbols-rounded">add</span> Add SKU</button>
    </div>
    <div id="sku-table" class="card">Loading...</div>`;
});

function init_skus() {
  loadSKUs();
  document.getElementById('sku-search')?.addEventListener('input', debounce(loadSKUs, 400));
}

async function loadSKUs() {
  const search = document.getElementById('sku-search')?.value || '';
  try {
    const res = await API.get(`/skus?search=${search}&page_size=50`);
    document.getElementById('sku-table').innerHTML = renderTable([
      { label: 'SKU Code', render: r => `<strong>${r.sku_code}</strong>` },
      { label: 'Barcode', render: r => r.barcode || '<span style="color:var(--text-muted)">-</span>' },
      { label: 'Name', key: 'name' },
      { label: 'Replenish Limit', key: 'replenish_limit' },
      { label: '', render: r => `<button class="btn btn-sm btn-secondary" onclick="editSKU('${r.id}')">✏️</button>` }
    ], res.data || [], 'No SKUs found');
  } catch(e) { document.getElementById('sku-table').innerHTML = `<div class="alert alert-danger">${e.message}</div>`; }
}

function showCreateSKU() {
  Modal.show('Add SKU', `
    <div class="form-group"><label class="form-label">SKU Code *</label><input class="form-input" id="sku-code"></div>
    <div class="form-group"><label class="form-label">Barcode</label><input class="form-input" id="sku-barcode"></div>
    <div class="form-group"><label class="form-label">Name *</label><input class="form-input" id="sku-name"></div>
    <div class="form-group"><label class="form-label">Replenish Limit</label><input type="number" class="form-input" id="sku-limit" value="5"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="createSKU()">Create</button>`);
}

async function createSKU() {
  try {
    await API.post('/skus', {
      sku_code: document.getElementById('sku-code').value,
      barcode: document.getElementById('sku-barcode').value,
      name: document.getElementById('sku-name').value,
      replenish_limit: parseInt(document.getElementById('sku-limit').value) || 5
    });
    Toast.success('SKU created!');
    Modal.hide();
    loadSKUs();
  } catch(e) { Toast.error(e.message); }
}

async function editSKU(id) {
  try {
    const res = await API.get(`/skus/${id}`);
    const s = res.data;
    Modal.show('Edit SKU', `
      <div class="form-group"><label class="form-label">SKU Code</label><input class="form-input" id="edit-sku-code" value="${s.sku_code}"></div>
      <div class="form-group"><label class="form-label">Barcode</label><input class="form-input" id="edit-sku-barcode" value="${s.barcode || ''}"></div>
      <div class="form-group"><label class="form-label">Name</label><input class="form-input" id="edit-sku-name" value="${s.name}"></div>
      <div class="form-group"><label class="form-label">Replenish Limit</label><input type="number" class="form-input" id="edit-sku-limit" value="${s.replenish_limit}"></div>`,
      `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
       <button class="btn btn-primary" onclick="updateSKU('${id}')">Save</button>`);
  } catch(e) { Toast.error(e.message); }
}

async function updateSKU(id) {
  try {
    await API.put(`/skus/${id}`, {
      sku_code: document.getElementById('edit-sku-code').value,
      barcode: document.getElementById('edit-sku-barcode').value,
      name: document.getElementById('edit-sku-name').value,
      replenish_limit: parseInt(document.getElementById('edit-sku-limit').value) || 5
    });
    Toast.success('SKU updated!');
    Modal.hide();
    loadSKUs();
  } catch(e) { Toast.error(e.message); }
}
