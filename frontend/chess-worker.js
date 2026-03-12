importScripts("wasm_exec.js");

const go = new Go();
let wasmInstance;

// Some static hosts don’t serve chess.wasm with the correct MIME type,
// which breaks instantiateStreaming. Fall back to fetch+instantiate.
(async () => {
    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch("chess.wasm"),
            go.importObject
        );
        wasmInstance = result.instance;
        go.run(wasmInstance);
        postMessage({ type: "READY" });
    } catch (err) {
        console.warn("instantiateStreaming failed, falling back to arrayBuffer:", err);
        try {
            const resp = await fetch("chess.wasm");
            const bytes = await resp.arrayBuffer();
            const result = await WebAssembly.instantiate(bytes, go.importObject);
            wasmInstance = result.instance;
            go.run(wasmInstance);
            postMessage({ type: "READY" });
        } catch (err2) {
            console.error("Worker WASM load failed:", err2);
            postMessage({ type: "ERROR", error: String(err2) });
        }
    }
})();

onmessage = function (e) {
    const { type, payload, fen, movesToSearch } = e.data;

    switch (type) {
        case "INIT_BOARD":
            console.log("Worker: Initializing board with FEN:", payload.fen);
            self.init_board_wasm(payload.fen);
            postMessage({ type: "INIT_BOARD_RESULT" });
            break;
        case "VALIDATE_MOVE":
            console.log("Worker: Validating move string", payload);
            const valResult = self.validate_move_string_wasm(payload.moveString);
            postMessage({ type: "VALIDATE_MOVE_RESULT", payload: valResult });
            break;
        case "GET_AI_MOVE":
            console.log("Worker: Getting AI move");
            const aiResult = self.get_ai_move_string_wasm();
            postMessage({ type: "GET_AI_MOVE_RESULT", payload: aiResult });
            break;
        case "GET_ALL_MOVES":
            console.log("Worker: Getting all legal moves for FEN:", fen);
            const isWhiteTurn = e.data.isWhiteTurn !== undefined ? e.data.isWhiteTurn : false;
            const movesJson = self.get_all_legal_moves_wasm(fen, isWhiteTurn);
            postMessage({ type: "GET_MOVES_RESULT", data: movesJson });
            break;
        case "SEARCH_SUBSET":
            // Root splitting: search only the assigned moves
            console.log("Worker: Searching subset of moves", movesToSearch);
            const resultJson = self.search_subset_wasm(fen, JSON.stringify(movesToSearch));
            postMessage({ type: "SEARCH_SUBSET_RESULT", data: resultJson });
            break;
        case "APPLY_MOVE":
            // Apply a move and return new FEN
            const moveJson = e.data.moveJson;
            const applyResult = self.apply_move_wasm(e.data.fen, moveJson);
            postMessage({ type: "APPLY_MOVE_RESULT", data: applyResult });
            break;
    }
};
