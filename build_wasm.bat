@echo off
set GOOS=js
set GOARCH=wasm

echo Building WASM...
go build -o frontend/chess.wasm main.go

if %errorlevel% equ 0 (
    echo WASM built successfully at frontend/chess.wasm
    if exist wasm_exec.js (
        copy /Y wasm_exec.js frontend\wasm_exec.js
        echo wasm_exec.js updated in frontend/
    ) else (
        echo WARNING: wasm_exec.js not found in root.
    )
) else (
    echo WASM build failed
)
