#!/usr/bin/env python
import pika
import traceback
import sys
import json
import os
import pathlib
from autochecker_io import save_code_in_file
from DockerSandbox import DockerSandbox
from dotenv import load_dotenv
import uuid

dotenv_path = os.path.join(os.path.dirname(__file__), ".env")
load_dotenv(dotenv_path)

code_storage_path = os.getenv("CODE_STORAGE_PATH")
network = os.getenv("DOCKER_NETWORK")
task_queue = os.getenv("TASK_QUEUE")
queue_exchange = os.getenv("QUEUE_EXCHANGE")
hostname = os.getenv("HOSTNAME")
container_name = os.getenv("CONTAINER_NAME")

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()
channel.queue_declare(queue=task_queue, durable=True)
print("Wait for messages")


def callback(ch, method, props, body):
    body = json.loads(body.decode("utf-8"))
    solution_id = body["id"]
    task_id = body["taskId"]
    user_id = body["userId"]
    lang = body["language"]
    source_code = body["sourceCode"]
    is_test_creation = False

    code_file_name = str(solution_id)
    file_name = code_file_name
    dir_path = code_storage_path + "{dir}/".format(dir=code_file_name)
    print("Directory name: {}".format(dir_path))
    if (not os.path.exists(dir_path)):
        pathlib.Path(dir_path).mkdir(parents=True, exist_ok=True)

    save_code_in_file(
        code=source_code,
        lang=lang,
        file_name=file_name,
        dir_path=dir_path)

    docker_Sandbox = DockerSandbox(
        compiler_name=lang,
        path=dir_path,
        file_name=file_name,
        task_id=task_id,
        corr_id=uuid.uuid4(),
        user_id=user_id,
        hostname=hostname,
        solution_id=solution_id,
        is_test_creation=is_test_creation,
        container_name=container_name,
        network=network)
    docker_Sandbox.execute()


channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue=task_queue, on_message_callback=callback)

try:
    channel.start_consuming()
except Exception:
    channel.stop_consuming()
    traceback.print_exc(file=sys.stdout)
