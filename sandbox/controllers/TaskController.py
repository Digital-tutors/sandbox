from models.Task import Task
import requests

class TaskController:
    def get_task_by_url(self, url):
        result = requests.get(url).json()
        task = Task(
            task_id=result["id"],
            topic_id=result["topicId"],
            author_id=result["authorId"],
            description=result["description"],
            options=result["options"],
            tests=result["tests"])
        return task
    


