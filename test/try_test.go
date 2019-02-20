package test

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

type point struct {
	x int
	y int
}

func (p *point) add(target *point) *point {
	return &point{p.x + target.x, p.y + target.y}
}

func (p *point) move(target *point) *point {
	p.x = p.x + target.x
	p.y = p.y + target.y
	return p
}

func (p *point) equal(target *point) bool {
	return p.x == target.x && p.y == target.y
}

func (p *point) at(grid [][]int) (int, bool) {

	if p.x < 0 || p.y < 0 {
		return -1, false
	}

	if p.y >= len(grid) {
		return -1, false
	}

	if p.x >= len(grid[p.y]) {
		return -1, false
	}

	return grid[p.y][p.x], true
}

func readmaze(filename string) [][]int {

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var smaze = string(bytes[:])
	var srows = strings.Split(smaze, "\n")
	var irow = len(srows)
	var maze = make([][]int, irow)
	for r := 0; r < irow; r++ {
		var srow = srows[r]
		var scol = strings.Split(srow, " ")
		var icol = len(scol)
		maze[r] = make([]int, icol)
		for c := 0; c < icol; c++ {
			var val = scol[c]
			val = strings.Replace(val, "\r", "", -1)
			val = strings.Replace(val, "\n", "", -1)
			maze[r][c], err = strconv.Atoi(val)
			if err != nil {
				maze[r][c] = 1
			}
		}
	}
	return maze
}

func labyrinth_walk(maze [][]int, start point, end point) ([][]int, bool) {

	defer func() {
		var pan = recover()
		if pan != nil {
			panic(pan)
		}
	}()

	var directions = [4]point{
		{-1, 0},
		{0, -1},
		{1, 0},
		{0, 1},
	}

	var irow = len(maze)
	var steps = make([][]int, irow)
	for r := range steps {
		var icol = len(maze[r])
		steps[r] = make([]int, icol)
	}

	var current *point
	var queue = []*point{&start}
	for len(queue) > 0 {
		current = queue[0]

		if current.equal(&end) {
			break
		}

		queue = queue[1:]

		for _, direction := range directions {
			var next = current.add(&direction)
			val, valid := next.at(maze)

			if val == 1 || !valid {
				continue
			}
			val, valid = next.at(steps)
			if val != 0 || !valid {
				continue
			}
			if next.equal(&start) {
				continue
			}
			curval, _ := current.at(steps)
			steps[next.y][next.x] = curval + 1

			queue = append(queue, next)
		}
	}
	var found = current.equal(&end)
	return steps, found
}

func printgrid(grid [][]int) {
	fmt.Println()
	for _, row := range grid {
		for _, val := range row {
			fmt.Printf("%4d", val)
		}
		fmt.Println()
	}
}

func labyrinth() {
	maze := readmaze("res/maze.txt")
	printgrid(maze)

	var steps, found = labyrinth_walk(maze,
		point{0, 0},
		point{len(maze[0]) - 1, len(maze) - 1},
	)
	printgrid(steps)
	fmt.Println("found:", found)
}

/* ===================================================== */

func crawler_simple() {

}

/* ===================================================== */

func TestTry(t *testing.T) {
	//labyrinth()
	crawler_simple()
}

func BenchmarkTry(b *testing.B) {

}
