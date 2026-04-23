// Modal component
const Modal = {
  show(title, bodyHtml, footerHtml = '') {
    const overlay = document.getElementById('modal-container');
    document.getElementById('modal-title').textContent = title;
    document.getElementById('modal-body').innerHTML = bodyHtml;
    document.getElementById('modal-footer').innerHTML = footerHtml;
    overlay.classList.remove('hidden');
  },
  hide() {
    document.getElementById('modal-container').classList.add('hidden');
  },
  confirm(title, message, onConfirm) {
    this.show(title, `<p style="color:var(--text-secondary)">${message}</p>`,
      `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
       <button class="btn btn-primary" id="modal-confirm-btn">Confirm</button>`);
    document.getElementById('modal-confirm-btn').onclick = () => { Modal.hide(); onConfirm(); };
  }
};

// Event listeners
document.getElementById('modal-close')?.addEventListener('click', () => Modal.hide());
document.getElementById('modal-container')?.addEventListener('click', (e) => {
  if (e.target.id === 'modal-container') Modal.hide();
});
