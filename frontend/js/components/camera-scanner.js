// Camera Barcode Scanner using html5-qrcode
const CameraScanner = {
  _scanner: null,
  _active: false,

  /**
   * Open camera scanner modal
   * @param {function} onResult - callback(decodedText) when barcode detected
   */
  open(onResult) {
    if (this._active) return;
    this._active = true;

    // Create scanner modal
    const overlay = document.createElement('div');
    overlay.id = 'camera-scanner-overlay';
    overlay.innerHTML = `
      <div class="camera-scanner-modal">
        <div class="camera-scanner-header">
          <span class="material-symbols-rounded" style="font-size:1.2rem">photo_camera</span>
          <span>Scan Barcode / QR</span>
          <button class="btn btn-sm btn-danger" id="camera-scanner-close">✕</button>
        </div>
        <div id="camera-scanner-reader"></div>
        <div class="camera-scanner-footer">
          <div id="camera-scanner-result" style="font-size:0.85rem;color:var(--text-muted)">Arahkan kamera ke barcode...</div>
        </div>
      </div>
    `;
    document.body.appendChild(overlay);

    // Close handler
    document.getElementById('camera-scanner-close').onclick = () => this.close();
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) this.close();
    });

    // Init scanner
    this._scanner = new Html5Qrcode('camera-scanner-reader');
    
    const config = {
      fps: 10,
      qrbox: { width: 280, height: 150 },
      aspectRatio: 1.0,
      formatsToSupport: [
        Html5QrcodeSupportedFormats.QR_CODE,
        Html5QrcodeSupportedFormats.EAN_13,
        Html5QrcodeSupportedFormats.EAN_8,
        Html5QrcodeSupportedFormats.CODE_128,
        Html5QrcodeSupportedFormats.CODE_39,
        Html5QrcodeSupportedFormats.UPC_A,
        Html5QrcodeSupportedFormats.UPC_E,
        Html5QrcodeSupportedFormats.ITF,
      ]
    };

    this._scanner.start(
      { facingMode: 'environment' },
      config,
      (decodedText) => {
        Sound.success();
        document.getElementById('camera-scanner-result').innerHTML = 
          `<span style="color:var(--success)">✓ Terdeteksi: <strong>${decodedText}</strong></span>`;
        
        // Small delay so user sees the result, then callback
        setTimeout(() => {
          this.close();
          if (onResult) onResult(decodedText);
        }, 400);
      },
      () => {} // ignore scan failures (continuous scanning)
    ).catch(err => {
      document.getElementById('camera-scanner-result').innerHTML = 
        `<span style="color:var(--danger)">Gagal akses kamera: ${err}</span>`;
    });
  },

  close() {
    if (this._scanner) {
      this._scanner.stop().catch(() => {});
      this._scanner.clear().catch(() => {});
      this._scanner = null;
    }
    const overlay = document.getElementById('camera-scanner-overlay');
    if (overlay) overlay.remove();
    this._active = false;
  }
};
