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
        <div id="camera-scanner-reader" style="width:100%;min-height:300px"></div>
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

    // Start camera after DOM is ready
    setTimeout(() => this._startCamera(onResult), 200);
  },

  _startCamera(onResult) {
    try {
      // formatsToSupport goes in constructor config
      const config = { 
        formatsToSupport: [
          Html5QrcodeSupportedFormats.QR_CODE,
          Html5QrcodeSupportedFormats.EAN_13,
          Html5QrcodeSupportedFormats.EAN_8,
          Html5QrcodeSupportedFormats.CODE_128,
          Html5QrcodeSupportedFormats.CODE_39,
          Html5QrcodeSupportedFormats.UPC_A,
          Html5QrcodeSupportedFormats.UPC_E,
          Html5QrcodeSupportedFormats.ITF,
        ],
        verbose: false
      };

      this._scanner = new Html5Qrcode('camera-scanner-reader', config);

      const scanConfig = {
        fps: 10,
        qrbox: { width: 250, height: 120 },
        aspectRatio: 1.333,
      };

      this._scanner.start(
        { facingMode: 'environment' },
        scanConfig,
        (decodedText) => {
          // Barcode detected
          Sound.success();
          const resultEl = document.getElementById('camera-scanner-result');
          if (resultEl) {
            resultEl.innerHTML = `<span style="color:var(--success)">✓ <strong>${decodedText}</strong></span>`;
          }
          // Brief delay so user sees result
          setTimeout(() => {
            this.close();
            if (onResult) onResult(decodedText);
          }, 500);
        },
        (errorMessage) => {
          // Ignore continuous scan failures - this fires every frame without a barcode
        }
      ).catch(err => {
        console.error('Camera start error:', err);
        const resultEl = document.getElementById('camera-scanner-result');
        if (resultEl) {
          resultEl.innerHTML = `<span style="color:var(--danger)">Gagal akses kamera. Pastikan izin kamera diberikan.<br><small>${err}</small></span>`;
        }
      });
    } catch(err) {
      console.error('Scanner init error:', err);
      const resultEl = document.getElementById('camera-scanner-result');
      if (resultEl) {
        resultEl.innerHTML = `<span style="color:var(--danger)">Error: ${err.message || err}</span>`;
      }
    }
  },

  close() {
    if (this._scanner) {
      try {
        const state = this._scanner.getState();
        // Only stop if scanning (state 2 = SCANNING)
        if (state === 2) {
          this._scanner.stop().then(() => {
            this._scanner.clear();
          }).catch(() => {});
        } else {
          this._scanner.clear();
        }
      } catch(e) {
        // Ignore cleanup errors
      }
      this._scanner = null;
    }
    const overlay = document.getElementById('camera-scanner-overlay');
    if (overlay) overlay.remove();
    this._active = false;
  }
};
