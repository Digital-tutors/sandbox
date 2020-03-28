import json

class TaskEncoder(json.JSONEncoder):

    def default(self, o):
        try:
            to_serialize = {

            }
            return to_serialize
        except AttributeError:
            return super().default(o)