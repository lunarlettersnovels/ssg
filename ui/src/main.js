import './index.css';

// Theme Toggling
document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.getElementById('theme-toggle');
    const readerToggle = document.getElementById('reader-theme-toggle');

    const toggleTheme = () => {
        const current = document.documentElement.getAttribute('data-theme');
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('theme', next);
        updateIcons(next);
    };

    const updateIcons = (theme) => {
        const icon = theme === 'dark' ? '☀' : '☾';
        if (themeToggle) themeToggle.innerHTML = icon;
        if (readerToggle) readerToggle.innerHTML = icon;
    }

    if (themeToggle) themeToggle.addEventListener('click', toggleTheme);
    if (readerToggle) readerToggle.addEventListener('click', toggleTheme);

    // Init icon
    const saved = localStorage.getItem('theme');
    updateIcons(saved || 'light');

    // Reader Immersive Controls
    const topBar = document.getElementById('top-bar');
    const bottomBar = document.getElementById('bottom-bar');

    if (topBar && bottomBar) {
        let showControls = true;
        let timeout;

        const toggleControls = (force) => {
            if (typeof force === 'boolean') showControls = force;
            else showControls = !showControls;

            if (showControls) {
                topBar.classList.add('visible');
                bottomBar.classList.add('visible');
                if (timeout) clearTimeout(timeout);
            } else {
                topBar.classList.remove('visible');
                bottomBar.classList.remove('visible');
            }
        };

        // Scroll logic (show on bottom)
        window.addEventListener('scroll', () => {
            if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 50) {
                toggleControls(true);
            }
        });

        // Click logic
        window.addEventListener('click', (e) => {
            if (e.target.closest('button') || e.target.closest('a')) return;
            toggleControls();
        });

        // Auto hide after 2s
        setTimeout(() => toggleControls(false), 2000);
    }
});
