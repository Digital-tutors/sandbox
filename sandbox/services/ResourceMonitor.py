import resource
from time import sleep
import time

TO_MEGABYTES = 1024
TO_BYTES = 1024 * 1024

class ResourceMonitor:
    def __init__(self, timeout, memory):
        self.keep_measuring = True
        self.timeout_value = timeout
        self.memory = memory * TO_BYTES

    # Calculate maximum memory usage by program
    # Return memory usage by program
    def memory_usage(self):
        max_usage = 0
        while self.keep_measuring:
            max_usage = max(
                max_usage,
                resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
            )
            if max_usage > self.memory:
                self.keep_measuring = False
                return max_usage
            sleep(0.000001)

        return max_usage / TO_MEGABYTES

    # Timer that compute time using to run the program
    # Return time usage. Else return -1
    def timeout_usage(self):
        max_time = 0
        start_time = time.time()
        while self.keep_measuring:
            max_time = time.time() - start_time
            if max_time > self.timeout_value:
                self.keep_measuring = False
                return -1
            sleep(0.000001)
        return max_time

    # Set virtual memory limit in bytes
    # If it impossible, will give ValueError
    def set_memory_limit(self):
        resource.setrlimit(resource.RLIMIT_AS, (self.memory, resource.RLIM_INFINITY))
