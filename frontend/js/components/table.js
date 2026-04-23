// Reusable table component
function renderTable(columns, rows, emptyMsg = 'No data') {
  if (!rows || rows.length === 0) {
    return `<div class="empty-state"><span class="material-symbols-rounded">inbox</span><h3>${emptyMsg}</h3></div>`;
  }
  const ths = columns.map(c => `<th>${c.label}</th>`).join('');
  const trs = rows.map(row => {
    const tds = columns.map(c => `<td>${c.render ? c.render(row) : (row[c.key] || '-')}</td>`).join('');
    return `<tr>${tds}</tr>`;
  }).join('');
  return `<div class="table-container"><table><thead><tr>${ths}</tr></thead><tbody>${trs}</tbody></table></div>`;
}

function statusBadge(status) {
  return `<span class="badge badge-${status}">${status}</span>`;
}
