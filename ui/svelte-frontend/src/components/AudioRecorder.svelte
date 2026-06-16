<script>
    import { createEventDispatcher, onDestroy } from "svelte";

    const dispatch = createEventDispatcher();
    const NUM_BARS = 32;

    let bars = Array(NUM_BARS).fill(3); // heights in %
    let barColors = Array(NUM_BARS).fill("var(--border)");
    let mediaRecorder = null;
    let chunks = [];
    let audioCtx = null;
    let analyser = null;
    let animFrame = null;
    let timerFrame = null;
    let startTime = null;
    let timerDisplay = '00:00<span class="ms">.00</span>';
    let audioSrc = "";
    let recording = false;
    let stopped = true;

    function formatTime(ms) {
        const totalSec = Math.floor(ms / 1000);
        const m = String(Math.floor(totalSec / 60)).padStart(2, "0");
        const s = String(totalSec % 60).padStart(2, "0");
        const cs = String(Math.floor((ms % 1000) / 10)).padStart(2, "0");
        return { m, s, cs };
    }

    function tickTimer() {
        const { m, s, cs } = formatTime(Date.now() - startTime);
        timerDisplay = `${m}:${s}.${cs}`;
        timerFrame = requestAnimationFrame(tickTimer);
    }

    function drawMeter() {
        if (!analyser) return;
        const data = new Uint8Array(analyser.frequencyBinCount);
        analyser.getByteFrequencyData(data);
        const step = Math.floor(data.length / NUM_BARS);
        bars = bars.map((_, i) => {
            const val = data[i * step] / 255;
            return Math.max(3, val * 100);
        });
        barColors = bars.map((h, i) => {
            const val = (data[i * step] ?? 0) / 255;
            return val > 0.7
                ? "var(--danger)"
                : val > 0.4
                  ? "var(--accent)"
                  : "var(--border)";
        });
        animFrame = requestAnimationFrame(drawMeter);
    }

    async function startRecording() {
        const stream = await navigator.mediaDevices.getUserMedia({
            audio: true,
        });
        audioCtx = new AudioContext();
        const src = audioCtx.createMediaStreamSource(stream);
        analyser = audioCtx.createAnalyser();
        analyser.fftSize = 256;
        src.connect(analyser);

        mediaRecorder = new MediaRecorder(stream);
        chunks = [];
        mediaRecorder.ondataavailable = (e) => chunks.push(e.data);
        mediaRecorder.onstop = () => {
            const blob = new Blob(chunks, { type: "audio/webm" });
            audioSrc = URL.createObjectURL(blob);
        };
        mediaRecorder.start();

        startTime = Date.now();
        recording = true;
        stopped = false;
        tickTimer();
        drawMeter();
    }

    function stopRecording() {
        if (mediaRecorder?.state !== "inactive") {
            mediaRecorder.stop();
            mediaRecorder.stream.getTracks().forEach((t) => t.stop());
        }
        cancelAnimationFrame(animFrame);
        cancelAnimationFrame(timerFrame);
        audioCtx?.close();
        audioCtx = null;
        analyser = null;
        bars = Array(NUM_BARS).fill(3);
        barColors = Array(NUM_BARS).fill("var(--border)");
        recording = false;
        stopped = true;
        timerDisplay = "00:00.00";
    }

    onDestroy(stopRecording);
</script>

<div id="recorder-container" class="flex">
    <button
        id="quit-recording"
        on:click={() => {
            stopRecording();
            dispatch("close");
        }}>X</button
    >
    <button class="rec-btn" on:click={startRecording} disabled={recording}
        >● Rec</button
    >
    <button class="rec-btn" on:click={stopRecording} disabled={stopped}
        >■ Stop</button
    >
    <div class="timer">{timerDisplay}</div>
    <div class="meter-wrap">
        {#each bars as h, i}
            <div
                class="bar"
                style="height: {h}%; background: {barColors[i]}"
            ></div>
        {/each}
    </div>
    {#if audioSrc}
        <audio controls src={audioSrc}></audio>
    {/if}
</div>
