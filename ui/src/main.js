import './index.css';

// ==========================================
// LibDB: IndexedDB Helper for Library
// ==========================================
const LibDB = {
    DB_NAME: 'LunarLettersDB',
    DB_VERSION: 1,
    db: null,

    async open() {
        if (this.db) return this.db;
        return new Promise((resolve, reject) => {
            const req = indexedDB.open(this.DB_NAME, this.DB_VERSION);
            req.onupgradeneeded = (e) => {
                const d = e.target.result;
                if (!d.objectStoreNames.contains('bookmarks')) {
                    const bStore = d.createObjectStore('bookmarks', { keyPath: 'slug' });
                    bStore.createIndex('addedAt', 'addedAt', { unique: false });
                }
                if (!d.objectStoreNames.contains('history')) {
                    const hStore = d.createObjectStore('history', { keyPath: 'id' });
                    hStore.createIndex('readAt', 'readAt', { unique: false });
                    hStore.createIndex('seriesSlug', 'seriesSlug', { unique: false });
                }
                if (!d.objectStoreNames.contains('progress')) {
                    d.createObjectStore('progress', { keyPath: 'seriesId' });
                }
            };
            req.onsuccess = (e) => {
                this.db = e.target.result;
                resolve(this.db);
            };
            req.onerror = (e) => reject(e.target.error);
        });
    },

    async toggleBookmark(series) {
        await this.open();
        const tx = this.db.transaction('bookmarks', 'readwrite');
        const store = tx.objectStore('bookmarks');

        const existing = await new Promise((resolve) => {
            const req = store.get(series.slug);
            req.onsuccess = () => resolve(req.result);
        });

        if (existing) {
            store.delete(series.slug);
            return false; // Removed
        } else {
            store.put({
                slug: series.slug,
                title: series.title,
                author: series.author,
                cover: series.cover,
                addedAt: Date.now()
            });
            return true; // Added
        }
    },

    async isBookmarked(slug) {
        await this.open();
        return new Promise((resolve) => {
            const tx = this.db.transaction('bookmarks', 'readonly');
            const req = tx.objectStore('bookmarks').get(slug);
            req.onsuccess = () => resolve(!!req.result);
        });
    },

    async addToHistory(chapter) {
        await this.open();
        const tx = this.db.transaction('history', 'readwrite');
        const store = tx.objectStore('history');
        store.put({
            id: chapter.id,
            title: chapter.title,
            seriesTitle: chapter.seriesTitle,
            seriesSlug: chapter.seriesSlug,
            chapterNum: chapter.chapterNum,
            readAt: Date.now()
        });
    }
};

window.LibDB = LibDB;

