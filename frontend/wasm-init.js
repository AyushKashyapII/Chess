// Create multiple workers 
const numWorkers = navigator.hardwareConcurrency || 4;
window.chessWorkers = [];
window.readyWorkers = 0;

for (let i = 0; i < numWorkers; i++) {
    const w = new Worker("chess-worker.js");
    w.onmessage = function (e) {
        const { type } = e.data;
        if (type === "READY") {
            window.readyWorkers++;
            console.log(`Chess Engine Worker ${i} Ready (${window.readyWorkers}/${numWorkers})`);
            if (window.readyWorkers === numWorkers) {
                console.log("All Chess Engine Workers Ready");
                if (window.onChessWorkerReady) window.onChessWorkerReady();
            }
        }
    };
    window.chessWorkers.push(w);
}

window.chessWorker = window.chessWorkers[0];
