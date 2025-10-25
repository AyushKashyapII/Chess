# Go Chess Engine

## 📖 About The Project

This project is a classic chess engine built from the ground up in the **Go** programming language. It started as a personal learning journey to master Go's fundamentals and has evolved from a simple Fyne-based desktop application into a fully-fledged web application.

The engine features a robust AI powered by a Minimax search algorithm with several advanced optimizations. The frontend is a clean, modern web interface built with standard HTML, CSS, and JavaScript, designed to communicate with the Go backend via a simple API.

The primary goal of this project was to learn and implement the core concepts of chess engine development, including move generation, board evaluation, search algorithms, and deployment strategies for web services.

---

## ✨ Features

The engine includes several key features that are standard in modern chess AI:

### AI Engine
*   **Search Algorithm:** A **Minimax** search core that explores the game tree to find the optimal move.
*   **Alpha-Beta Pruning:** A critical optimization that dramatically reduces the search space, allowing the engine to think several moves deeper in the same amount of time.
*   **Quiescence Search:** A specialized search that runs at the "leaves" of the main search tree to resolve tactical situations (like capture sequences), preventing the "horizon effect" and improving tactical accuracy.
*   **Transposition Tables:** The engine uses **Zobrist Hashing** to implement a transposition table, which acts as a memory to store the results of previously analyzed positions. This avoids re-calculating the same position and provides a massive performance boost.

### Board Evaluation
*   **Material Advantage:** The core evaluation is based on the standard point values of the pieces.
*   **Positional Awareness:** The engine uses **Piece-Square Tables (PSTs)** to understand that the value of a piece depends on its position on the board. For example, a knight in the center is more valuable than a knight in the corner.

### Web Application
*   **Go Backend:** A lightweight backend server built with Go's native `net/http` package.
*   **RESTful API:** A simple `/get_move` endpoint that accepts a board state (in FEN format) and returns the AI's best move in JSON format.
*   **Interactive Frontend:** A clean user interface built with HTML, CSS, and vanilla JavaScript that allows a human player to play against the engine in their web browser.

---

## 🛠️ Tech Stack

*   **Backend:** [Go](https://golang.org/) (`net/http`)
*   **Frontend:** HTML5, CSS3, JavaScript
*   **Deployment:** Designed for platforms like [Railway](https://railway.app/) or any service that supports Go applications or Docker.

---

## 🚀 How to Run Locally

To run this project on your local machine, follow these steps.

### Prerequisites
*   You must have Go installed (version 1.18 or later).
*   A modern web browser.

### Installation & Execution
1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/your-repo-name.git
    ```
2.  **Navigate to the project directory:**
    ```bash
    cd your-repo-name
    ```
3.  **Build the application:**
    ```bash
    go build
    ```
    This will create an executable file (`your-repo-name.exe` on Windows, or `your-repo-name` on macOS/Linux).

4.  **Run the server:**
    *   On Windows:
        ```powershell
        .\your-repo-name.exe
        ```
    *   On macOS / Linux:
        ```bash
        ./your-repo-name
        ```
    You should see a message in your terminal: `Starting chess server on http://localhost:8080`.

5.  **Play the game!**
    Open your web browser and navigate to:
    [http://localhost:8080](http://localhost:8080)

---

## ☁️ Deployment

This application is ready to be deployed on a platform like Railway. To do so, you will need a `Procfile` in the root of your repository with the following content:

```
web: ./your-repo-name
```
Railway will automatically detect the Go project, build it using `go build`, and run the executable as defined in the `Procfile`.

---

## 🔮 Future Plans

This project is a solid foundation, and there are many exciting features that could be added next:

*   [ ] **Implement the UCI Protocol:** Allow the engine to communicate with standard chess GUIs like Arena or Cute Chess to play against other engines.
*   [ ] **Add an Opening Book:** Improve the engine's opening play by using a pre-computed book of moves.
*   [ ] **Enhance Evaluation:** Add more advanced evaluation terms, such as:
    *   Pawn structure (passed pawns, doubled pawns)
    *   King safety
    *   Bishop pair bonus
*   [ ] **Validate Human Moves:** Add a backend endpoint to validate that the human player's move is legal before accepting it.

---

## 💡 My Journey

This project was built as a hands-on way to learn Go and the fascinating world of computer chess. It served as a practical exercise in algorithm design, data structures, performance optimization, and the transition from desktop to web-based application architecture.
