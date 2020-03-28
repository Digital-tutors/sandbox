import abc
from typing import Iterable

class Task(object):
    def __init__(self, id:str=None, id_author:str=None, description:str=None, options:dict=None, tests:dict=None):
        self.id = id
        self.id_author = id_author
        self.description = description
        self.options = options
        self.tests = tests

class TaskDTO(object, metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def create(self, task: Task) -> Task:
        pass

    @abc.abstractmethod
    def update(self, task: Task) -> Task:
        pass

    @abc.abstractmethod
    def get_all(self) -> Iterable[Task]:
        pass

    @abc.abstractmethod
    def get_by_id(self, task_id: str) -> Task:
        pass

class TaskNotFound(object):
    pass