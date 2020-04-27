import subprocess
import os
import docker

TARGET_FILE_PATH = "/app/target/"
IMAGES = {
    "c++": "autochecker-cpp",
    "c": "autochecker-clang",
    "python3": "autochecker-student-python"
}

class DockerSandbox(object):

    def __init__(self, compiler_name, path, file_name, task_id, corr_id, user_id, is_test_creation, network):
        self.__client = docker.from_env()
        self.__compiler_name = str(compiler_name).lower()
        self.__file_name = str(file_name)
        self.__path = path
        self.__task_id = task_id
        self.__corr_id = str(corr_id)
        self.__user_id = user_id
        self.__is_test_creation = is_test_creation
        self.__enviroment = [
            "COMPILER={}".format(self.__compiler_name),
            "FILE_NAME={}".format(self.__file_name),
            "TASK_ID={}".format(self.__task_id),
            "CORR_ID={}".format(self.__corr_id),
            "USER_ID={}".format(self.__user_id),
            "IS_TEST_CREATION={}".format(self.__is_test_creation)
        ]
        self.__sandbox_image = IMAGES[self.__compiler_name]
        self.__network = network
        self.__volume = {
            self.__path: {
                "bind": self.set_volume(),
                "mode": "ro"
            }
        }

    def set_volume(self):
        file_name = self.__file_name
        '_'.join(file_name.split('_')[:-1])
        return TARGET_FILE_PATH + file_name

    def execute(self):
        self.__client.containers.run(
            image=self.__sandbox_image,
            environment=self.__enviroment,
            network=self.__network,
            remove=True,
            volumes=self.__volume)