// Stock Movement Log Page
Router.register('stocklog', async () => {
  return `
    <div class="card" style="margin-bottom:1rem">
      <div class="card-header" style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:0.5rem">
        <span class="card-title"><span class="material-symbols-rounded" style="vertical-align:middle">history</span> Stock Movement Log</span>
        <div style="display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap">
          <input type="text" class="form-input" id="log-search" placeholder="Cari SKU code / nama..." 
            style="width:220px;font-size:0.8rem;padding:0.4rem 0.6rem">
          <select class="form-select" id="log-action-filter" style="width:auto;font-size:0.8rem;padding:0.4rem 0.6rem">
            <option value="">Semua Action</option>
            <option value="inbound">Inbound</option>
            <option value="allocate">Allocate</option>
            <option value="deallocate">Deallocate</option>
            <option value="pick">Pick</option>
            <option value="ship">Ship</option>
            <option value="replenish_out">Replenish Out</option>
            <option value="replenish_in">Replenish In</option>
            <option value="transfer_out">Transfer Out</option>
            <option value="transfer_in">Transfer In</option>
            <option value="adjust">Adjust</option>
            <option value="return">Return</option>
          </select>
          <button class="btn btn-sm btn-primary" onclick="loadStockLogs()">
            <span class="material-symbols-rounded" style="font-size:0.9rem">search</span> Filter
          </button>
        </div>
      </div>
      <div id="stock-log-summary" style="padding:0.75rem 1rem;border-bottom:1px solid var(--border-color)"></div>
      <div id="stock-log-table" style="padding:0">Loading...</div>
    </div>`;
});

function init_stocklog() {
  loadStockLogs();
  document.getElementById('log-search')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') loadStockLogs();
  });
}

async function loadStockLogs() {
  const search = document.getElementById('log-search')?.value?.trim() || '';
  const actionFilter = document.getElementById('log-action-filter')?.value || '';
  const eventId = window.currentEventId;

  const tableDiv = document.getElementById('stock-log-table');
  const summaryDiv = document.getElementById('stock-log-summary');
  tableDiv.innerHTML = '<div style="text-align:center;padding:2rem;color:var(--text-muted)">Loading...</div>';

  try {
    let url = `/inventory/logs?event_id=${eventId}&limit=100`;
    if (search) url += `&search=${encodeURIComponent(search)}`;
    const res = await API.get(url);
    let logs = res.data || [];

    // Client-side filter by action type
    if (actionFilter) {
      logs = logs.filter(l => l.action === actionFilter);
    }

    // Calculate summary per SKU
    renderLogSummary(summaryDiv, logs, search);

    // Render table
    renderLogTable(tableDiv, logs);
  } catch(e) {
    tableDiv.innerHTML = `<div class="alert alert-danger" style="margin:1rem">${e.message}</div>`;
  }
}

function renderLogSummary(el, logs, search) {
  if (!search || logs.length === 0) {
    el.innerHTML = `<span style="font-size:0.8rem;color:var(--text-muted)">${logs.length} movement records</span>`;
    return;
  }

  // Aggregate by SKU
  const skuMap = {};
  logs.forEach(l => {
    if (!skuMap[l.sku_code]) {
      skuMap[l.sku_code] = { name: l.sku_name, inbound: 0, sold: 0, transferred: 0, adjusted: 0 };
    }
    const s = skuMap[l.sku_code];
    if (l.action === 'inbound') s.inbound += l.qty_change;
    if (l.action === 'ship') s.sold += Math.abs(l.qty_change);
    if (['transfer_in', 'transfer_out', 'replenish_in', 'replenish_out'].includes(l.action)) s.transferred += l.qty_change;
    if (l.action === 'adjust') s.adjusted += l.qty_change;
  });

  const skuEntries = Object.entries(skuMap);
  const cards = skuEntries.map(([code, s]) => `
    <div style="background:var(--bg-input);border-radius:0.5rem;padding:0.6rem 0.8rem;min-width:180px">
      <div style="font-size:0.8rem;font-weight:700;color:var(--accent-light);margin-bottom:0.3rem">${code}</div>
      <div style="font-size:0.7rem;color:var(--text-muted);margin-bottom:0.3rem">${s.name}</div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.2rem;font-size:0.75rem">
        <span>📦 Inbound:</span><span style="color:var(--success);font-weight:600">+${s.inbound}</span>
        <span>🚚 Sold:</span><span style="color:var(--danger);font-weight:600">-${s.sold}</span>
        ${s.transferred !== 0 ? `<span>🔄 Transfer:</span><span style="font-weight:600">${s.transferred > 0 ? '+' : ''}${s.transferred}</span>` : ''}
        ${s.adjusted !== 0 ? `<span>⚙️ Adjust:</span><span style="font-weight:600">${s.adjusted > 0 ? '+' : ''}${s.adjusted}</span>` : ''}
      </div>
    </div>
  `).join('');

  el.innerHTML = `
    <div style="font-size:0.8rem;color:var(--text-muted);margin-bottom:0.5rem">${logs.length} records found</div>
    <div style="display:flex;gap:0.5rem;overflow-x:auto;padding-bottom:0.25rem">${cards}</div>`;
}

