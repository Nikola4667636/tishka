// app.js — вся клиентская логика прохождения теста.
//
// Устроено сознательно просто (без сборки, без фреймворка): один файл,
// один объект-состояние, несколько маленьких функций рендера. Для сайта
// с парой страниц и одним интерактивным экраном это самый быстрый путь
// от идеи до рабочего прототипа. Если фронт начнёт расти (больше 2-3
// интерактивных экранов, общий стейт между страницами) — это первое
// место, которое стоит переписать на что-то вроде отдельных ES-модулей
// или лёгкого фреймворка; сама HTML-разметка (data-role атрибуты) для
// этого уже готова, переписывать её не придётся.

(function () {
  const dataEl = document.getElementById("test-data");
  if (!dataEl) {
    // Не страница теста — например, главная или список тестов. Выходим.
    return;
  }

  const test = JSON.parse(dataEl.textContent);

  const root = document.querySelector(".quiz");
  const els = {
    progressBar: root.querySelector('[data-role="progress-bar"]'),
    questionView: root.querySelector('[data-role="question-view"]'),
    counter: root.querySelector('[data-role="counter"]'),
    questionText: root.querySelector('[data-role="question-text"]'),
    options: root.querySelector('[data-role="options"]'),
    btnBack: root.querySelector('[data-role="btn-back"]'),
    btnNext: root.querySelector('[data-role="btn-next"]'),
    resultView: root.querySelector('[data-role="result-view"]'),
    resultImage: root.querySelector('[data-role="result-image"]'),
    resultName: root.querySelector('[data-role="result-name"]'),
    resultDescription: root.querySelector('[data-role="result-description"]'),
    btnRetry: root.querySelector('[data-role="btn-retry"]'),
    errorView: root.querySelector('[data-role="error-view"]'),
  };

  /** @type {Record<string,string>} questionID -> optionID */
  let answers = {};
  let currentIndex = 0;

  function currentQuestion() {
    return test.questions[currentIndex];
  }

  function renderQuestion() {
    hideError();

    const q = currentQuestion();
    const selected = answers[q.id];

    els.counter.textContent = `Вопрос ${currentIndex + 1} из ${test.questions.length}`;
    els.questionText.textContent = q.text;
    els.progressBar.style.width = `${(currentIndex / test.questions.length) * 100}%`;

    els.options.innerHTML = "";
    q.options.forEach((opt) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "quiz-option" + (opt.id === selected ? " is-selected" : "");
      btn.textContent = opt.text;
      btn.addEventListener("click", () => selectOption(q.id, opt.id));
      els.options.appendChild(btn);
    });

    els.btnBack.style.visibility = currentIndex === 0 ? "hidden" : "visible";
    els.btnNext.disabled = !selected;
    els.btnNext.textContent =
      currentIndex === test.questions.length - 1 ? "Узнать результат →" : "Далее →";
  }

  function selectOption(questionID, optionID) {
    answers[questionID] = optionID;
    renderQuestion();
  }

  function goBack() {
    if (currentIndex === 0) return;
    currentIndex -= 1;
    renderQuestion();
  }

  function goNext() {
    const isLast = currentIndex === test.questions.length - 1;
    if (isLast) {
      submitAnswers();
      return;
    }
    currentIndex += 1;
    renderQuestion();
  }

  async function submitAnswers() {
    els.btnNext.disabled = true;
    els.btnNext.textContent = "Считаем...";
    hideError();

    try {
      const res = await fetch(`/api/tests/${encodeURIComponent(test.slug)}/result`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ answers }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `Сервер ответил ${res.status}`);
      }

      const result = await res.json();
      showResult(result);
    } catch (err) {
      showError(err.message || "Что-то пошло не так. Попробуй ещё раз.");
      renderQuestion();
    }
  }

  function showResult(result) {
    els.progressBar.style.width = "100%";
    els.questionView.hidden = true;
    els.resultView.hidden = false;

    els.resultImage.src = result.image_url;
    els.resultImage.alt = result.name;
    els.resultName.textContent = result.name;
    els.resultDescription.textContent = result.description;
  }

  function retry() {
    answers = {};
    currentIndex = 0;
    els.progressBar.style.width = "0%";
    els.resultView.hidden = true;
    els.questionView.hidden = false;
    renderQuestion();
  }

  function showError(message) {
    els.errorView.textContent = message;
    els.errorView.hidden = false;
  }

  function hideError() {
    els.errorView.hidden = true;
  }

  els.btnBack.addEventListener("click", goBack);
  els.btnNext.addEventListener("click", goNext);
  els.btnRetry.addEventListener("click", retry);

  renderQuestion();
})();