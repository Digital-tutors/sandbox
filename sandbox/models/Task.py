class Task:
    def __init__(self, task_id, topic_id, author_id, description, options, tests):
        self.task_id = task_id
        self.topic_id = topic_id
        self.author_id = author_id
        self.description = description
        self.options = options
        self.tests = tests
