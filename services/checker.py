from app.storage.task import TaskNotFound
from app.storage.context import Context
from app.services.configurer import parse_config
from app.services.MemoryMonitor import MemoryMonitor
from concurrent.futures import ThreadPoolExecutor
from app.services.receiver import TutorRpcClient
import datetime

class Checker(object):

    def __init__(self, task_id:str=None, lang:str=None, file_name:str=None, user_id:str=None, corr_id:str=None):
        self.__task_id = task_id
        self.__lang = lang
        self.__file_name = file_name
        self.__user_id = user_id
        self.__corr_id = corr_id
        self.context = Context()

    def get_result(self, code_return:str=None, message_out:str=None, runtime:str="", memory:str=""):
        attempt:str = "0"
        message = {
            "taskId": self.__task_id, 
            "userId": self.__user_id,
            "attempt": attempt,
            "codeReturn": code_return,
            "messageOut": message_out,
            "runtime": runtime,
            "memory": memory
            }
        self.send_message(message, corr_id)
    
    def send_message(self, message:dict):
        rpc_receiver = TutorRpcClient()
        response = rpc_receiver.call(message, self.__corr_id)

    def get_tests(self):
        task = self.context.task_dao.get_by_id(self.__task_id)
        return task.tests
    
    def compilate(self, compiler_path:str, path_args:str):
        start_time = datetime.datetime.now()
        compilation = subprocess.run(' '.join([compiler_path, path_args]), 
                shell=True, 
                stdout=subprocess.PIPE, 
                stderr=subprocess.STDOUT,
                universal_newlines=True)
        end_time = datetime.datetime.now()
        timeout = end_time - start_time
        return [compilation, timeout]
    
    def __run_and_check(self, what_to_run:str, test_input:list):
        start_time = datetime.datetime.now()
        completed_run = subprocess.run(
                        what_to_run,
                        input=test_input, 
                        stdout=subprocess.PIPE, 
                        stderr=subprocess.STDOUT,
                        universal_newlines=True,
                        shell=True
                        )
        end_time = datetime.datetime.now()
        timeout = end_time - start_time
        return [completed_run, timeout]

    def run_code(self, what_to_run:str, test_input_arr:list, required_output:list):
        max_time = 0
        max_memory = 0
        test_i = 0
        test_j = 0
        if len(test_input_arr) != 0:
            for i in range(len(test_input_arr)):
                with ThreadPoolExecutor() as executor:
                    monitor = MemoryMonitor()
                    mem_thread = executor.submit(monitor.memory_usage)
                    fn_thread = executor.submit(self.run_code, what_to_run, test_input_arr[i])
                    result = fn_thread.result()
                    monitor.keep_measuring = False
                    max_usage = mem_thread.result()
                    if(max_time < result[1]):
                        max_time = timeout
                    if(max_memory < max_usage):
                        max_memory = max_usage
                if result[0].returncode != 0:
                    mssg="Runtime error, test #{}".format(str(test_i))
                    test_i = test_i + 1
                    break
                elif result[0].stdout != required_output[i]:
                    mssg="Wrong answer, test #{}".format(str(test_j))
                    test_j=i+1
                    break
                else:
                    mssg = "Correct answer"
        else:
            with ThreadPoolExecutor() as executor:
                    monitor = MemoryMonitor()
                    mem_thread = executor.submit(monitor.memory_usage)
                    fn_thread = executor.submit(self.run_code, what_to_run, test_input_arr[i])
                    result = fn_thread.result()
                    monitor.keep_measuring = False
                    max_usage = mem_thread.result()
                    if(max_time < result[1]):
                        max_time = timeout
                    if(max_memory < max_usage):
                        max_memory = max_usage
            if result[0].returncode != 0:
                    mssg="Runtime error"
                    break
                elif result[0].stdout != required_output[0]:
                    mssg="Wrong answer"
                    break
                else:
                    mssg = "Correct answer"

        return [result, mssg, max_time, max_memory]

    def test_code(self) -> str:
        config = parse_config()
        is_compilable = config["is_compilable"]
        
        source_file_full_name = config["source_code_path"]
        exec_file_full_name = config["code_path"]
        compiler_path = config["lang_config"]["compiler_path"]
        
        args = __lang_config[lang]["args_format"]\
                .replace("$source_file_full_name" , source_file_full_name)\
                .replace("$exec_file_full_name", exec_file_full_name)
        
        if is_compilable:
            with ThreadPoolExecutor() as executor:
                monitor = MemoryMonitor()
                mem_thread = executor.submit(monitor.memory_usage)
                fn_thread = executor.submit(self.compilate, compiler_path, args)
                result = fn_thread.result()
                monitor.keep_measuring = False
                max_usage = mem_thread.result()

            if result[0].returncode !=0:
                mssg = "Compilation error"
                set_stdout(code_return=str(result[0].returncode), message_out=mssg, runtime=str(result[1]), memory=str(max_usage))
                return result

            tests = get_tests()
            task_tests = tests[str(task_num)]
            test_input = task_tests['input']
            required_output = task_tests['output']

            what_to_run = os.path.join('.', exec_file_full_name) if is_compilable else ' '.join([compiler_path, args])
            run_code(what_to_run=what_to_run, test_input_arr=test_input, required_output=required_output)
            set_stdout(code_return=str(result[0].returncode), message_out=result[1], runtime=str(result[2]), memory=str(result[3]))
            return result[1]
