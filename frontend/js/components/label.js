// Label component for 10x10cm printing with QR
const Label = {
  generateQRDataURL(text, size = 150) {
    // Use qrcode-generator library if loaded, otherwise use API fallback
    if (typeof qrcode !== 'undefined') {
      const qr = qrcode(0, 'M');
      qr.addData(text);
      qr.make();

      const canvas = document.createElement('canvas');
      const cellSize = Math.max(Math.floor(size / (qr.getModuleCount() + 2)), 2);
      const realSize = cellSize * (qr.getModuleCount() + 2);
      canvas.width = realSize;
      canvas.height = realSize;
      const ctx = canvas.getContext('2d');

      ctx.fillStyle = '#fff';
      ctx.fillRect(0, 0, realSize, realSize);

      ctx.fillStyle = '#000';
      const offset = cellSize;
      for (let r = 0; r < qr.getModuleCount(); r++) {
        for (let c = 0; c < qr.getModuleCount(); c++) {
          if (qr.isDark(r, c)) {
            ctx.fillRect(offset + c * cellSize, offset + r * cellSize, cellSize, cellSize);
          }
        }
      }
      return canvas.toDataURL();
    }

    // Fallback: use Google Charts QR API
    return `https://chart.googleapis.com/chart?cht=qr&chs=${size}x${size}&chl=${encodeURIComponent(text)}&choe=UTF-8`;
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
      <script src="https://cdnjs.cloudflare.com/ajax/libs/qrcode-generator/1.4.4/qrcode.min.js"><\/script>
      <style>
        * { margin:0; padding:0; box-sizing:border-box; }
        body { font-family: Inter, sans-serif; }
        .print-label { width:100mm; height:100mm; padding:5mm; display:flex; flex-direction:column; }
        .label-header { display:flex; justify-content:space-between; border-bottom:2px solid #000; padding-bottom:3mm; margin-bottom:3mm; }
        .label-order { font-size:14pt; font-weight:800; }
        .label-qr { width:25mm; height:25mm; image-rendering: pixelated; }
        .label-buyer { font-size:10pt; font-weight:600; margin-bottom:2mm; }
        .label-items { font-size:9pt; flex:1; }
        .label-item { display:flex; justify-content:space-between; padding:1mm 0; border-bottom:1px dotted #999; }
        .label-footer { font-size:7pt; color:#666; text-align:center; padding-top:3mm; }
        @media print { @page { size:100mm 100mm; margin:0; } }
      </style></head><body>
      <div class="print-label">
        <div class="label-header">
          <div>
            <div class="label-order">#${order.order_number}</div>
            <div class="label-buyer">${order.buyer_name || order.buyer_username || '-'}</div>
          </div>
          <canvas id="qr-canvas" style="width:25mm;height:25mm;image-rendering:pixelated"></canvas>
        </div>
        <div class="label-items">${(order.items || []).map(it =>
          `<div class="label-item"><span>${it.sku_code} - ${it.variation_name || it.sku_name || ''}</span><span>x${it.qty_ordered}</span></div>`
        ).join('') || '<div style="color:#999">No items</div>'}</div>
        <div class="label-footer">Bazzar Makuku • ${new Date().toLocaleDateString('id-ID')}</div>
      </div>
      <script>
        // Wait for qrcode library to load, then render QR
        function renderQR() {
          if (typeof qrcode === 'undefined') { setTimeout(renderQR, 100); return; }
          var qr = qrcode(0, 'M');
          qr.addData('${order.order_number}');
          qr.make();
          var canvas = document.getElementById('qr-canvas');
          var size = 200;
          var cellSize = Math.floor(size / (qr.getModuleCount() + 2));
          var realSize = cellSize * (qr.getModuleCount() + 2);
          canvas.width = realSize;
          canvas.height = realSize;
          var ctx = canvas.getContext('2d');
          ctx.fillStyle = '#fff';
          ctx.fillRect(0, 0, realSize, realSize);
          ctx.fillStyle = '#000';
          var offset = cellSize;
          for (var r = 0; r < qr.getModuleCount(); r++) {
            for (var c = 0; c < qr.getModuleCount(); c++) {
              if (qr.isDark(r, c)) {
                ctx.fillRect(offset + c * cellSize, offset + r * cellSize, cellSize, cellSize);
              }
            }
          }
          setTimeout(function(){ window.print(); window.close(); }, 300);
        }
        renderQR();
      <\/script>
      </body></html>`);
  }
};
