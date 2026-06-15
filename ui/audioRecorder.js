const NUM_BARS = 32;
const meter = document.getElementById('meter');
const timerEl = document.getElementById('timer');
// const statusLine = document.getElementById('status-line');
const btnRecord = document.getElementById('btn-record');
const btnStop = document.getElementById('btn-stop');
const audioEl = document.getElementById('audio-controls');

// close recorder
document.getElementById("quit-recording").addEventListener("click", () => {
    document.getElementById("recorder-container").classList.add("hidden");
    document.getElementById("recorder-container").classList.remove("flex");
    document.getElementById("activate-voice").classList.add("flex");
    document.getElementById("activate-voice").classList.remove("hidden");
    document.getElementById("prompt-input").classList.remove("hidden");
});

// Build meter bars
const bars = Array.from({ length: NUM_BARS }, () => {
    const b = document.createElement('div');
    b.className = 'bar';
    meter.appendChild(b);
    return b;
});

let mediaRecorder = null;
let chunks = [];
let audioCtx = null;
let analyser = null;
let animFrame = null;
let startTime = null;
let timerFrame = null;

function formatTime(ms) {
    const totalSec = Math.floor(ms / 1000);
    const m = String(Math.floor(totalSec / 60)).padStart(2, '0');
    const s = String(totalSec % 60).padStart(2, '0');
    const cs = String(Math.floor((ms % 1000) / 10)).padStart(2, '0');
    return { m, s, cs };
}

function tickTimer() {
    const { m, s, cs } = formatTime(Date.now() - startTime);
    timerEl.innerHTML = `${m}:${s}<span class="ms">.${cs}</span>`;
    timerFrame = requestAnimationFrame(tickTimer);
}

function drawMeter() {
    if (!analyser) return;
    const data = new Uint8Array(analyser.frequencyBinCount);
    analyser.getByteFrequencyData(data);
    const step = Math.floor(data.length / NUM_BARS);
    bars.forEach((bar, i) => {
        const val = data[i * step] / 255;
        const h = Math.max(3, val * 100);
        bar.style.height = h + '%';
        bar.style.background = val > 0.7
        ? 'var(--danger)'
        : val > 0.4
          ? 'var(--accent)'
          : 'var(--border)';
    });
    animFrame = requestAnimationFrame(drawMeter);
}

function resetMeter() {
    bars.forEach(b => {
        b.style.height = '3px';
        b.style.background = 'var(--border)';
    });
}

async function startRecording() {
    try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

        audioCtx = new AudioContext();
        const src = audioCtx.createMediaStreamSource(stream);
        analyser = audioCtx.createAnalyser();
        analyser.fftSize = 256;
        src.connect(analyser);

        mediaRecorder = new MediaRecorder(stream);
        chunks = [];
        mediaRecorder.ondataavailable = e => chunks.push(e.data);
        mediaRecorder.onstop = () => saveRecording();
        mediaRecorder.start();

        startTime = Date.now();
        tickTimer();
        drawMeter();

        btnRecord.disabled = true;
        btnRecord.classList.add('recording');
        btnStop.disabled = false;
        // statusLine.innerHTML = '<span class="pulse"></span>recording';
    } catch (err) {
        // statusLine.textContent = 'microphone access denied';
        console.error(err);
    }
}

function stopRecording() {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
        mediaRecorder.stop();
        mediaRecorder.stream.getTracks().forEach(t => t.stop());
    }
    cancelAnimationFrame(animFrame);
    cancelAnimationFrame(timerFrame);
    if (audioCtx) { audioCtx.close(); audioCtx = null; analyser = null; }
    resetMeter();
    btnRecord.disabled = false;
    btnRecord.classList.remove('recording');
    btnStop.disabled = true;
    // statusLine.textContent = 'ready';
    timerEl.innerHTML = '00:00<span class="ms">.00</span>';
}

function saveRecording() {
    const duration = Date.now() - startTime;
    const blob = new Blob(chunks, { type: 'audio/webm' });
    const url = URL.createObjectURL(blob);
    audioEl.src = url;
}

btnRecord.addEventListener('click', startRecording);
btnStop.addEventListener('click', stopRecording);
