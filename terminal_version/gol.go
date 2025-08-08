package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"golang.org/x/term"
)

var grid [][]int
var width, height int

func clearScreen() {
	fmt.Print("\033[H")
}

func printGrid() {
	clearScreen()
	var output string
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] == 1 {
				output += "█"
			} else {
				output += " "
			}
		}
		output += "\n"
	}
	fmt.Print(output)
}

func countNeightbours(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dy
			ny := y + dx
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				count += grid[ny][nx]
			}
		}
	}
	return count
}

func updateGrid() {
	newGrid := make([][]int, height)
	for y := 0; y < height; y++ {
		newGrid[y] = make([]int, width)
		for x := 0; x < width; x++ {
			neighbors := countNeightbours(x, y)
			if grid[y][x] == 1 && (neighbors == 2 || neighbors == 3) {
				newGrid[y][x] = 1
			} else if grid[y][x] == 0 && neighbors == 3 {
				newGrid[y][x] = 1
			}
		}
	}
	grid = newGrid
}

func initGrid() {

	grid = make([][]int, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]int, width)
	}

	choice := menu()

	switch choice {
	case 1:
		addGlider(width/2, height/2)
	case 2:
		addBlinker(width/2, height/2)
	case 3:
		addToad(width/2, height/2)
	case 4:
		addBeacon(width/2, height/2)
	case 5:
		rand.Seed(time.Now().UnixNano())
		for y := 0; y < height; y++ {
			grid[y] = make([]int, width)
			for x := 0; x < width; x++ {
				if rand.Float64() < 0.2 {
					grid[y][x] = 1
				}
			}
		}
	}
}

func addGlider(x, y int) {
	grid[y][x+1] = 1
	grid[y+1][x+2] = 1
	grid[y+2][x] = 1
	grid[y+2][x+1] = 1
	grid[y+2][x+2] = 1
}

func addBlinker(x, y int) {
	grid[y][x] = 1
	grid[y][x+1] = 1
	grid[y][x+2] = 1
}

func addToad(x, y int) {
	grid[y][x+1] = 1
	grid[y][x+2] = 1
	grid[y][x+3] = 1
	grid[y+1][x] = 1
	grid[y+1][x+1] = 1
	grid[y+1][x+2] = 1
}

func addBeacon(x, y int) {
	grid[y][x] = 1
	grid[y][x+1] = 1
	grid[y+1][x] = 1
	grid[y+1][x+1] = 1

	grid[y+2][x+2] = 1
	grid[y+2][x+3] = 1
	grid[y+3][x+2] = 1
	grid[y+3][x+3] = 1
}

func menu() int {
	fmt.Println("Choose a pattern to initialize: ")
	fmt.Println("1. Glider")
	fmt.Println("2. Blinker")
	fmt.Println("3. Toad")
	fmt.Println("4. Beacon")
	fmt.Println("5. Random")
	fmt.Print("Enter choice (1-5): ")

	var choice int
	fmt.Scan(&choice)
	return choice
}

func main() {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width = 80
		height = 24
	} else {
		width = w
		height = h - 2
	}

	initGrid()

	err = keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	keyChan := make(chan rune)

	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				continue
			}
			if key == keyboard.KeyEsc {
				keyChan <- 'q'
				return
			}
			keyChan <- char
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			printGrid()
			updateGrid()

		case key := <-keyChan:
			if key == 'q' {
				fmt.Println("\nSimulation exited gracefully.")
				return
			}
		}
	}
}
