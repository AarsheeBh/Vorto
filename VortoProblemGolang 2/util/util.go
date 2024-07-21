// util/util.go
package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

type Point struct {
	X, Y float64
}

type Driver struct {
	DistanceTravelled float64
	Route             []*Load
}

func NewDriver() *Driver {
	return &Driver{
		DistanceTravelled: 0.0,
		Route:             []*Load{},
	}
}

type Load struct {
	ID               string
	Pickup, Dropoff  Point
	Assigned         *Driver
	DeliveryDistance float64
}

func NewLoad(id string, pickup, dropoff Point) *Load {
	return &Load{
		ID:               id,
		Pickup:           pickup,
		Dropoff:          dropoff,
		DeliveryDistance: DistanceBetweenPoints(pickup, dropoff),
	}
}

func DistanceBetweenPoints(p1, p2 Point) float64 {
	xDiff := p1.X - p2.X
	yDiff := p1.Y - p2.Y
	return math.Sqrt(xDiff*xDiff + yDiff*yDiff)
}

func LoadProblemFromFile(filePath string) ([]Load, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return LoadProblemFromProblemStr(string(content)), nil
}

func getPointFromPointStr(pointStr string) Point {
	pointStr = strings.ReplaceAll(pointStr, "(", "")
	pointStr = strings.ReplaceAll(pointStr, ")", "")
	splits := strings.Split(pointStr, ",")
	x, _ := strconv.ParseFloat(splits[0], 64)
	y, _ := strconv.ParseFloat(splits[1], 64)
	return Point{X: x, Y: y}
}

func LoadProblemFromProblemStr(problemStr string) []Load {
	var loads []Load
	scanner := bufio.NewScanner(strings.NewReader(problemStr))
	gotHeader := false
	for scanner.Scan() {
		if !gotHeader {
			gotHeader = true
			continue
		}
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			break
		}
		splits := strings.Fields(line)
		id := splits[0]
		pickup := getPointFromPointStr(splits[1])
		dropoff := getPointFromPointStr(splits[2])
		loads = append(loads, *NewLoad(id, pickup, dropoff))
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading problem string:", err)
	}
	return loads
}

type SavingKey struct {
	I, J string
}

type Saving struct {
	Key    SavingKey
	Amount float64
}
