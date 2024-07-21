import math
import io

class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

class item:
    def __init__(self, id, pickup, dropoff):
        self.id = id
        self.pickup = pickup
        self.dropoff = dropoff
        self.assigned = None
        # Calculate the distance between pickup and dropoff upon initialization
        self.delivery_distance = distanceBetweenPoints(pickup, dropoff)

class Driver:
    def __init__(self, distance=0.0, route=[]):
        self.distanceTravelled = distance
        self.route = route

def distanceBetweenPoints(p1, p2):
    """
    Calculate Euclidean distance between two points.
    """
    xDiff = p1.x - p2.x
    yDiff = p1.y - p2.y
    return math.sqrt(xDiff*xDiff + yDiff*yDiff)

def getPointFromPointStr(pointStr):
    """
    Convert a string representation of a point to a Point object.
    """
    pointStr = pointStr.replace("(", "").replace(")", "")
    splits = pointStr.split(",")
    return Point(float(splits[0]), float(splits[1]))

def loadProblemFromProblemStr(problemStr):
    """
    Parse problem string into a list of items.
    """
    items = []
    buf = io.StringIO(problemStr)
    gotHeader = False
    while True:
        line = buf.readline()
        if not gotHeader:
            gotHeader = True
            continue
        if len(line) == 0:
            break
        line = line.strip()  # Remove any trailing newline or spaces
        splits = line.split()
        id = splits[0]
        pickup = getPointFromPointStr(splits[1])
        dropoff = getPointFromPointStr(splits[2])
        items.append(item(id, pickup, dropoff))
    return items

def loadProblemFromFile(filePath):
    """
    Load problem from a file, returning a list of items.
    """
    with open(filePath, "r") as f:
        problemStr = f.read()
    return loadProblemFromProblemStr(problemStr)
