const $ = (id) => document.getElementById(id);
const PREFS_KEY = "cartitas:prefs";
const AGE_ORDER = ["toddler", "early", "middle", "teen"];

const state = {
  cards: [],
  queue: [],
  index: 0,
  session: { correct: 0, missed: 0 },
};

function prefs() {
  return {
    lang: $("lang").value,
    topic: $("topic").value,
    age: $("age").value,
  };
}

function savePrefs() {
  localStorage.setItem(PREFS_KEY, JSON.stringify(prefs()));
}

function readStoredPrefs() {
  const params = new URLSearchParams(location.search);
  let stored = {};
  try {
    stored = JSON.parse(localStorage.getItem(PREFS_KEY) || "{}");
  } catch {
    stored = {};
  }
  return {
    lang: params.get("lang") || stored.lang || "en",
    topic: params.get("topic") || stored.topic || "",
    age: params.get("age") || stored.age || "",
  };
}

function applySelect(select, value) {
  if ([...select.options].some((o) => o.value === value)) select.value = value;
}

function unique(values) {
  return [...new Set(values.filter(Boolean))].sort((a, b) => a.localeCompare(b));
}

function fillSelect(select, values, allLabel) {
  select.innerHTML = "";
  const all = document.createElement("option");
  all.value = "";
  all.textContent = allLabel;
  select.append(all);
  for (const value of values) {
    const opt = document.createElement("option");
    opt.value = value;
    opt.textContent = value;
    select.append(opt);
  }
}

function filtered() {
  const { lang, topic, age } = prefs();
  return state.cards.filter((c) => {
    if (c.lang !== lang) return false;
    if (topic && c.topic !== topic) return false;
    if (age && c.age !== age) return false;
    return true;
  });
}

function refreshFilters() {
  const lang = $("lang").value;
  const pool = state.cards.filter((c) => c.lang === lang);
  const wanted = readStoredPrefs();
  fillSelect($("topic"), unique(pool.map((c) => c.topic)), "All topics");
  const ages = unique(pool.map((c) => c.age)).sort(
    (a, b) => AGE_ORDER.indexOf(a) - AGE_ORDER.indexOf(b)
  );
  fillSelect($("age"), ages, "All ages");
  applySelect($("topic"), wanted.topic);
  applySelect($("age"), wanted.age);
  const n = filtered().length;
  $("status").textContent = n ? `${n} cards ready.` : "No cards for that mix.";
}

async function fetchJSON(path) {
  const res = await fetch(path, { cache: "no-store" });
  if (!res.ok) throw new Error(`${path} ${res.status}`);
  return res.json();
}

function shuffle(items) {
  const copy = items.slice();
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}

function showCard() {
  const card = state.queue[state.index];
  if (!card) {
    $("question").textContent = "Session done.";
    $("answer").hidden = true;
    $("reveal-row").hidden = true;
    $("grade-row").hidden = true;
    $("progress").textContent =
      `Got it: ${state.session.correct} · Missed: ${state.session.missed}`;
    return;
  }
  $("deck").textContent = `${card.topic} · ${card.subtopic} · ${card.age}`;
  $("progress").textContent = `${state.index + 1} / ${state.queue.length}`;
  $("question").textContent = card.question;
  $("answer").textContent = card.answer;
  $("answer").hidden = true;
  $("reveal-row").hidden = false;
  $("grade-row").hidden = true;
}

function start() {
  savePrefs();
  const pool = filtered();
  if (!pool.length) {
    $("status").textContent = "No cards for that mix.";
    return;
  }
  state.queue = shuffle(pool).slice(0, 12);
  state.index = 0;
  state.session = { correct: 0, missed: 0 };
  $("drill").hidden = false;
  $("status").textContent = "";
  showCard();
}

function reveal() {
  $("answer").hidden = false;
  $("reveal-row").hidden = true;
  $("grade-row").hidden = false;
}

function throwConfetti() {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  const canvas = $("confetti");
  const ctx = canvas.getContext("2d");
  const dpr = window.devicePixelRatio || 1;
  canvas.hidden = false;
  canvas.width = Math.floor(window.innerWidth * dpr);
  canvas.height = Math.floor(window.innerHeight * dpr);
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  const colors = ["#c45c26", "#2f6f4e", "#e6b422", "#4a7fd4", "#e07a9a", "#fffdf8"];
  const origin = $("got-it").getBoundingClientRect();
  const x0 = origin.left + origin.width / 2;
  const y0 = origin.top + origin.height / 2;
  const bits = Array.from({ length: 90 }, () => {
    const angle = -Math.PI / 2 + (Math.random() - 0.5) * Math.PI;
    const speed = 6 + Math.random() * 10;
    return {
      x: x0,
      y: y0,
      vx: Math.cos(angle) * speed,
      vy: Math.sin(angle) * speed - 4,
      w: 6 + Math.random() * 6,
      h: 8 + Math.random() * 8,
      rot: Math.random() * Math.PI,
      vr: (Math.random() - 0.5) * 0.4,
      color: colors[Math.floor(Math.random() * colors.length)],
    };
  });
  const started = performance.now();
  function frame(now) {
    const t = (now - started) / 1000;
    ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    for (const bit of bits) {
      bit.vy += 0.28;
      bit.x += bit.vx;
      bit.y += bit.vy;
      bit.rot += bit.vr;
      ctx.save();
      ctx.translate(bit.x, bit.y);
      ctx.rotate(bit.rot);
      ctx.globalAlpha = Math.max(0, 1 - t / 1.2);
      ctx.fillStyle = bit.color;
      ctx.fillRect(-bit.w / 2, -bit.h / 2, bit.w, bit.h);
      ctx.restore();
    }
    if (t < 1.2) {
      requestAnimationFrame(frame);
    } else {
      ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
      canvas.hidden = true;
    }
  }
  requestAnimationFrame(frame);
}

function grade(gotIt) {
  if (gotIt) {
    state.session.correct += 1;
    throwConfetti();
  } else {
    state.session.missed += 1;
  }
  state.index += 1;
  showCard();
}

function onFilterChange() {
  savePrefs();
  $("status").textContent = `${filtered().length} cards ready.`;
}

async function init() {
  try {
    state.cards = (await fetchJSON("data/cards.json")).cards || [];
  } catch {
    $("status").textContent = "Could not load cards.json.";
    return;
  }
  applySelect($("lang"), readStoredPrefs().lang);
  refreshFilters();
}

$("lang").addEventListener("change", () => {
  savePrefs();
  refreshFilters();
});
$("topic").addEventListener("change", onFilterChange);
$("age").addEventListener("change", onFilterChange);
$("start").addEventListener("click", start);
$("reveal").addEventListener("click", reveal);
$("got-it").addEventListener("click", () => grade(true));
$("missed").addEventListener("click", () => grade(false));
$("end").addEventListener("click", () => {
  state.queue = [];
  $("drill").hidden = true;
});

init();
