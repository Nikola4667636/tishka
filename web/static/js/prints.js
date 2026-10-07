// prints.js — горизонтальный скролл коллекций принтов.
// Работает независимо от app.js: запускается на любой странице,
// где есть .prints-strip, и тихо выходит, если их нет.

(function () {
  const strips = document.querySelectorAll('[data-role="strip"]');
  if (!strips.length) return;

  strips.forEach((strip) => {
    const track = strip.querySelector('[data-role="strip-track"]');
    const btnPrev = strip.querySelector('[data-role="strip-prev"]');
    const btnNext = strip.querySelector('[data-role="strip-next"]');

    // Шаг прокрутки = ширина одной карточки + gap.
    const step = () => {
      const card = track.querySelector('.print-card');
      if (!card) return 260;
      const gap = 20;
      return card.getBoundingClientRect().width + gap;
    };

    const updateButtons = () => {
      const maxScroll = track.scrollWidth - track.clientWidth - 1;
      btnPrev.disabled = track.scrollLeft <= 0;
      btnNext.disabled = track.scrollLeft >= maxScroll;
    };

    btnPrev.addEventListener('click', () => {
      track.scrollBy({ left: -step() * 2, behavior: 'smooth' });
    });

    btnNext.addEventListener('click', () => {
      track.scrollBy({ left: step() * 2, behavior: 'smooth' });
    });

    track.addEventListener('scroll', updateButtons, { passive: true });
    window.addEventListener('resize', updateButtons);

    // Drag-to-scroll мышью (на тач-устройствах работает нативно)
    let isDown = false;
    let startX = 0;
    let startScroll = 0;

    track.addEventListener('mousedown', (e) => {
      isDown = true;
      track.classList.add('is-dragging');
      startX = e.pageX;
      startScroll = track.scrollLeft;
      e.preventDefault();
    });

    const stop = () => {
      isDown = false;
      track.classList.remove('is-dragging');
    };

    track.addEventListener('mouseleave', stop);
    track.addEventListener('mouseup', stop);

    track.addEventListener('mousemove', (e) => {
      if (!isDown) return;
      const walk = e.pageX - startX;
      track.scrollLeft = startScroll - walk;
    });

    updateButtons();
  });
})();