# Go Chess Engine

## 📖 About The Project

This project is a classic chess engine built from the ground up in the **Go** programming language. It started as a personal learning journey to master Go's fundamentals and has evolved from a simple Fyne-based desktop application into:

- a **CLI engine** you can play against in the terminal, and  
- a **browser-based engine** compiled to **WebAssembly (WASM)** with a modern HTML/JS frontend.

The engine features a robust AI powered by a Minimax search algorithm with several advanced optimizations (alpha–beta pruning, quiescence search, transposition tables, aspiration search, and root-splitting parallelism in the browser).

The primary goal of this project was to learn and implement the core concepts of chess engine development, including move generation, board evaluation, search algorithms, and deployment strategies for both native and web environments.

---

## ✨ Features

The engine includes several key features that are standard in modern chess AI:

### AI Engine
- **Search Algorithm**: A **Minimax** core that explores the game tree to find the optimal move.
- **Alpha–Beta Pruning**: Dramatically reduces the search space, allowing deeper searches in the same time.
- **Quiescence Search**: Extends leaf nodes in noisy positions (captures, tactics) to reduce the horizon effect.
- **Transposition Table**: Uses **Zobrist hashing** and a fixed-size transposition table to cache scores and best moves.
- **Move Ordering**: Moves are scored with a custom `score_move` heuristic and ordered to improve alpha–beta efficiency.
- **Aspiration Search**: Iterative deepening with a narrow search window around the previous iteration’s score for speed.
- **Root Splitting (Browser)**: In the WASM build, the root move list is split across **multiple Web Workers** so different move branches are searched in parallel, utilizing multi-core CPUs.

### Board Evaluation
- **Material Advantage**: Standard material values (P,Q,R,B,N,p) are the base of the evaluation.
- **Piece-Square Tables (PSTs)**: Position-dependent bonuses/penalties for every piece type (including king phase tables), so, e.g., knights are rewarded in the center and pawns for advancing.

### Frontends

- **CLI Engine (`engine_cli.go`)**
  - Play against the engine directly in the terminal.
  - Uses simple text input like `e2e4` for moves.
  - Prints the board, engine move, timing, and profiling info for each engine move.

- **Browser UI (`frontend/`)**
  - **WASM Engine**: Core engine is compiled to `frontend/chess.wasm` and loaded via `wasm_exec.js` and `wasm-init.js`.
  - **Web Workers**:
    - `chess-worker.js` hosts the Go WASM runtime and exposes JS-visible functions:
      - `init_board_wasm(fen)` – set board from FEN.
      - `validate_move_wasm` / `validate_move_string_wasm` – validate human moves.
      - `get_ai_move_wasm` / `get_ai_move_string_wasm` – compute best engine move.
      - `get_all_legal_moves_wasm` – enumerate all legal moves for a side from a FEN.
      - `search_subset_wasm` – search a specific subset of root moves (used for root splitting).
      - `apply_move_wasm` – apply a move (including castling, en passant, promotion) and return the new FEN.
    - Multiple workers are spawned so root moves can be searched in parallel.
  - **UI Logic (`script.js`)**:
    - Renders the board and pieces from a FEN string.
    - Converts clicks to algebraic moves (`e2e4`) and sends them to the worker for validation.
    - Maintains the **single source of truth** for the position as a FEN string, always updated from Go/WASM.
    - Manages **root-splitting parallel search**:
      - Gets all legal moves via `get_all_legal_moves_wasm`.
      - Splits the move list into chunks and distributes to workers via `SEARCH_SUBSET`.
      - Each worker searches only its chunk and returns a best move + score.
      - The main thread picks the globally best score and applies that move via `APPLY_MOVE`, updating the FEN and board state.

---

## 🛠️ Tech Stack

