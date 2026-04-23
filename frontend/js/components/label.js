// Label component for 10x10cm printing with QR
const Label = {
  generateQRDataURL(text, size = 150) {
    // Simple QR-like visual using Canvas (for actual QR, we'll use a library)
    // This generates a data URL with a text-encoded pattern
    const canvas = document.createElement('canvas');
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext('2d');

    // Background
    ctx.fillStyle = '#fff';
    ctx.fillRect(0, 0, size, size);

    // Generate deterministic pattern from text
    ctx.fillStyle = '#000';
    const cellSize = Math.floor(size / 21);
    const hash = text.split('').reduce((a, c) => ((a << 5) - a + c.charCodeAt(0)) | 0, 0);

    // Draw finder patterns (corners)
    const drawFinder = (x, y) => {
      ctx.fillRect(x, y, cellSize * 7, cellSize * 7);
      ctx.fillStyle = '#fff';
      ctx.fillRect(x + cellSize, y + cellSize, cellSize * 5, cellSize * 5);
      ctx.fillStyle = '#000';
      ctx.fillRect(x + cellSize * 2, y + cellSize * 2, cellSize * 3, cellSize * 3);
    };
    drawFinder(0, 0);
    ctx.fillStyle = '#000';
    drawFinder(cellSize * 14, 0);
    ctx.fillStyle = '#000';
    drawFinder(0, cellSize * 14);

    // Fill data area with hash-based pattern
    ctx.fillStyle = '#000';
    for (let r = 0; r < 21; r++) {
      for (let c = 0; c < 21; c++) {
        if ((r < 7 && c < 7) || (r < 7 && c > 13) || (r > 13 && c < 7)) continue;
        const val = Math.abs((hash * (r + 1) * (c + 1)) % 3);
        if (val === 0) ctx.fillRect(c * cellSize, r * cellSize, cellSize, cellSize);
      }
    }

    // Add text below
    ctx.fillStyle = '#000';
    ctx.font = `bold ${Math.floor(size / 15)}px Inter, sans-serif`;
    ctx.textAlign = 'center';
    ctx.fillText(text, size / 2, size - 4);

    return canvas.toDataURL();
  },

  render(order) {
    const qrDataURL = this.generateQRDataURL(order.order_number);
    const items = (order.items || []).map(it =>
      `<div class="label-item"><span>${it.sku_code} - ${it.variation_name || it.sku_name || ''}</span><span>x${it.qty_ordered}</span></div>`
    ).join('');

    return `
      <div class="print-label" id="print-label-${order.id}">
        <div class="label-header">
          <div>
            <div class="label-order">#${order.order_number}</div>
            <div class="label-buyer">${order.buyer_name || order.buyer_username || '-'}</div>
          </div>
          <img class="label-qr" src="${qrDataURL}" alt="QR">
        </div>
        <div class="label-items">${items || '<div style="color:#999">No items</div>'}</div>
        <div class="label-footer">Bazzar Makuku • ${new Date().toLocaleDateString('id-ID')}</div>
      </div>
    `;
  },

  print(order) {
    const win = window.open('', '_blank', 'width=400,height=400');
    win.document.write(`
      <!DOCTYPE html><html><head>
      <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
      <style>
        * { margin:0; padding:0; box-sizing:border-box; }
        body { font-family: Inter, sans-serif; }
        .print-label { width:100mm; height:100mm; padding:5mm; }
        .label-header { display:flex; justify-content:space-between; border-bottom:2px solid #000; padding-bottom:3mm; margin-bottom:3mm; }
        .label-order { font-size:14pt; font-weight:800; }
        .label-qr { width:25mm; height:25mm; }
        .label-buyer { font-size:10pt; font-weight:600; margin-bottom:2mm; }
        .label-items { font-size:9pt; }
        .label-item { display:flex; justify-content:space-between; padding:1mm 0; border-bottom:1px dotted #999; }
        .label-footer { margin-top:auto; font-size:7pt; color:#666; text-align:center; padding-top:3mm; }
        @media print { @page { size:100mm 100mm; margin:0; } }
      </style></head><body>
      ${this.render(order)}
      <script>setTimeout(()=>{window.print();window.close()},500)<\/script>
      </body></html>`);
  }
};
