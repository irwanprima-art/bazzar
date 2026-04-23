const CACHE_NAME = 'bazzar-v1';
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/css/style.css',
  '/js/api.js',
  '/js/auth.js',
  '/js/router.js',
  '/js/app.js',
  '/js/components/toast.js',
  '/js/components/modal.js',
  '/js/components/table.js',
  '/js/components/label.js',
  '/manifest.json'
];

self.addEventListener('install', (e) => {
  e.waitUntil(caches.open(CACHE_NAME).then(c => c.addAll(STATIC_ASSETS)));
  self.skipWaiting();
});

self.addEventListener('activate', (e) => {
  e.waitUntil(caches.keys().then(keys =>
    Promise.all(keys.filter(k => k !== CACHE_NAME).map(k => caches.delete(k)))
  ));
  self.clients.claim();
});

self.addEventListener('fetch', (e) => {
  if (e.request.url.includes('/api/')) return; // Don't cache API calls
  e.respondWith(
    caches.match(e.request).then(r => r || fetch(e.request))
  );
});
