window.chessWorker = new Worker("chess-worker.js");

window.chessWorker.onmessage = function (e) {
    const { type, payload } = e.data;
    if (type === "READY") {
        console.log("Chess Engine (WASM Worker) Ready");
        // Trigger initial game setup after worker is ready
        if (window.onChessWorkerReady) window.onChessWorkerReady();
    }
};
