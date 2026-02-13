import './index.css';

document.addEventListener('DOMContentLoaded', () => {
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
});