document.addEventListener('DOMContentLoaded', async () => {
    // Init DB
    try {
        await LibDB.open();
    } catch (e) {
        console.error('Failed to open IndexedDB:', e);
    }

    // ========== Bookmark Button Logic ==========
    const bookmarkBtn = document.getElementById('bookmark-btn');
    if (bookmarkBtn) {
        const slug = bookmarkBtn.dataset.slug;
        const title = bookmarkBtn.dataset.title;
        const author = bookmarkBtn.dataset.author;
        const cover = bookmarkBtn.dataset.cover; // URL or empty

        // Check initial state
        const isBookmarked = await LibDB.isBookmarked(slug);
        updateBookmarkIcon(bookmarkBtn, isBookmarked);

        bookmarkBtn.addEventListener('click', async () => {
            const added = await LibDB.toggleBookmark({ slug, title, author, cover });
            updateBookmarkIcon(bookmarkBtn, added);
        });
    }

    function updateBookmarkIcon(btn, active) {
        if (active) {
            btn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 -960 960 960" width="24" fill="currentColor"><path d="M200-120v-640q0-33 23.5-56.5T280-840h400q33 0 56.5 23.5T760-760v640L480-240 200-120Z"/></svg>`;
            btn.classList.add('active');
        } else {
            btn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 -960 960 960" width="24" fill="currentColor"><path d="M200-120v-640q0-33 23.5-56.5T280-840h400q33 0 56.5 23.5T760-760v640L480-240 200-120Zm80-122 200-86 200 86v-518H280v518Zm0-518h400-400Z"/></svg>`;
            btn.classList.remove('active');
        }
    }

    // ========== Theme Toggle ==========
    const themeToggle = document.getElementById('theme-toggle');
    const readerToggle = document.getElementById('reader-theme-toggle');

    const toggleTheme = () => {
        const current = document.documentElement.getAttribute('data-theme');
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('theme', next);
    };

    if (themeToggle) themeToggle.addEventListener('click', toggleTheme);
    if (readerToggle) readerToggle.addEventListener('click', toggleTheme);

    // ========== Filter Pills (Homepage) ==========
    const pills = document.querySelectorAll('.pill[data-filter]');
    const novelList = document.getElementById('novel-list');

    if (pills.length && novelList) {
        pills.forEach(pill => {
            pill.addEventListener('click', () => {
                pills.forEach(p => p.classList.remove('active'));
                pill.classList.add('active');

                const filter = pill.dataset.filter;
                const rows = novelList.querySelectorAll('.novel-row');

                rows.forEach(row => {
                    if (filter === 'all') {
                        row.style.display = '';
                    } else if (filter === 'ongoing') {
                        const status = (row.dataset.status || '').toLowerCase();
                        const isOngoing = status.includes('ongoing') || status.includes('publishing');
                        row.style.display = isOngoing ? '' : 'none';
                    } else {
                        const status = (row.dataset.status || '').toLowerCase();
                        row.style.display = status.includes(filter) ? '' : 'none';
                    }
                });
            });
        });
    }

    // ========== Smart Header (Hide on scroll down, show on scroll up) ==========
    const header = document.getElementById('app-header');
    if (header) {
        let lastScroll = 0;
        let ticking = false;

        window.addEventListener('scroll', () => {
            if (!ticking) {
                requestAnimationFrame(() => {
                    const currentScroll = window.scrollY;

                    if (currentScroll <= 10) {
                        header.style.transform = '';
                    } else if (currentScroll > lastScroll && currentScroll > 80) {
                        header.style.transform = 'translateY(-100%)';
                        header.style.transition = 'transform 0.3s ease';
                    } else {
                        header.style.transform = 'translateY(0)';
                        header.style.transition = 'transform 0.3s ease';
                    }

                    lastScroll = currentScroll;
                    ticking = false;
                });
                ticking = true;
            }
        }, { passive: true });
    }

    // ========== Reader Controls (immersive toggle) ==========
    const topBar = document.getElementById('top-bar');
    const bottomBar = document.getElementById('bottom-bar');

    if (topBar && bottomBar) {
        let controlsVisible = true;
        let hideTimeout;

        const showControls = () => {
            controlsVisible = true;
            topBar.classList.add('visible');
            bottomBar.classList.add('visible');
            clearTimeout(hideTimeout);
        };

        const hideControls = () => {
            controlsVisible = false;
            topBar.classList.remove('visible');
            bottomBar.classList.remove('visible');
        };

        const scheduleHide = () => {
            clearTimeout(hideTimeout);
            hideTimeout = setTimeout(hideControls, 3000);
        };

        // Show on scroll near edges
        window.addEventListener('scroll', () => {
            const atTop = window.scrollY < 100;
            const atBottom = (window.innerHeight + window.scrollY) >= document.body.offsetHeight - 100;

            if (atTop || atBottom) {
                showControls();
            }
        }, { passive: true });

        // Toggle on tap (not on links/buttons)
        window.addEventListener('click', (e) => {
            if (e.target.closest('button') || e.target.closest('a')) return;
            if (controlsVisible) {
                hideControls();
            } else {
                showControls();
                scheduleHide();
            }
        });

        // Keep visible when hovering bars
        topBar.addEventListener('mouseenter', () => clearTimeout(hideTimeout));
        bottomBar.addEventListener('mouseenter', () => clearTimeout(hideTimeout));
        topBar.addEventListener('mouseleave', scheduleHide);
        bottomBar.addEventListener('mouseleave', scheduleHide);

        // Initial: show then auto-hide
        showControls();
        scheduleHide();
    }

    // ========== Series Page Tab Switching ==========
    const tabs = document.querySelectorAll('.series-tab');
    if (tabs.length) {
        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                tabs.forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));
                tab.classList.add('active');
                const panel = document.getElementById('panel-' + tab.dataset.tab);
                if (panel) panel.classList.add('active');
            });
        });
    }

    // ========== Start Reading Button (Switch to Chapters Tab) ==========
    const startReadingBtn = document.getElementById('start-reading-btn');
    if (startReadingBtn) {
        startReadingBtn.addEventListener('click', () => {
            const chaptersTab = document.querySelector('.series-tab[data-tab="chapters"]');
            if (chaptersTab) {
                chaptersTab.click();
            }
        });
    }
});
