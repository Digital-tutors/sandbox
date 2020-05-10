import datetime
import subprocess
from services.configurer import parse_config
from services.ResourceMonitor import ResourceMonitor
from concurrent.futures import ThreadPoolExecutor
from controllers.TaskController import TaskController
from services.sender import Sender
import os

class Checker:
    def __init__(self, task_id: str = None, lang: str = None, file_name: str = None, user_id: str = None,
                 corr_id: str = None, solution_id: str = None):
        self.__task_id = task_id
        self.__lang = lang
        self.__file_name = file_name
        self.__user_id = user_id
        self.__corr_id = corr_id
        self.__solution_id = solution_id
        taskObj = TaskController()
        #self.task = taskObj.get_task_by_url("http://172.17.0.1:3000/task/{taskId}/admin/".format(taskId=self.__task_id))
        self.task = taskObj.get_task_by_url("http://172.17.0.1:3000")
        self.__tests = self.task.tests
        self.__time_limit = self.task.options["timeLimit"]
        self.__memory_limit = self.task.options["memoryLimit"]
        self.test_code()

    def get_result(self, code_return: str = None, message_out: str = None, time_usage: str = "",
                   memory_usage: str = ""):
        sender = Sender(task_id=self.__task_id, corr_id=self.__corr_id, solution_id=self.__solution_id, user_id=self.__user_id, code_return=code_return,
                        message_out=message_out, time_usage=time_usage, memory_usage=memory_usage)
        sender.send_students_result()

    def test_code(self) -> str:
        config = parse_config(file_name=self.__file_name, lang=self.__lang)
        __lang_config = config["lang_config"]

        is_compilable = __lang_config["is_compilable"]
        is_need_compile = __lang_config["is_need_compile"]
        source_file_full_name = config["source_code_path"]
        exec_file_full_name = config["executable_code_path"]
        file_full_name = config["code_path"]

        if is_compilable and is_need_compile:
            compiler_args = __lang_config["compiler"]["compiler_args"] \
                .replace("$source_file_full_name", source_file_full_name) \
                .replace("$exec_file_full_name", exec_file_full_name) \
                .replace("$file_full_name", file_full_name)
            compiler_path = __lang_config["compiler"]["path"]

            result = self.compile_file(compiler_path, compiler_args)

            if result[0].returncode != 0:
                mssg = "Compilation error"
                self.get_result(code_return=str(result[0].returncode), message_out=mssg, time_usage=str(result[1]),
                                memory_usage=str(result[1]))
                return result


        test_input = self.__tests['input']
        required_output = self.__tests['output']

        run_command = __lang_config["run_command"] \
            .replace("$source_file_full_name", source_file_full_name) \
            .replace("$exec_file_full_name", exec_file_full_name) \
            .replace("$file_full_name", file_full_name)

        result = self.run_code(what_to_run=run_command, test_input_arr=test_input, required_output=required_output)
        self.get_result(code_return=str(result[0].returncode), message_out=result[1], time_usage=str(result[2]),
                        memory_usage=str(result[3]))
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
                    if result.returncode != 0:
                        mssg = "Runtime error, test #{}".format(str(test_i+1))
                        test_i = test_i + 1
                        break
                    elif result.stdout != required_output[i]:
                        mssg = "Wrong answer, test #{}".format(str(test_j+1))
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
                if result.returncode != 0:
                    mssg = "Runtime error"
                elif result.stdout != required_output[0]:
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
