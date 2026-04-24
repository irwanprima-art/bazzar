// SPA Router
const Router = {
  currentPage: 'dashboard',
  pages: {},

  register(name, renderFn) { this.pages[name] = renderFn; },

  async navigate(page) {
    this.currentPage = page;
    const content = document.getElementById('main-content');
    const title = document.getElementById('page-title');

    // Update nav active states
    document.querySelectorAll('.nav-link').forEach(l => l.classList.toggle('active', l.dataset.page === page));
    document.querySelectorAll('.bottom-nav-item').forEach(l => l.classList.toggle('active', l.dataset.page === page));

    // Close mobile nav
    document.getElementById('side-nav')?.classList.remove('open');

    // Titles
    const titles = {
      dashboard: 'Dashboard', orders: 'Orders', picking: 'Picking',
      handover: 'Handover', inbound: 'Inbound', inventory: 'Inventory',
      skus: 'SKU Master', users: 'Users', events: 'Events', stocklog: 'Stock Log'
    };
    title.textContent = titles[page] || page;

    // Render
    if (this.pages[page]) {
      content.innerHTML = '<div class="animate-in" style="animation:pulse 1s infinite;text-align:center;padding:3rem;color:var(--text-muted)">Loading...</div>';
      try {
        const html = await this.pages[page]();
        content.innerHTML = `<div class="animate-in">${html}</div>`;
        // Run page init if exists
        if (window[`init_${page}`]) window[`init_${page}`]();
      } catch (e) {
        content.innerHTML = `<div class="empty-state"><span class="material-symbols-rounded">error</span><h3>Error</h3><p>${e.message}</p></div>`;
      }
    } else {
      content.innerHTML = `<div class="empty-state"><span class="material-symbols-rounded">construction</span><h3>Coming Soon</h3><p>This page is under construction</p></div>`;
    }
  }
};
