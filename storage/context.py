import os

from pymongo import MongoClient
from pymongo.database import Database

import app.dev_settings
from app.storage.task import TaskDAO
from app.storage.task_impl import MongoTaskDAO


class Context(object):

    def __init__(self):

        self.mongo_client: MongoClient = MongoClient(
            host=app.dev_settings.MONGO_HOST,
            port=app.dev_settings.MONGO_PORT)
        self.mongo_database: Database = self.mongo_client[self.app.dev_settings.MONGO_DATABASE]
        self.task_dao: TaskDAO = MongoTaskDAO(self.mongo_database)