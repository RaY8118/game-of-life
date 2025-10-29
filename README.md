# 🎮 Conway's Game of Life in Go

A multi-implementation of Conway's Game of Life written in Go 🐹, featuring a web-based version with real-time WebSocket updates 🌐 and a terminal-based version with interactive patterns ⌨️.

## ✨ Features

- **🌐 Web Version**: Real-time simulation via WebSocket, served by a Go backend
- **💻 Terminal Version**: Interactive terminal simulation with predefined patterns
- **🐳 Docker Support**: Containerized deployment for the web version
- **📱 Responsive UI**: Web interface with controls for start/stop and grid resizing
- **🎯 Multiple Patterns**: Terminal version includes Glider, Blinker, Toad, Beacon, and random initialization

## 📁 Project Structure

```
go-gol/
├── backend/              # WebSocket server
│   ├── websocket_server.go
│   ├── go.mod
│   └── go.sum
├── frontend/             # Web client
│   └── index.html
├── terminal_version/     # Terminal client
│   ├── gol.go
│   ├── go.mod
│   └── go.sum
├── Dockerfile           # Container build
├── start.sh            # Startup script
└── README.md
```

## 🔧 Prerequisites

- Go 1.24 or later
- For terminal version: Linux/macOS (uses terminal-specific libraries)

## 🚀 Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-gol.git
   cd go-gol
   ```

2. For the web version, install backend dependencies:
   ```bash
   cd backend
   go mod tidy
   cd ..
   ```

3. For the terminal version, install dependencies:
   ```bash
   cd terminal_version
   go mod tidy
   cd ..
   ```

## ▶️ Running

### 🌐 Web Version

Start the WebSocket server:
```bash
cd backend
go run websocket_server.go
```

Open your browser to `http://localhost:8080` to view the simulation.

Alternatively, view the live deployment on Render: [Live Demo](https://game-of-life-ar9q.onrender.com)

### 💻 Terminal Version

Run the terminal simulation:
```bash
cd terminal_version
go run gol.go
```

Follow the menu prompts to select an initial pattern, then press 'q' to quit.

### 🐳 Docker

Build and run with Docker:
```bash
docker build -t go-gol .
docker run -p 8080:8080 go-gol
```

## 📖 Usage

### 🌐 Web Version
- 🔧 Use the resolution dropdown to change grid size
- ✅ Click "Apply" to resize the grid
- ▶️ Click "Start" to begin the simulation
- ⏸️ Click "Stop" to pause
- 🔄 The grid updates in real-time via WebSocket

### 💻 Terminal Version
- 🎯 Choose from 5 initial patterns: Glider, Blinker, Toad, Beacon, or Random
- ⚡ The simulation runs automatically at ~20 FPS
- 🚪 Press 'q' or ESC to exit

## 🏗️ Architecture

- **🔧 Backend**: Go server using Gorilla WebSocket for real-time communication
- **🎨 Frontend**: Vanilla HTML/CSS/JavaScript with WebSocket client
- **💻 Terminal**: Go program using keyboard input and terminal clearing for animation
- **🧬 Simulation**: Standard Conway's Game of Life rules implemented in Go

## 🤝 Contributing

1. 🍴 Fork the repository
2. 🌿 Create a feature branch
3. 🔄 Make your changes
4. 🧪 Add tests if applicable
5. 📤 Submit a pull request

## 📄 License

MIT License - see LICENSE file for details
