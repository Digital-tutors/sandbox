import datetime
import subprocess
from sandbox.services.configurer import parse_config
from sandbox.services.ResourceMonitor import ResourceMonitor
from concurrent.futures import ThreadPoolExecutor
from sandbox.controllers.TaskController import get_task_by_url
from sandbox.services.sender import Sender
import os


class Checker:
    def __init__(self, task_id: str = None, lang: str = None, file_name: str = None, user_id: str = None,
                 corr_id: str = None):
        self.__task_id = task_id
        self.__lang = lang
        self.__file_name = file_name
        self.__user_id = user_id
        self.__corr_id = corr_id
        # set url
        task = get_task_by_url("{}".format(str(self.__task_id)))
        self.__tests = task.tests
        self.__time_limit = task.options["timeLimit"]
        self.__memory_limit = task.options["memoryLimit"]

    def get_result(self, code_return: str = None, message_out: str = None, time_usage: str = "", memory_usage: str = ""):
        attempt: str = "0"
        sender = Sender(task_id=self.__task_id, corr_id=self.__corr_id, user_id=self.__user_id, code_return=code_return, message_out=message_out, time_usage=time_usage, memory_usage=memory_usage)
        sender.send_students_result()

    def test_code(self) -> str:
        config = parse_config()
        is_compilable = config["is_compilable"]
        __lang_config = config["lang_config"]
        source_file_full_name = config["source_code_path"]
        exec_file_full_name = config["code_path"]
        compiler_path = config["lang_config"]["compiler_path"]

        args = __lang_config["args_format"] \
            .replace("$source_file_full_name", source_file_full_name) \
            .replace("$exec_file_full_name", exec_file_full_name)

        if is_compilable:
            result = self.compile_file(self.compilate, compiler_path, args)
            if result[0].returncode != 0:
                mssg = "Compilation error"
                self.get_result(code_return=str(result[0].returncode), message_out=mssg, runtime=str(result[1]),
                                memory=str(result[1]))
                return result

            test_input = self.__tests['input']
            required_output = self.__tests['output']

            what_to_run = os.path.join('.', exec_file_full_name) if is_compilable else ' '.join([compiler_path, args])
            self.run_code(what_to_run=what_to_run, test_input_arr=test_input, required_output=required_output)
            self.get_result(code_return=str(result[0].returncode), message_out=result[1], runtime=str(result[2]),
                            memory=str(result[3]))
            return result[1]

    def run_code(self, what_to_run: str, test_input_arr: list, required_output: list):
        test_i = 0
        test_j = 0
        monitor = ResourceMonitor(self.__time_limit, self.__memory_limit)
        with ThreadPoolExecutor() as executor:
            if len(test_input_arr) != 0:
                for i in range(len(test_input_arr)):
                    mem_thread = executor.submit(monitor.memory_usage)
                    time_thread = executor.submit(monitor.timeout_usage)
                    fn_thread = executor.submit(self.__run_and_check, what_to_run, test_input_arr[i], monitor)
                    result = fn_thread.result()
                    monitor.keep_measuring = False
                    max_mem_usage = mem_thread.result()
                    max_time_usage = time_thread.result()
                    if result[0].returncode != 0:
                        mssg = "Runtime error, test #{}".format(str(test_i))
                        test_i = test_i + 1
                        break
                    elif result[0].stdout != required_output[i]:
                        mssg = "Wrong answer, test #{}".format(str(test_j))
                        test_j = i + 1
                        break
                    else:
                        mssg = "Correct answer"

            else:
                mem_thread = executor.submit(monitor.memory_usage)
                time_thread = executor.submit(monitor.timeout_usage)
                fn_thread = executor.submit(self.__run_and_check, what_to_run, test_input_arr[0], monitor)
                result = fn_thread.result()
                monitor.keep_measuring = False
                max_mem_usage = mem_thread.result()
                max_time_usage = time_thread.result()
                if result[0].returncode != 0:
                    mssg = "Runtime error"
                elif result[0].stdout != required_output[0]:
                    mssg = "Wrong answer"
                else:
                    mssg = "Correct answer"

        return [result, mssg, max_time_usage, max_mem_usage]

    def __run_and_check(self, what_to_run: str, test_input: str = None, monitor: ResourceMonitor = None):
        try:
            completed_run = subprocess.run(
                what_to_run,
                input=test_input,
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                timeout=self.__time_limit,
                preexec_fn=monitor.set_memory_limit)
        except subprocess.TimeoutExpired:
            monitor.keep_measuring = False
        finally:
            return completed_run

    def compile_file(self, compiler_path, args):
        with ThreadPoolExecutor() as executor:
            monitor = ResourceMonitor(self.__time_limit, self.__memory_limit)
            mem_thread = executor.submit(monitor.memory_usage)
            time_thread = executor.submit(monitor.timeout_usage)
            fn_thread = executor.submit(self.compile, compiler_path, args, monitor)
            result = fn_thread.result()
            monitor.keep_measuring = False
            max_usage = mem_thread.result()
            time_usage = time_thread.result()
            return [result, max_usage, time_usage]

    # Function that compile source code.
    # Return result of compilation
    def compile(self, compiler_path: str, path_args: str, monitor: ResourceMonitor):
        try:
            compilation = subprocess.run(' '.join([compiler_path, path_args]),
                                         shell=True,
                                         stdout=subprocess.PIPE,
                                         stderr=subprocess.STDOUT,
                                         universal_newlines=True,
                                         timeout=self.__time_limit,
                                         preexec_fn=monitor.set_memory_limit)
        except subprocess.TimeoutExpired:
            monitor.keep_measuring = False
        finally:
            return compilation
