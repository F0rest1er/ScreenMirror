document.addEventListener("DOMContentLoaded", () => {
  const fullScreenBtn = document.getElementById("fullScreenBtn");
  const bodyEl = document.body;

  fullScreenBtn.addEventListener("click", () => {
    if (!document.fullscreenElement) {
      document.documentElement
        .requestFullscreen()
        .then(() => {
          bodyEl.classList.add("is-fullscreen");
        })
        .catch((err) => {
          console.error(`Ошибка при переходе в полноэкранный режим: ${err.message} (${err.name})`);
        });
    } else {
      document.exitFullscreen().then(() => {
        bodyEl.classList.remove("is-fullscreen");
      });
    }
  });

  document.addEventListener("fullscreenchange", () => {
    if (!document.fullscreenElement) {
      bodyEl.classList.remove("is-fullscreen");
    }
  });

  const loginForm = document.getElementById("loginForm");
  const viewerContent = document.getElementById("viewerContent");
  const streamVideo = document.getElementById("streamVideo");
  const passwordInput = document.getElementById("passwordInput");
  const loginBtn = document.getElementById("loginBtn");
  const loginError = document.getElementById("loginError");
  const streamLoader = document.getElementById("streamLoader");

  if (streamVideo && streamLoader) {
    streamVideo.addEventListener("load", () => {
      streamLoader.style.display = "none";
    });
    streamVideo.addEventListener("error", () => {
      streamLoader.textContent = "Ошибка загрузки потока.";
      streamLoader.style.display = "block";
    });
  }

  const noSleep = new NoSleep();
  const cursorEl = document.getElementById("customCursor");
  let pc = null;
  if (typeof PerfectCursor !== "undefined") {
    pc = new PerfectCursor.PerfectCursor((point) => {
      cursorEl.style.transform = `translate(${point[0]}px, ${point[1]}px)`;
    });
  }

  function startCursorStream() {
    cursorEl.style.display = "block";
    const evtSource = new EventSource("/api/cursor");
    evtSource.onmessage = function (event) {
      const data = JSON.parse(event.data);
      const rect = streamVideo.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0 || data.w === 0 || data.h === 0) return;
      
      const scaleX = rect.width / data.w;
      const scaleY = rect.height / data.h;
      
      let x = rect.left + data.x * scaleX;
      let y = rect.top + data.y * scaleY;
      
      if (x < rect.left) x = rect.left;
      if (x > rect.right) x = rect.right;
      if (y < rect.top) y = rect.top;
      if (y > rect.bottom) y = rect.bottom;

      if (pc) {
        pc.addPoint([x, y]);
      } else {
        cursorEl.style.transform = `translate(${x}px, ${y}px)`;
      }
    };
  }

  async function sha256(message) {
    const K = [
      0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
      0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da, 0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
      0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070, 0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
      0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
    ];
    const H0 = [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
    let words = [];
    let strLen = message.length * 8;
    message += "\x80";
    while (message.length % 64 !== 56) message += "\x00";
    for (let i = 0; i < message.length; i++) {
      words[i >> 2] |= (message.charCodeAt(i) & 0xff) << ((3 - (i % 4)) * 8);
    }
    words[words.length] = (strLen / Math.pow(2, 32)) | 0;
    words[words.length] = strLen | 0;

    let H = H0.slice(0);
    for (let i = 0; i < words.length; i += 16) {
      let w = words.slice(i, i + 16);
      let a = H[0],
        b = H[1],
        c = H[2],
        d = H[3],
        e = H[4],
        f = H[5],
        g = H[6],
        h = H[7];

      for (let j = 0; j < 64; j++) {
        if (j >= 16) {
          let w15 = w[j - 15],
            w2 = w[j - 2];
          let s0 = ((w15 >>> 7) | (w15 << 25)) ^ ((w15 >>> 18) | (w15 << 14)) ^ (w15 >>> 3);
          let s1 = ((w2 >>> 17) | (w2 << 15)) ^ ((w2 >>> 19) | (w2 << 13)) ^ (w2 >>> 10);
          w[j] = (w[j - 16] + s0 + w[j - 7] + s1) | 0;
        }
        let S1 = ((e >>> 6) | (e << 26)) ^ ((e >>> 11) | (e << 21)) ^ ((e >>> 25) | (e << 7));
        let ch = (e & f) ^ (~e & g);
        let temp1 = (h + S1 + ch + K[j] + w[j]) | 0;
        let S0 = ((a >>> 2) | (a << 30)) ^ ((a >>> 13) | (a << 19)) ^ ((a >>> 22) | (a << 10));
        let maj = (a & b) ^ (a & c) ^ (b & c);
        let temp2 = (S0 + maj) | 0;

        h = g;
        g = f;
        f = e;
        e = (d + temp1) | 0;
        d = c;
        c = b;
        b = a;
        a = (temp1 + temp2) | 0;
      }

      H[0] = (H[0] + a) | 0;
      H[1] = (H[1] + b) | 0;
      H[2] = (H[2] + c) | 0;
      H[3] = (H[3] + d) | 0;
      H[4] = (H[4] + e) | 0;
      H[5] = (H[5] + f) | 0;
      H[6] = (H[6] + g) | 0;
      H[7] = (H[7] + h) | 0;
    }

    let hex = "";
    for (let i = 0; i < 8; i++) {
      for (let j = 3; j >= 0; j--) {
        let b = (H[i] >>> (j * 8)) & 0xff;
        hex += (b < 16 ? "0" : "") + b.toString(16);
      }
    }
    return hex;
  }

  async function performLogin() {
    const password = passwordInput.value.trim();
    if (!password) {
      loginError.textContent = "Введите пароль";
      return;
    }

    loginBtn.disabled = true;
    loginError.textContent = "";

    try {
      const chalRes = await fetch("/api/auth/challenge");
      if (!chalRes.ok) throw new Error("Не удалось получить challenge");
      const chalData = await chalRes.json();
      const nonce = chalData.nonce;

      const hash = await sha256(password + nonce);

      const verifyRes = await fetch("/api/auth/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nonce, hash }),
      });

      if (verifyRes.ok) {
        loginForm.style.display = "none";
        viewerContent.style.display = "flex";
        streamVideo.src = "/stream";
        startCursorStream();
      } else {
        loginError.textContent = "Неверный пароль или вы заблокированы";
        passwordInput.value = "";
      }
    } catch (err) {
      loginError.textContent = "Ошибка сети";
    } finally {
      loginBtn.disabled = false;
    }
  }

  document.addEventListener("click", function enableNoSleep() {
    document.removeEventListener("click", enableNoSleep, false);
    noSleep.enable();
  }, false);

  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible" && viewerContent.style.display === "flex") {
      noSleep.enable();
    }
  });

  loginBtn.addEventListener("click", performLogin);
  passwordInput.addEventListener("keypress", (e) => {
    if (e.key === "Enter") performLogin();
  });
});
