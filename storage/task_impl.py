# backend/storage/Task_impl.py

from typing import Iterable

import bson
import bson.errors
from pymongo.collection import Collection
from pymongo.database import Database

from domain.storage.Task import Task, TaskDAO, TaskNotFound


class MongoTaskDAO(TaskDAO):

    def __init__(self, mongo_database: Database):
        self.mongo_database = mongo_database

    @property
    def collection(self) -> Collection:
        return self.mongo_database["Tasks"]

    @classmethod
    def to_bson(cls, task: Task):
        # MongoDB хранит документы в формате BSON. Здесь
        # мы должны сконвертировать нашу задачу в BSON-
        # сериализуемый объект, что бы в ней ни хранилось.
        result = {
            k: v
            for k, v in Task.__dict__.items()
            if v is not None
        }
        if "id" in result:
            result["_id"] = bson.ObjectId(result.pop("id"))
        return result

    @classmethod
    def from_bson(cls, document) -> Task:
        # С другой стороны, мы хотим абстрагировать весь
        # остальной код от того факта, что мы храним задачи
        # в монге. Но при этом id будет неизбежно везде
        # использоваться, так что сконвертируем-ка его в строку.
        document["id"] = str(document.pop("_id"))
        return Task(**document)

    def create(self, task: Task) -> Task:
        task.id = str(self.collection.insert_one(self.to_bson(task)).inserted_id)
        return task

    def update(self, task: Task) -> Task:
        task_id = bson.ObjectId(task.id)
        self.collection.update_one({"_id": task_id}, {"$set": self.to_bson(task)})
        return task

    def get_all(self) -> Iterable[Task]:
        for document in self.collection.find():
            yield self.from_bson(document)

    def get_by_id(self, task_id: str) -> Task:
        return self._get_by_query({"_id": bson.ObjectId(task_id)})