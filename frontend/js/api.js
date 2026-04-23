// API Client for Bazzar Makuku
const API = {
  baseURL: '/api/v1',

  getToken() { return localStorage.getItem('bazzar_token'); },
  setToken(t) { localStorage.setItem('bazzar_token', t); },
  clearToken() { localStorage.removeItem('bazzar_token'); },

  async request(method, path, body, isFormData = false) {
    const headers = {};
    const token = this.getToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (!isFormData) headers['Content-Type'] = 'application/json';

    const opts = { method, headers };
    if (body) opts.body = isFormData ? body : JSON.stringify(body);

    const res = await fetch(this.baseURL + path, opts);
    if (res.status === 401 && !path.includes('/auth/login')) {
      this.clearToken();
      window.Auth?.logout();
      throw new Error('Session expired');
    }
    const data = await res.json();
    if (!data.success) throw new Error(data.message || 'Request failed');
    return data;
  },

  get(path) { return this.request('GET', path); },
  post(path, body) { return this.request('POST', path, body); },
  put(path, body) { return this.request('PUT', path, body); },
  del(path) { return this.request('DELETE', path); },
  upload(path, formData) { return this.request('POST', path, formData, true); },
};
