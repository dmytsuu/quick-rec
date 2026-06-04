const PORT = 9999;

const CONTROL_SELECTORS = [
  '[data-a-target="player-controls-right"]',
  '.player-controls__right-control-group',
  '[class*="player-controls"] [class*="right"]',
];

function findControlsBar() {
  for (const sel of CONTROL_SELECTORS) {
    const el = document.querySelector(sel);
    if (el) return el;
  }
  return null;
}

function injectButton() {
  if (document.getElementById('quickrec-btn')) return;

  const bar = findControlsBar();
  if (!bar) return;

  const btn = document.createElement('button');
  btn.id = 'quickrec-btn';
  btn.textContent = '⏺ REC';
  btn.title = 'Start yt-dlp recording';
  btn.style.cssText = [
    'background: transparent',
    'color: #fff',
    'border: none',
    'border-radius: 4px',
    'padding: 0 10px',
    'height: 30px',
    'cursor: pointer',
    'font-size: 12px',
    'font-weight: 700',
    'margin: 0 4px',
    'transition: background 0.15s',
  ].join(';');

  btn.addEventListener('mouseenter', () => { btn.style.background = 'rgba(255,255,255,0.1)'; });
  btn.addEventListener('mouseleave', () => { btn.style.background = 'transparent'; });

  btn.addEventListener('click', async () => {
    const url = window.location.href.split('?')[0];

    try {
      const res = await fetch(
        `http://127.0.0.1:${PORT}/run?url=${encodeURIComponent(url)}`
      );
      if (res.ok) {
        btn.textContent = '✓ Started';
        btn.style.background = '#00c44f';
      } else {
        btn.textContent = '✗ Error';
        btn.style.background = '#e91916';
      }
    } catch {
      btn.textContent = '✗ Offline';
      btn.style.background = '#e91916';
    }

    setTimeout(() => {
      btn.textContent = '⏺ REC';
      btn.style.background = 'transparent';
    }, 2500);
  });

  bar.prepend(btn);
}

// Re-inject on SPA navigation (Twitch React SPA — DOM resets between channels)
let lastPath = location.pathname;
const observer = new MutationObserver(() => {
  if (location.pathname !== lastPath) {
    lastPath = location.pathname;
    document.getElementById('quickrec-btn')?.remove();
  }
  injectButton();
});
observer.observe(document.body, { childList: true, subtree: true });

injectButton();
