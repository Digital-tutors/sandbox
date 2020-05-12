import subprocess
from services.configurer import parse_config
from services.ResourceMonitor import ResourceMonitor
from concurrent.futures import ThreadPoolExecutor
from controllers.TaskController import TaskController
from services.sender import Sender


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
        self.task = taskObj.get_task_by_url("http://host.docker.internal:8080/task/{taskId}/admin/".format(taskId=self.__task_id))
        self.__tests = self.task.tests
        self.__time_limit = int(self.task.options["timeLimit"])
        self.__memory_limit = int(self.task.options["memoryLimit"])
        self.test_code()

    def get_result(self, code_return: str = None, message_out: str = None, time_usage: str = "",
                   memory_usage: str = "", is_completed: bool = False):
        sender = Sender(task_id=self.__task_id, corr_id=self.__corr_id, solution_id=self.__solution_id, user_id=self.__user_id, code_return=code_return,
                        message_out=message_out, time_usage=time_usage, memory_usage=memory_usage, is_completed=is_completed)
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

            process_object = result["completed_process_object"]
            memory_usage = result["memory_usage"]
            time_usage = result["time_usage"]

            if str(process_object.returncode) != "0":
                mssg = "Compilation error"
                self.get_result(code_return=str(process_object.returncode), message_out=mssg, time_usage=str(time_usage),
                                memory_usage=str(memory_usage), is_completed=False)
                return result


        test_input = self.__tests['input']
        required_output = self.__tests['output']

        run_command = __lang_config["run_command"] \
            .replace("$source_file_full_name", source_file_full_name) \
            .replace("$exec_file_full_name", exec_file_full_name) \
            .replace("$file_full_name", file_full_name)

        result = self.run_code(what_to_run=run_command, test_input_arr=test_input, required_output=required_output)

        process_object = result["completed_process_object"]
        memory_usage = result["memory_usage"]
        time_usage = result["time_usage"]
        is_completed = result["is_completed"]
        message = result["message"]

        self.get_result(code_return=str(process_object.returncode), message_out=message, time_usage=str(time_usage),
                        memory_usage=str(memory_usage), is_completed=is_completed)
        return message

    def run_code(self, what_to_run: str, test_input_arr: list, required_output: list):
        monitor = ResourceMonitor(self.__time_limit, self.__memory_limit)
        is_completed = True
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
                    if str(result.returncode) != "0":
                        mssg = "Runtime error, test #{}".format(str(i+1))
                        is_completed = False
                        break
                    elif result.stdout != required_output[i]:
                        mssg = "Wrong answer, test #{}".format(str(i+1))
                        is_completed = False
                        break
                    else:
                        mssg = "Correct answer"
                        is_completed = True


            else:
                mem_thread = executor.submit(monitor.memory_usage)
                time_thread = executor.submit(monitor.timeout_usage)
                fn_thread = executor.submit(self.__run_and_check, what_to_run=what_to_run, monitor=monitor)
                result = fn_thread.result()
                monitor.keep_measuring = False
                max_mem_usage = mem_thread.result()
                max_time_usage = time_thread.result()
                if str(result.returncode) != "0":
                    mssg = "Runtime error"
                    is_completed = False
                elif result.stdout != required_output[0]:
                    mssg = "Wrong answer"
                    is_completed = False
                else:
                    mssg = "Correct answer"

        result_obj = {
            "completed_process_object": result,
            "memory_usage": max_mem_usage,
            "time_usage": max_time_usage,
            "is_completed": is_completed,
            "message": mssg
        }

        return result_obj

    def __run_and_check(self, what_to_run: str, test_input: str = None, monitor: ResourceMonitor = None):
        if test_input is not None:
            input_test_string = test_input
        else:
            input_test_string = None
        completed_run = subprocess.run(
                what_to_run,
                input=input_test_string,
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                timeout=self.__time_limit,
                preexec_fn=monitor.set_memory_limit)
        if completed_run.stderr == subprocess.TimeoutExpired:
            monitor.keep_measuring = False
        return completed_run

    def compile_file(self, compiler_path, args):
        with ThreadPoolExecutor() as executor:
            monitor = ResourceMonitor(self.__time_limit, self.__memory_limit)
            mem_thread = executor.submit(monitor.memory_usage)
            time_thread = executor.submit(monitor.timeout_usage)
            fn_thread = executor.submit(self.compile, compiler_path, args, monitor)
            result = fn_thread.result()
            monitor.keep_measuring = False
            mem_usage = mem_thread.result()
            time_usage = time_thread.result()

            result = {
                "completed_process_object": result,
                "memory_usage": mem_usage,
                "time_usage": time_usage
            }

            return result

    # Function that compile source code.
    # Return result of compilation
    def compile(self, compiler_path: str, path_args: str, monitor: ResourceMonitor):
            compilation = subprocess.run(' '.join([compiler_path, path_args]),
                                         shell=True,
                                         stdout=subprocess.PIPE,
                                         stderr=subprocess.STDOUT,
                                         universal_newlines=True,
                                         timeout=self.__time_limit,
                                         preexec_fn=monitor.set_memory_limit)
            if compilation.stderr == subprocess.TimeoutExpired:
                monitor.keep_measuring = False
            return compilation
