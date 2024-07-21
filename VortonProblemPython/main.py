import subprocess
import os
import sys
import io
import math
import time
import argparse

# Data class to represent a point in 2D space
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    
    def toString(self):
        return f"({self.x},{self.y})"

# Utility function to calculate Euclidean distance between two points
def euclideandistance(p1, p2):
    dx = p1.x - p2.x
    dy = p1.y - p2.y
    return math.sqrt(dx * dx + dy * dy)

# Data class to represent a item with pickup and dropoff points
class item:
    def __init__(self, id, pickup, dropoff):
        self.id = id
        self.pickup = pickup
        self.dropoff = dropoff

# Class to represent the Vehicle Routing Problem (VRP)
class VRP:
    def __init__(self, items):
        self.items = items
    
    def toProblemString(self):
        """
        Convert the VRP problem to a string format.
        """
        sum = "itemNumber pickup dropoff\n"
        for idx, item in enumerate(self.items):
            sum += f"{idx+1} {item.pickup.toString()} {item.dropoff.toString()}\n"
        return sum

# Utility function to read the problem from a file and return a VRP object
def fileparsermethod(filePath):
    with open(filePath, "r") as f:
        problemStr = f.read()
    return input(problemStr)

# Utility function to parse a string representation of a point
def getstringinput(pointStr):
    pointStr = pointStr.strip("()")
    parts = pointStr.split(",")
    return Point(float(parts[0]), float(parts[1]))

# Utility function to item a problem from its string representation
def input(problemStr):
    items = []
    buf = io.StringIO(problemStr)
    _ = buf.readline()  # Skip header
    for input in buf:
        input = input.strip()
        if not input:
            break
        parts = input.split()
        id = parts[0]
        pickup = getstringinput(parts[1])
        dropoff = getstringinput(parts[2])
        items.append(item(id, pickup, dropoff))
    return VRP(items)

# Utility function to parse the solution from its string representation
def loadSolutionFromString(solutionStr):
    timings = []
    buf = io.StringIO(solutionStr)
    for input in buf:
        input = input.strip()
        if '[' not in input or ']' not in input:
            return timings, f"Solution format incorrect. Expected all inputs to be in format [{item_id}, {item_id}, ...], but got this: {input}"
        input = input.strip('[]').replace(' ', '')
        parts = input.split(',')
        timing = [itemID for itemID in parts]
        timings.append(timing)
    return timings, ""

# Utility function to check for errors in the solution and count assignments
def amount(problem, solutiontimings):
    personId = set()
    for timing in solutiontimings:
        for itemID in timing:
            if itemID in personId:
                return f"item {itemID} was included in at least two driver timings"
            personId.add(itemID)

    if len(personId) != len(problem.items):
        return "the solution item count is not equal to the problem item count"
        
    for item in problem.items:
        if item.id not in personId:
            return f"item {item.id} was not assigned to a driver"
    
    return ""

# Utility function to calculate the distance of a timing with return start
def completeTask(timing, itemByID):
    distance = 0.0
    start = Point(0, 0)
    currentLoc = start
    for itemID in timing:
        item = itemByID[itemID]
        # To pickup
        distance += euclideandistance(currentLoc, item.pickup)
        currentLoc = item.pickup
        # To dropoff
        distance += euclideandistance(currentLoc, item.dropoff)
        currentLoc = item.dropoff
    # Return start
    distance += euclideandistance(currentLoc, start)
    return distance

# Utility function to calculate the total cost of the solution and check for errors
def getSolutionCostWithError(problem, solutiontimings):
    err = amount(problem, solutiontimings)
    if err:
        return 0, err
    return getSolutionCost(problem, solutiontimings)

# Utility function to calculate the total cost of the solution
def getSolutionCost(problem, solutiontimings):
    itemByID = {item.id: item for item in problem.items}
    timeofdriving = 0.0
    for idx, timing in enumerate(solutiontimings):
        timingMinutes = completeTask(timing, itemByID)
        if timingMinutes > 12 * 60:
            return 0, f"timing idx {idx} is invalid: driver runs for {timingMinutes} minutes"
        timeofdriving += timingMinutes
    
    return 500 * len(solutiontimings) + timeofdriving, ""

# Utility function to print the expected solution format for the user
def givenoutput():
    print("Program should only print a solution (no debugging messages) in format that looks like this:")
    print("[1,4,9,7]")
    print("[5,2,3,8]")
    print("[10,6]")

if __name__ == '__main__':
    # Parse command-line arguments
    parser = argparse.ArgumentParser()
    parser.add_argument("--problemDir", help="Path to folder containing problems")
    parser.add_argument("--cmd", help="Command to run your program (not including a problem file)")
    args = parser.parse_args()

    # Process each file in the problem directory
    files = [f for f in os.listdir(args.problemDir) if not f.startswith(".")]
    costs = []
    total = 0.0
    
    for inputFile in files:
        print(inputFile)
        print("\trunning...")
        inputPath = os.path.join(args.problemDir, inputFile)
        
        # Run the command on the input file
        cmd = args.cmd.split()
        cmd.append(inputPath)
        startTime = time.time()
        output = subprocess.check_output(cmd).decode("utf-8").replace('\r\n', '\n')
        runTime = time.time() - startTime
        print("\trun time:", runTime, "s")
        if runTime > 30:
            print("\t\tRun time constraint of 30s exceeded! Please reduce program runtime!")
        total += runTime

        print("\tevaluating solution...")
        problem = fileparsermethod(inputPath)
        timings, err = loadSolutionFromString(output)
        if err:
            print(err)
            givenoutput()
            print("Observed:")
            print(output)
            exit()
        
        cost, err = getSolutionCostWithError(problem, timings)
        if err:
            print(err)
            exit()
        costs.append(cost)
        print("\tcost:", cost)

    # Calculate and print mean cost and average runtime
    meanCost = sum(costs) / len(costs)
    meanRunTime = (total * 1000) / len(costs)
    print("mean cost:", meanCost)
    print("mean run time:", meanRunTime, "ms")
