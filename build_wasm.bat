@echo off
set GOOS=js
set GOARCH=wasm

echo Building WASM...
go build -o frontend/chess.wasm main.go

if %errorlevel% equ 0 (
    echo WASM built successfully at frontend/chess.wasm
    copy wasm_exec.js frontend\wasm_exec.js
    echo wasm_exec.js copied to frontend/
) else (
    echo WASM build failed
)
