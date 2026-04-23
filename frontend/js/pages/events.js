// Events Page
Router.register('events', async () => {
  let events = [];
  try { const r = await API.get('/events'); events = r.data || []; } catch(e) {}
  return `
    <div class="toolbar">
      <button class="btn btn-primary" onclick="showCreateEvent()"><span class="material-symbols-rounded">add</span> New Event</button>
    </div>
    <div class="card">
      ${renderTable([
        { label: 'Name', render: r => `<strong>${r.name}</strong>` },
        { label: 'Start', key: 'start_date' },
        { label: 'End', key: 'end_date' },
        { label: 'Active', render: r => r.is_active ? '<span class="badge badge-completed">Active</span>' : '<span class="badge badge-cancelled">Inactive</span>' },
      ], events, 'No events')}
    </div>`;
});

function showCreateEvent() {
  Modal.show('Create Event', `
    <div class="form-group"><label class="form-label">Event Name</label><input class="form-input" id="evt-name"></div>
    <div class="form-group"><label class="form-label">Description</label><input class="form-input" id="evt-desc"></div>
    <div class="form-group"><label class="form-label">Start Date</label><input type="date" class="form-input" id="evt-start"></div>
    <div class="form-group"><label class="form-label">End Date</label><input type="date" class="form-input" id="evt-end"></div>`,
    `<button class="btn btn-secondary" onclick="Modal.hide()">Cancel</button>
     <button class="btn btn-primary" onclick="createEvent()">Create</button>`);
}

async function createEvent() {
  try {
    await API.post('/events', {
      name: document.getElementById('evt-name').value,
      description: document.getElementById('evt-desc').value,
      start_date: document.getElementById('evt-start').value,
      end_date: document.getElementById('evt-end').value
    });
    Toast.success('Event created!');
    Modal.hide();
    Router.navigate('events');
  } catch(e) { Toast.error(e.message); }
}
