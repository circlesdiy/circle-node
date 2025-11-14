// Service Worker for circles.diy PWA
// Basic caching strategy for static assets and core pages

const CACHE_NAME = 'circles-diy-v2';
const STATIC_CACHE_URLS = [
  '/',
  '/dashboard',
  '/circles',
  '/chat', 
  '/gather',
  '/profile',
  '/static/css/style.css',
  '/static/js/htmx.min.js',
  '/static/img/icon-192.png',
  '/static/img/icon-512.png',
  '/static/img/favicon-light.svg',
  '/static/img/favicon-dark.svg'
];

// Install event - cache static resources
self.addEventListener('install', event => {
  console.log('Service Worker installing...');
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => {
        console.log('Caching static resources');
        return cache.addAll(STATIC_CACHE_URLS);
      })
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
  console.log('Service Worker activating...');
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames
          .filter(cacheName => cacheName !== CACHE_NAME)
          .map(cacheName => caches.delete(cacheName))
      );
    }).then(() => self.clients.claim())
  );
});

// Helper function to determine if request is for HTML page
function isHTMLRequest(request) {
  const acceptHeader = request.headers.get('Accept');
  return request.mode === 'navigate' ||
         (acceptHeader && acceptHeader.includes('text/html'));
}

// Fetch event - network-first for HTML, cache-first for static assets
self.addEventListener('fetch', event => {
  // Only handle GET requests
  if (event.request.method !== 'GET') return;

  // Skip cross-origin requests
  if (!event.request.url.startsWith(self.location.origin)) return;

  // Network-first strategy for HTML pages (always fetch fresh content)
  if (isHTMLRequest(event.request)) {
    event.respondWith(
      fetch(event.request)
        .then(response => {
          // Cache the fresh HTML for offline fallback
          if (response.status === 200) {
            const responseClone = response.clone();
            caches.open(CACHE_NAME)
              .then(cache => {
                cache.put(event.request, responseClone);
              });
          }
          return response;
        })
        .catch(() => {
          // Fallback to cache when offline
          return caches.match(event.request)
            .then(cachedResponse => {
              return cachedResponse || caches.match('/dashboard');
            });
        })
    );
  } else {
    // Cache-first strategy for static assets (CSS, JS, images)
    event.respondWith(
      caches.match(event.request)
        .then(response => {
          if (response) {
            return response; // Return cached version immediately
          }
          // Fetch from network and cache for next time
          return fetch(event.request)
            .then(fetchResponse => {
              if (fetchResponse.status === 200) {
                const responseClone = fetchResponse.clone();
                caches.open(CACHE_NAME)
                  .then(cache => {
                    cache.put(event.request, responseClone);
                  });
              }
              return fetchResponse;
            });
        })
    );
  }
});