import sys
import util

class Solution:
    def __init__(self):
        """
        Initialize the Solution object with necessary attributes.
        """
        self.id_of_driver = []
        self.idloadassigned = {}
        self.starting = util.Point(0, 0)  # Depot or starting point
        self.limit = 12 * 60  # Max allowable distance (in minutes)

    def fileparse(self, file_path):
        """
        Parse the problem from the file and populate the idloadassigned dictionary.
        """
        loads = util.loadProblemFromFile(file_path)
        for load in loads:
            self.idloadassigned[int(load.id)] = load

    def meancost(self):
        """
        Calculate savings for each possible link between loads and sort them in descending order.
        """
        savings = []
        for i in self.idloadassigned:
            for j in self.idloadassigned:
                if i != j:
                    load1 = self.idloadassigned[i]
                    load2 = self.idloadassigned[j]
                    key = (i, j)
                    saving = (key, util.distanceBetweenPoints(load1.dropoff, self.starting) \
                                    + util.distanceBetweenPoints(self.starting, load2.pickup) \
                                    - util.distanceBetweenPoints(load1.dropoff, load2.pickup))
                    savings.append(saving)

        savings = sorted(savings, key=lambda x: x[1], reverse=True)
        return savings

    def calculate(self, nodes):
        """
        Calculate the total distance for a given route of nodes.
        """
        if not nodes:
            return 0.0
        
        distance = 0.0
        for i in range(len(nodes)):
            distance += nodes[i].delivery_distance
            if i != (len(nodes) - 1):
                distance += util.distanceBetweenPoints(nodes[i].dropoff, nodes[i + 1].pickup)

        distance += util.distanceBetweenPoints(self.starting, nodes[0].pickup)
        distance += util.distanceBetweenPoints(nodes[-1].dropoff, self.starting)
        
        return distance

    def methodforsavings(self):
        """
        Implement the Clarke-Wright Savings Algorithm to methodforsavings the VRP.
        """
        
        savings = self.meancost()
        for link, _ in savings:
            load1 = self.idloadassigned[link[0]]
            load2 = self.idloadassigned[link[1]]

            if not load1.assigned and not load2.assigned:
                cost = self.calculate([load1, load2])
                if cost <= self.limit:
                    worker = util.Driver()
                    worker.route = [load1, load2]
                    self.id_of_driver.append(worker)
                    load1.assigned = worker
                    load2.assigned = worker

            elif load1.assigned and not load2.assigned:
                worker = load1.assigned
                i = worker.route.index(load1)
                if i == len(worker.route) - 1:
                    cost = self.calculate(worker.route + [load2])
                    if cost <= self.limit:
                        worker.route.append(load2)
                        load2.assigned = worker

            elif not load1.assigned and load2.assigned:
                worker = load2.assigned
                i = worker.route.index(load2)
                if i == 0:
                    cost = self.calculate([load1] + worker.route)
                    if cost <= self.limit:
                        worker.route = [load1] + worker.route
                        load1.assigned = worker

            else:
                worker1 = load1.assigned
                i1 = worker1.route.index(load1)

                worker2 = load2.assigned
                i2 = worker2.route.index(load2)

                if (i1 == len(worker1.route) - 1) and (i2 == 0) and (worker1 != worker2):
                    cost = self.calculate(worker1.route + worker2.route)
                    if cost <= self.limit:
                        worker1.route = worker1.route + worker2.route
                        for load in worker2.route:
                            load.assigned = worker1
                        
                        self.id_of_driver.remove(worker2)

        for load in self.idloadassigned.values():
            if not load.assigned:
                worker = util.Driver(0, [])
                worker.route.append(load)
                self.id_of_driver.append(worker)
                load.assigned = worker

    def print_solution(self):
        """
        Print the final solution, displaying the routes for each driver.
        """
        for worker in self.id_of_driver:
            print([int(load.id) for load in worker.route])

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python solution.py <file_path>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    solution = Solution()
    solution.fileparse(file_path)
    solution.methodforsavings()
    solution.print_solution()
