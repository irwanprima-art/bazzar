// Users Page
Router.register('users', async () => {
  let users = [];
  try { const r = await API.get('/users'); users = r.data || []; } catch(e) {}
  return `
    <div class="toolbar">
      <button class="btn btn-primary" onclick="showCreateUser()"><span class="material-symbols-rounded">person_add</span> Add User</button>
    </div>
    <div class="card">
      ${renderTable([
        { label: 'Username', render: r => `<strong>${r.username}</strong>` },
        { label: 'Name', key: 'full_name' },
        { label: 'Role', render: r => statusBadge(r.role) },
        { label: 'Active', render: r => r.is_active ? '<span style="color:var(--success)">✓</span>' : '<span style="color:var(--danger)">✗</span>' },
      ], users, 'No users')}
    </div>`;
});

function showCreateUser() {
  Modal.show('Add User', `
    <div class="form-group"><label class="form-label">Username</label><input class="form-input" id="new-username"></div>
    <div class="form-group"><label class="form-label">Password</label><input type="password" class="form-input" id="new-password"></div>
    <div class="form-group"><label class="form-label">Full Name</label><input class="form-input" id="new-fullname"></div>
    <div class="form-group"><label class="form-label">Role</label>
      <select class="form-input" id="new-role"><option value="picker">Picker</option><option value="admin">Admin</option></select>
    </div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="createUser()">Create</button>`);
}

async function createUser() {
  try {
    await API.post('/users', {
      username: document.getElementById('new-username').value,
      password: document.getElementById('new-password').value,
      full_name: document.getElementById('new-fullname').value,
      role: document.getElementById('new-role').value
    });
    Toast.success('User created!');
    Modal.hide();
    Router.navigate('users');
  } catch(e) { Toast.error(e.message); }
}
