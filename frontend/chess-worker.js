importScripts("wasm_exec.js");

const go = new Go();
let wasmInstance;

WebAssembly.instantiateStreaming(fetch("chess.wasm"), go.importObject).then((result) => {
    wasmInstance = result.instance;
    go.run(wasmInstance);
    postMessage({ type: "READY" });
}).catch((err) => {
    console.error("Worker WASM load failed:", err);
});

onmessage = function (e) {
    const { type, payload } = e.data;

    switch (type) {
        case "INIT_BOARD":
            console.log("Worker: Initializing board");
            self.init_board_wasm(payload.fen);
            break;
        case "VALIDATE_MOVE":
            console.log("Worker: Validating move string", payload);
            // Use string format like "e2e4"
            const valResult = self.validate_move_string_wasm(payload.moveString);
            postMessage({ type: "VALIDATE_MOVE_RESULT", payload: valResult });
            break;
        case "GET_AI_MOVE":
            console.log("Worker: Getting AI move");
            // Use string format function
            const aiResult = self.get_ai_move_string_wasm();
            postMessage({ type: "GET_AI_MOVE_RESULT", payload: aiResult });
            break;
    }
};
