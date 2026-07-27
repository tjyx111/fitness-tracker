const STATIC_CACHE = 'fitness-static-v4';
const DATA_CACHE = 'fitness-data-v4';
const STATIC_ASSETS = [
    '/',
    '/styles.css',
    '/app-theme.css',
    '/app.js',
    '/stats.js'
];

self.addEventListener('install', event => {
    event.waitUntil(
        caches.open(STATIC_CACHE)
            .then(cache => cache.addAll(STATIC_ASSETS))
            .then(() => self.skipWaiting())
    );
});

self.addEventListener('activate', event => {
    const currentCaches = new Set([STATIC_CACHE, DATA_CACHE]);
    event.waitUntil(
        caches.keys()
            .then(keys => Promise.all(
                keys.filter(key => !currentCaches.has(key)).map(key => caches.delete(key))
            ))
            .then(() => self.clients.claim())
    );
});

self.addEventListener('fetch', event => {
    const request = event.request;
    if (request.method !== 'GET') {
        return;
    }

    const url = new URL(request.url);
    if (url.origin !== self.location.origin
            || url.pathname === '/api/sync/database'
            || url.pathname.startsWith('/downloads/')) {
        return;
    }

    if (request.mode === 'navigate') {
        event.respondWith(networkFirst(request, STATIC_CACHE, '/'));
        return;
    }

    if (url.pathname.startsWith('/api/')) {
        event.respondWith(networkFirst(request, DATA_CACHE));
        return;
    }

    event.respondWith(networkFirst(request, STATIC_CACHE));
});

async function networkFirst(request, cacheName, fallbackPath) {
    const cache = await caches.open(cacheName);
    try {
        const response = await fetch(request);
        if (response.ok) {
            await cache.put(request, response.clone());
        }
        return response;
    } catch (error) {
        const cached = await cache.match(request);
        if (cached) {
            return cached;
        }
        if (fallbackPath) {
            const fallback = await cache.match(fallbackPath);
            if (fallback) {
                return fallback;
            }
        }
        throw error;
    }
}
