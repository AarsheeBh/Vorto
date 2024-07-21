package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Point represents a Cartesian coordinate
type Point struct {
	X float64
	Y float64
}

// Load represents a load with a pickup and dropoff point
type Load struct {
	ID      string
	Pickup  Point
	Dropoff Point
}

// Driver represents a driver with a route of loads
type Driver struct {
	Route []*Load
}

// NewDriver creates a new driver
func NewDriver() *Driver {
	return &Driver{
		Route: []*Load{},
	}
}

// DistanceBetweenPoints calculates the Euclidean distance between two points
func DistanceBetweenPoints(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p2.X-p1.X, 2) + math.Pow(p2.Y-p1.Y, 2))
}

// LoadProblemFromFile loads loads from a given file
func LoadProblemFromFile(filePath string) ([]Load, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var loads []Load
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		id := parts[0]
		pickupCoords := strings.Trim(parts[1], "()")
		dropoffCoords := strings.Trim(parts[2], "()")
		pickupParts := strings.Split(pickupCoords, ",")
		dropoffParts := strings.Split(dropoffCoords, ",")
		if len(pickupParts) < 2 || len(dropoffParts) < 2 {
			continue
		}
		pickup := Point{X: toFloat(pickupParts[0]), Y: toFloat(pickupParts[1])}
		dropoff := Point{X: toFloat(dropoffParts[0]), Y: toFloat(dropoffParts[1])}
		load := Load{ID: id, Pickup: pickup, Dropoff: dropoff}
		loads = append(loads, load)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return loads, nil
}

// toFloat converts a string to a float64
func toFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Solution represents the VRP solution
type Solution struct {
	drivers     []*Driver
	loadByID    map[string]*Load
	depot       Point
	maxDistance float64
}

// NewSolution creates a new solution
func NewSolution() *Solution {
	return &Solution{
		drivers:     []*Driver{},
		loadByID:    make(map[string]*Load),
		depot:       Point{X: 0, Y: 0},
		maxDistance: 12 * 60, // 12 hours in minutes
	}
}

// computeDistance computes the total distance for a route
func (s *Solution) computeDistance(nodes []*Load) float64 {
	if len(nodes) == 0 {
		return 0.0
	}

	distance := 0.0
	for i := range nodes {
		if i == 0 {
			distance += DistanceBetweenPoints(s.depot, nodes[i].Pickup) // to first pickup
		} else {
			distance += DistanceBetweenPoints(nodes[i-1].Dropoff, nodes[i].Pickup) // to next pickup
		}
		distance += DistanceBetweenPoints(nodes[i].Pickup, nodes[i].Dropoff) // to dropoff
	}
	distance += DistanceBetweenPoints(nodes[len(nodes)-1].Dropoff, s.depot) // return to depot

	return distance
}

// calculateCost calculates the total cost of the solution
func (s *Solution) calculateCost() float64 {
	totalDistance := 0.0
	for _, driver := range s.drivers {
		totalDistance += s.computeDistance(driver.Route)
	}
	return 500*float64(len(s.drivers)) + totalDistance
}

// loadProblems loads problems from the given directory
func (s *Solution) loadProblems(filePath string) error {
	loads, err := LoadProblemFromFile(filePath)
	if err != nil {
		return err
	}

	for _, load := range loads {
		s.loadByID[load.ID] = &load
	}
	return nil
}

// initialSolution creates an initial solution using a more refined greedy approach
func (s *Solution) initialSolution() {
	// Use Clarke-Wright Savings Algorithm to create initial solution
	s.clarkeWrightSavings()
}

// clarkeWrightSavings implements the Clarke-Wright Savings Algorithm
func (s *Solution) clarkeWrightSavings() {
	// Step 1: Initialize savings list
	var savings []struct {
		From   *Load
		To     *Load
		Saving float64
	}

	for _, load1 := range s.loadByID {
		for _, load2 := range s.loadByID {
			if load1.ID == load2.ID {
				continue
			}
			saving := DistanceBetweenPoints(s.depot, load1.Pickup) + DistanceBetweenPoints(load2.Dropoff, s.depot) - DistanceBetweenPoints(load1.Pickup, load2.Dropoff)
			savings = append(savings, struct {
				From   *Load
				To     *Load
				Saving float64
			}{load1, load2, saving})
		}
	}

	// Step 2: Sort savings list in descending order
	sort.Slice(savings, func(i, j int) bool {
		return savings[i].Saving > savings[j].Saving
	})

	// Step 3: Merge routes based on savings
	for _, saving := range savings {
		fromDriver := s.findDriverForLoad(saving.From)
		toDriver := s.findDriverForLoad(saving.To)

		if fromDriver == nil && toDriver == nil {
			newDriver := NewDriver()
			newDriver.Route = append(newDriver.Route, saving.From, saving.To)
			s.drivers = append(s.drivers, newDriver)
		} else if fromDriver != nil && toDriver == nil {
			if s.computeDistance(append(fromDriver.Route, saving.To)) <= s.maxDistance {
				fromDriver.Route = append(fromDriver.Route, saving.To)
			}
		} else if fromDriver == nil && toDriver != nil {
			if s.computeDistance(append([]*Load{saving.From}, toDriver.Route...)) <= s.maxDistance {
				toDriver.Route = append([]*Load{saving.From}, toDriver.Route...)
			}
		} else if fromDriver != toDriver {
			if s.computeDistance(append(fromDriver.Route, toDriver.Route...)) <= s.maxDistance {
				fromDriver.Route = append(fromDriver.Route, toDriver.Route...)
				s.removeDriver(toDriver)
			}
		}
	}

	s.reassignLoads()
}

