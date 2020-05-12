from services.receiver import send_message


class Sender:
    def __init__(self, task_id, corr_id, user_id, solution_id, message_out, time_usage, memory_usage, is_completed=False, code_return=-1):
        self.task_id = task_id
        self.user_id = user_id
        self.code_return = code_return
        self.message_out = message_out
        self.time_usage = time_usage
        self.memory_usage = memory_usage
        self.corr_id = corr_id
        self.solution_id = solution_id
        self.completed = is_completed

    def send_students_result(self):
        message = {
            "id": self.solution_id,
            "taskId": str(self.task_id),
            "userId": self.user_id,
            "completed": self.completed,
            "codeReturn": self.code_return,
            "messageOut": self.message_out,
            "runtime": self.time_usage,
            "memory": self.memory_usage
        }
        response = send_message(message, self.corr_id)
