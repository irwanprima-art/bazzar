// App Initialization
window.currentEventId = null;

document.addEventListener('DOMContentLoaded', async () => {
  // Login form
  document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.getElementById('login-btn');
    const errDiv = document.getElementById('login-error');
    btn.querySelector('.btn-text').textContent = 'Signing in...';
    btn.disabled = true;
    errDiv.classList.add('hidden');

    try {
      const user = await Auth.login(
        document.getElementById('login-username').value,
        document.getElementById('login-password').value
      );
      showMainApp(user);
    } catch(e) {
      errDiv.textContent = e.message;
      errDiv.classList.remove('hidden');
    } finally {
      btn.querySelector('.btn-text').textContent = 'Sign In';
      btn.disabled = false;
    }
  });

  // Check existing session (silently clear if stale)
  if (Auth.isLoggedIn()) {
    try {
      const user = await Auth.getMe();
      showMainApp(user);
    } catch(e) {
      // Silently clear stale token, don't show error
      API.clearToken();
      Auth.user = null;
    }
  }

  // Nav clicks
  document.querySelectorAll('[data-page]').forEach(el => {
    el.addEventListener('click', (e) => {
      e.preventDefault();
      Router.navigate(el.dataset.page);
    });
  });

  // Menu toggle
  document.getElementById('menu-toggle')?.addEventListener('click', () => {
    document.getElementById('side-nav').classList.toggle('open');
  });
  document.getElementById('nav-overlay')?.addEventListener('click', () => {
    document.getElementById('side-nav').classList.remove('open');
  });

  // User dropdown
  document.getElementById('user-avatar')?.addEventListener('click', () => {
    document.getElementById('user-dropdown').classList.toggle('hidden');
  });
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.user-info')) {
      document.getElementById('user-dropdown')?.classList.add('hidden');
    }
  });

  // Logout
  document.getElementById('logout-btn')?.addEventListener('click', () => {
    Auth.logout();
  });
});

async function showMainApp(user) {
  document.getElementById('login-screen').classList.remove('active');
  document.getElementById('login-screen').classList.add('hidden');
  document.getElementById('main-shell').classList.remove('hidden');

  // Set user info
  document.getElementById('user-avatar').textContent = (user.full_name || user.username)[0].toUpperCase();
  document.getElementById('user-display').textContent = `${user.full_name} (${user.role})`;

  // Show/hide admin-only items
  document.querySelectorAll('.admin-only').forEach(el => {
    el.style.display = user.role === 'admin' ? '' : 'none';
  });

  // Load active event
  try {
    const res = await API.get('/events/active');
    window.currentEventId = res.data.id;
    document.getElementById('current-event-name').textContent = res.data.name;
  } catch(e) {
    document.getElementById('current-event-name').textContent = 'No Event';
  }

  // Navigate to dashboard
  Router.navigate('dashboard');
}