// localSearch performs local search to improve the solution
func (s *Solution) localSearch() {
	for improvement := true; improvement; {
		improvement = false
		for i := 0; i < len(s.drivers); i++ {
			for j := i + 1; j < len(s.drivers); j++ {
				if s.swapLoadsBetweenDrivers(s.drivers[i], s.drivers[j]) {
					improvement = true
				}
			}
		}
	}
}

// swapLoadsBetweenDrivers tries to swap loads between two drivers to reduce cost
func (s *Solution) swapLoadsBetweenDrivers(driver1, driver2 *Driver) bool {
	bestCost := s.calculateCost()
	bestSwap := false

	for i := 0; i < len(driver1.Route); i++ {
		for j := 0; j < len(driver2.Route); j++ {
			driver1.Route[i], driver2.Route[j] = driver2.Route[j], driver1.Route[i]
			newCost := s.calculateCost()
			if newCost < bestCost {
				bestCost = newCost
				bestSwap = true
			} else {
				driver1.Route[i], driver2.Route[j] = driver2.Route[j], driver1.Route[i]
			}
		}
	}

	return bestSwap
}

// reassignLoads reassigns unassigned loads
func (s *Solution) reassignLoads() {
	var unassigned []*Load
	for _, load := range s.loadByID {
		if s.findDriverForLoad(load) == nil {
			unassigned = append(unassigned, load)
		}
	}

	for _, load := range unassigned {
		bestDriver := -1
		bestCost := math.MaxFloat64

		for i, driver := range s.drivers {
			newRoute := append(driver.Route, load)
			if cost := s.computeDistance(newRoute); cost <= s.maxDistance && cost < bestCost {
				bestCost = cost
				bestDriver = i
			}
		}

		if bestDriver == -1 {
			newDriver := NewDriver()
			newDriver.Route = append(newDriver.Route, load)
			s.drivers = append(s.drivers, newDriver)
		} else {
			s.drivers[bestDriver].Route = append(s.drivers[bestDriver].Route, load)
		}
	}
}

// findDriverForLoad finds the driver for a specific load
func (s *Solution) findDriverForLoad(load *Load) *Driver {
	for _, driver := range s.drivers {
		for _, l := range driver.Route {
			if l.ID == load.ID {
				return driver
			}
		}
	}
	return nil
}

// removeDriver removes a driver from the solution
func (s *Solution) removeDriver(driver *Driver) {
	for i, d := range s.drivers {
		if d == driver {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			return
		}
	}
}

// copy creates a deep copy of the solution
func (s *Solution) copy() *Solution {
	newSolution := NewSolution()
	newSolution.depot = s.depot
	newSolution.maxDistance = s.maxDistance
	for _, driver := range s.drivers {
		newDriver := NewDriver()
		for _, load := range driver.Route {
			newDriver.Route = append(newDriver.Route, load)
		}
		newSolution.drivers = append(newSolution.drivers, newDriver)
	}
	for id, load := range s.loadByID {
		newSolution.loadByID[id] = load
	}
	return newSolution
}

// processFile processes a single file and returns cost and duration
func processFile(filePath string) (float64, time.Duration, error) {
	start := time.Now()
	solution := NewSolution()
	err := solution.loadProblems(filePath) // Load problems from file
	if err != nil {
		return 0, 0, err
	}

	solution.initialSolution()
	solution.localSearch()

	duration := time.Since(start)
	cost := solution.calculateCost()
	return cost, duration, nil
}

// processDirectory processes all files in a directory
func processDirectory(dirPath string) {
	var wg sync.WaitGroup
	var totalCost float64
	var totalDuration time.Duration
	var fileCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			go func(filePath string) {
				defer wg.Done()
				cost, duration, err := processFile(filePath)
				if err != nil {
					fmt.Printf("Error processing file %s: %v\n", filePath, err)
					return
				}
				totalCost += cost
				totalDuration += duration
				fileCount++
			}(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", dirPath, err)
		return
	}

	wg.Wait()

	if fileCount > 0 {
		fmt.Printf("Mean Cost: %.2f\n", totalCost/float64(fileCount))
		fmt.Printf("Mean Running Time: %v\n", totalDuration/time.Duration(fileCount))
	} else {
		fmt.Println("No files found in directory.")
	}
}

// main function
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go {path_to_directory}")
		return
	}

	dirPath := os.Args[1]
	processDirectory(dirPath)
}
