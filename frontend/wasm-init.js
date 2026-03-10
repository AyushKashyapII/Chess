const go = new Go();
WebAssembly.instantiateStreaming(fetch("chess.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    console.log("Go WebAssembly Initialized");
}).catch((err) => {
    console.error("Failed to load WASM:", err);
});
