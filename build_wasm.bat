@echo off
set GOOS=js
set GOARCH=wasm
go build -o frontend/chess.wasm main.go
if %errorlevel% equ 0 (
    echo WASM built successfully at frontend/chess.wasm
) else (
    echo WASM build failed
)