- **Engine / Core**: [Go](https://golang.org/)
- **WASM Runtime**: `GOOS=js GOARCH=wasm` + `wasm_exec.js`
- **Frontend**: HTML5, CSS3, vanilla JavaScript
- **Concurrency (Browser)**: Web Workers + message passing (root splitting)

---

## 🚀 How to Run Locally

There are two main ways to run the engine: **CLI** and **Browser (WASM)**.

### 1. CLI Engine (Terminal)

#### Prerequisites
- Go 1.18+ installed

#### Steps
1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/your-repo-name.git
   cd your-repo-name/Chess
   ```

2. **Run the CLI engine:**
   ```bash
   go run engine_cli.go
   ```

3. **Play vs engine:**
   - You are White by default.
   - Enter moves in **UCI-like** format, e.g.:
     ```text
     e2e4
     g1f3
     ```
   - The engine responds with its move, prints timing and profiling stats (`FindBestMove`, `Minimax`, `QuiescenceSearch`, move generation timings), and shows the updated board.

### 2. Browser Engine (WASM + Frontend)

#### Prerequisites
- Go 1.18+
- Any static file server (or VS Code Live Server, `python -m http.server`, etc.)

#### Build WASM
From the `Chess` directory:
```bash
build_wasm.bat
```
This will:
- Compile `main.go` to `frontend/chess.wasm` using `GOOS=js` / `GOARCH=wasm`.
- Copy `wasm_exec.js` into `frontend/` (if present at project root).

#### Serve the Frontend
From the `Chess/frontend` directory, run any static server. For example, with Python:
```bash
cd frontend
python -m http.server 8000
```
Then open:
```text
http://localhost:8000/index.html
```

#### How the Browser Game Works
- The **main thread** (`script.js`) renders the board and handles clicks.
- It spawns multiple **Web Workers** via `wasm-init.js`, each running `chess-worker.js` + `chess.wasm`.
- When you move:
  - The move is converted to a string like `e2e4` and sent to a worker (`VALIDATE_MOVE`).
  - Go validates the move with `handlers.IsValidMove`, applies it with `applyMove`, and returns a new **FEN**.
  - The JS board state is updated from that FEN (single source of truth).
- When it’s the engine’s turn:
  - The current FEN is sent to all workers (`INIT_BOARD`).
  - One worker gets all legal moves via `GET_ALL_MOVES`.
  - The move list is split into chunks and distributed to workers via `SEARCH_SUBSET` (root splitting).
  - Each worker returns its best move + score; the main thread picks the globally best move.
  - The chosen move is applied through `APPLY_MOVE` in Go, which returns the updated FEN; the JS board is synced from this FEN.

---

## ☁️ Deployment

Because the engine runs entirely in the browser via WASM and Web Workers, deployment is simple:

- Build `frontend/chess.wasm` using `build_wasm.bat`.
- Serve the `frontend/` directory with **any static hosting**:
  - GitHub Pages
  - Netlify / Vercel (static site)
  - Nginx / Apache
  - Railway / Render as a static site

No long-running Go HTTP server is required in production for the browser UI, since all engine logic executes inside the user’s browser.

---

## 🔮 Future Plans

This project is a solid foundation, and there are many exciting features that could be added next:

- [ ] **Implement the UCI Protocol:** Allow the engine to communicate with standard chess GUIs like Arena or Cute Chess to play against other engines.
- [ ] **Add an Opening Book:** Improve the engine's opening play by using a pre-computed book of moves.
- [ ] **Enhance Evaluation:** Add more advanced evaluation terms, such as:
  - Pawn structure (passed pawns, doubled pawns)
  - King safety
  - Bishop pair bonus
- [ ] **Time Management:** Proper iterative deepening based on a time budget per move (both CLI and WASM).
- [ ] **Configurable Difficulty:** Depth limits, randomization, or contempt factor to change playing style.

---

## 💡 My Journey

This project was built as a hands-on way to learn Go and the fascinating world of computer chess. It served as a practical exercise in:

- Algorithm design (Minimax, alpha–beta, quiescence, aspiration search)
- Data structures (board representations, Zobrist hashing, transposition tables)
- Performance optimization (move ordering, pruning, caching)
- Concurrency (browser Web Workers and root-splitting search)
- Bridging native Go and the web through WebAssembly

Both the CLI and browser versions now share the same core engine, so improvements in the Go search/evaluation code benefit all frontends automatically.