function renderLogTable(el, logs) {
  if (logs.length === 0) {
    el.innerHTML = '<div style="text-align:center;padding:2rem;color:var(--text-muted)">No movement records found</div>';
    return;
  }

  const actionBadge = (action) => {
    const map = {
      'inbound':       { color: '#00cec9', icon: '📦', label: 'Inbound' },
      'allocate':      { color: '#6c5ce7', icon: '📌', label: 'Allocate' },
      'deallocate':    { color: '#a29bfe', icon: '📌', label: 'Deallocate' },
      'pick':          { color: '#fdcb6e', icon: '✋', label: 'Pick' },
      'ship':          { color: '#e17055', icon: '🚚', label: 'Ship' },
      'replenish_out': { color: '#74b9ff', icon: '⬆', label: 'Repl Out' },
      'replenish_in':  { color: '#55efc4', icon: '⬇', label: 'Repl In' },
      'transfer_out':  { color: '#fab1a0', icon: '↗', label: 'Tfr Out' },
      'transfer_in':   { color: '#81ecec', icon: '↙', label: 'Tfr In' },
      'adjust':        { color: '#dfe6e9', icon: '⚙', label: 'Adjust' },
      'return':        { color: '#ffeaa7', icon: '↩', label: 'Return' },
    };
    const m = map[action] || { color: '#636e72', icon: '•', label: action };
    return `<span style="display:inline-flex;align-items:center;gap:0.2rem;padding:0.15rem 0.5rem;border-radius:0.3rem;font-size:0.7rem;font-weight:600;background:${m.color}22;color:${m.color};white-space:nowrap">${m.icon} ${m.label}</span>`;
  };

  const qtyColor = (q) => q > 0 ? 'var(--success)' : q < 0 ? 'var(--danger)' : 'var(--text-muted)';
  const qtySign = (q) => q > 0 ? `+${q}` : `${q}`;
  const fmtTime = (t) => {
    const d = new Date(t);
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' }) + ' ' +
           d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
  };

  const rows = logs.map(l => `
    <tr>
      <td style="white-space:nowrap;font-size:0.75rem;color:var(--text-muted)">${fmtTime(l.created_at)}</td>
      <td><strong style="color:var(--accent-light);font-size:0.8rem">${l.sku_code}</strong>
        <div style="font-size:0.7rem;color:var(--text-muted)">${l.sku_name || ''}</div></td>
      <td>${actionBadge(l.action)}</td>
      <td style="text-align:center;font-weight:700;color:${qtyColor(l.qty_change)};font-size:0.9rem">${qtySign(l.qty_change)}</td>
      <td style="font-size:0.75rem">${l.location_code || '-'}</td>
      <td style="font-size:0.75rem;color:var(--text-muted)">${l.reference_number || l.reference_type || '-'}</td>
      <td style="font-size:0.75rem;color:var(--text-muted)">${l.username || '-'}</td>
      <td style="font-size:0.7rem;color:var(--text-muted);max-width:120px;overflow:hidden;text-overflow:ellipsis">${l.notes || ''}</td>
    </tr>
  `).join('');

  el.innerHTML = `
    <div style="overflow-x:auto">
      <table style="width:100%;font-size:0.85rem">
        <thead>
          <tr>
            <th style="white-space:nowrap">Waktu</th>
            <th>SKU</th>
            <th>Action</th>
            <th style="text-align:center">Qty</th>
            <th>Lokasi</th>
            <th>Referensi</th>
            <th>User</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>`;
}
