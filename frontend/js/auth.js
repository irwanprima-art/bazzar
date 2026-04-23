// Auth Module
const Auth = {
  user: null,

  async login(username, password) {
    const res = await API.post('/auth/login', { username, password });
    API.setToken(res.data.token);
    this.user = res.data.user;
    return this.user;
  },

  async getMe() {
    const res = await API.get('/auth/me');
    this.user = res.data;
    return this.user;
  },

  logout() {
    API.clearToken();
    this.user = null;
    document.getElementById('login-screen').classList.add('active');
    document.getElementById('login-screen').classList.remove('hidden');
    document.getElementById('main-shell').classList.add('hidden');
  },

  isAdmin() { return this.user?.role === 'admin'; },
  isLoggedIn() { return !!API.getToken(); },
};
