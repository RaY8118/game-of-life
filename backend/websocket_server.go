// server.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	Type    string `json:"type"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Speed   int    `json:"speed,omitempty"`
}

type GridMessage struct {
	Grid       string `json:"grid"`
	Generation int    `json:"generation"`
	Population int    `json:"population"`
}

var (
	// initial default size (will be changed by resize requests)
	width, height = 170, 42

	grid        [][]int
	clients     = make(map[*websocket.Conn]bool)
	clientsLock sync.Mutex

	upgrader = websocket.Upgrader{
		// allow all origins for local/dev. In production you should tighten this.
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// simulation control
	running    = false
	ticker     *time.Ticker
	tickerStop chan bool
	speed      = 100 // milliseconds
	generation = 0

	// small limits to avoid runaway sizes
	maxWidth  = 1000
	maxHeight = 400
)

func main() {
	// create initial grid
	initGrid()

	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleConnections handles new websocket clients and their commands.
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	defer ws.Close()

	// register client
	clientsLock.Lock()
	clients[ws] = true
	clientsLock.Unlock()

	// send initial state immediately
	if err := ws.WriteMessage(websocket.TextMessage, []byte(getGridMessage())); err != nil {
		log.Println("write initial error:", err)
		clientsLock.Lock()
		delete(clients, ws)
		clientsLock.Unlock()
		return
	}

	// read loop for this client
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			// normal close or error: remove client and exit
			log.Println("Read error (client disconnected?):", err)
			clientsLock.Lock()
			delete(clients, ws)
			clientsLock.Unlock()
			return
		}

		// Try parse JSON first
		var cm ClientMessage
		if messageType == websocket.TextMessage && json.Unmarshal(p, &cm) == nil && cm.Type != "" {
			switch cm.Type {
			case "resize":
				// validate and apply resize
				wReq := cm.Width
				hReq := cm.Height
				if wReq <= 0 || hReq <= 0 {
					// ignore invalid
					break
				}
				if wReq > maxWidth {
					wReq = maxWidth
				}
				if hReq > maxHeight {
					hReq = maxHeight
				}
				applyResize(wReq, hReq)
			case "pattern":
				loadPattern(cm.Pattern)
				broadcast(getGridMessage())
			case "speed":
				if cm.Speed > 0 && cm.Speed <= 1000 {
					speed = cm.Speed
					if running {
						stopSimulation()
						startSimulation()
					}
				}
			default:
				log.Println("Unknown JSON message type:", cm.Type)
			}
			continue
		}

		// fallback: treat plain text commands
		if messageType == websocket.TextMessage {
			msg := string(p)
			switch msg {
			case "start":
				startSimulation()
			case "stop":
				stopSimulation()
			case "reset":
				// reset to random
				resetSimulation("random")
			default:
				log.Println("Unknown text message:", msg)
			}
		}
	}
}

// applyResize stops sim (if running), resizes grid, broadcasts, restarts if needed.
func applyResize(wReq, hReq int) {
	wasRunning := running
	stopSimulation()
	// apply new size
	width = wReq
	height = hReq
	log.Printf("Resizing grid to %dx%d\n", width, height)
	initGrid()
	broadcast(getGridMessage())
	if wasRunning {
		startSimulation()
	}
}

// runSimulation is the ticker loop that updates and broadcasts.
func runSimulation() {
	ticker = time.NewTicker(time.Duration(speed) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updateGrid()
			broadcast(getGridMessage())
		case <-tickerStop:
			return
		}
	}
}

func startSimulation() {
	if running {
		return
	}
	log.Println("Starting simulation...")
	running = true
	tickerStop = make(chan bool)
	go runSimulation()
}

func stopSimulation() {
	if !running {
		return
	}
	log.Println("Stopping simulation...")
	running = false
	close(tickerStop)
}

// resetSimulation resets the grid with the named pattern and optionally restarts.
func resetSimulation(pattern string) {
	stopSimulation()
	initGrid()
	broadcast(getGridAsString())
}

// broadcast sends msg to all connected clients (removes broken connections).
func broadcast(msg string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error (removing client):", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

// getPopulation returns the number of live cells.
func getPopulation() int {
	count := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			count += grid[y][x]
		}
	}
	return count
}

// getGridAsString returns grid as a multi-line string of '█' and spaces.
func getGridAsString() string {
	var out string
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] == 1 {
				out += "█"
			} else {
				out += " "
			}
		}
		out += "\n"
	}
	return out
}

// getGridMessage returns the grid data as JSON.
func getGridMessage() string {
	msg := GridMessage{
		Grid:       getGridAsString(),
		Generation: generation,
		Population: getPopulation(),
	}
	data, _ := json.Marshal(msg)
	return string(data)
}

// updateGrid applies the Game of Life rules into a new grid.
func updateGrid() {
	newGrid := make([][]int, height)
	for y := range newGrid {
		newGrid[y] = make([]int, width)
		for x := 0; x < width; x++ {
			neighbors := countNeighbours(x, y)
			if grid[y][x] == 1 && (neighbors == 2 || neighbors == 3) {
				newGrid[y][x] = 1
			} else if grid[y][x] == 0 && neighbors == 3 {
				newGrid[y][x] = 1
			}
		}
	}
	grid = newGrid
	generation++
}

func countNeighbours(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				count += grid[ny][nx]
			}
		}
	}
	return count
}

// initGrid creates grid with random centered.
func initGrid() {
	grid = make([][]int, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if rand.Float64() < 0.2 {
				grid[y][x] = 1
			}
		}
	}
	generation = 0
}

// clearGrid resets all cells to 0.
func clearGrid() {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid[y][x] = 0
		}
	}
}

// Pattern loading functions
func loadGlider(x, y int) {
	grid[y][x+1] = 1
	grid[y+1][x+2] = 1
	grid[y+2][x] = 1
	grid[y+2][x+1] = 1
	grid[y+2][x+2] = 1
}

func loadBlinker(x, y int) {
	grid[y][x] = 1
	grid[y][x+1] = 1
	grid[y][x+2] = 1
}

func loadToad(x, y int) {
	grid[y][x+1] = 1
	grid[y][x+2] = 1
	grid[y][x+3] = 1
	grid[y+1][x] = 1
	grid[y+1][x+1] = 1
	grid[y+1][x+2] = 1
}

func loadBeacon(x, y int) {
	grid[y][x] = 1
	grid[y][x+1] = 1
	grid[y+1][x] = 1
	grid[y+1][x+1] = 1
	grid[y+2][x+2] = 1
	grid[y+2][x+3] = 1
	grid[y+3][x+2] = 1
	grid[y+3][x+3] = 1
}

// loadPattern loads the specified pattern at the center.
func loadPattern(pattern string) {
	clearGrid()
	generation = 0
	cx := width / 2
	cy := height / 2
	switch pattern {
	case "glider":
		loadGlider(cx-1, cy-1)
	case "blinker":
		loadBlinker(cx-1, cy)
	case "toad":
		loadToad(cx-2, cy-1)
	case "beacon":
		loadBeacon(cx-2, cy-2)
	case "random":
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if rand.Float64() < 0.2 {
					grid[y][x] = 1
				}
			}
		}
	}
}
