import resource
from time import sleep
import time

TO_MEGABYTES = 1024 * 1024
TO_BYTES = 1024 * 1024


class ResourceMonitor:
    def __init__(self, timeout, memory):
        self.keep_measuring = True
        self.force_stopped = False
        self.timeout_value = int(timeout)
        self.memory = int(memory) * TO_BYTES


    # Calculate maximum memory usage by program
    # Return memory usage by program
    def memory_usage(self):
         max_usage = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
         while self.keep_measuring:
             max_usage = max(
                 max_usage,
                 resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
             )
             if max_usage > self.memory:
                 self.keep_measuring = False
                 self.force_stopped = True
                 return {"usage": max_usage / TO_MEGABYTES, "message": "Memory Expired"}
             sleep(0.01)

         return {"usage": max_usage / TO_MEGABYTES, "message": ""}

    # Timer that compute time using to run the program
    # Return time usage. Else return -1
    def timeout_usage(self):
        start_time = time.time()
        time_usage = time.time() - start_time
        while self.keep_measuring:
            time_usage = time.time() - start_time
            if time_usage > self.timeout_value:
                self.keep_measuring = False
                self.force_stopped = True
                return {"usage": time_usage, "message": "Timeout Expired"}
            time.sleep(0.01)
        return {"usage": time_usage, "message": ""}

    def set_memory_limit(self):
        try:
            resource.setrlimit(resource.RLIMIT_AS, (self.memory, resource.RLIM_INFINITY))
        except ValueError:
            pass
