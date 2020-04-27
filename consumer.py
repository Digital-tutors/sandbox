#!/usr/bin/env python
import pika
import traceback
import sys
import json
import os
import logging
import pathlib
from autochecker_io import save_code_in_file
from DockerSandbox import DockerSandbox
from logger import initialize_logger
from dotenv import load_dotenv

dotenv_path = os.join(os.path.dirname(__file__), ".env")
load_dotenv(dotenv_path)

code_storage_path = os.getenv("CODE_STORAGE_PATH")
network = os.getenv("DOCKER_NETWORK")
task_queue = os.getenv("TASK_QUEUE")
queue_exchange = os.getenv("QUEUE_EXCHANGE")

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()
channel.queue_declare(queue=task_queue, durable=True)


def callback(ch, method, props, body):
    body = json.loads(body.decode("utf-8"))
    lang = str(body[1]).lower()
    task_id = body[0]
    user_id = 0
    attempt = 0
    code = body[2]
    is_test_creation = False

    code_file_name = task_id + "_" + user_id
    if (is_test_creation):
        file_name = code_file_name + "_" + "test"
    # id задачи + id пользователя + номер попытки
    else:
        file_name = code_file_name + "_" + attempt
    dir_path = code_storage_path + "{dir}/".format(dir=code_file_name)
    if (not os.path.exists(dir_path)):
        pathlib.Path(dir_path).mkdir(parents=True, exist_ok=True)

    save_code_in_file(
        code=code,
        lang=lang,
        file_name=file_name,
        dir_path=dir_path)

    docker_Sandbox = DockerSandbox(
        compiler_name=lang,
        path=dir_path,
        file_name=file_name,
        task_id=task_id,
        corr_id=body[4],
        user_id=body[5],
        is_test_creation=is_test_creation)
    docker_Sandbox.execute()

    response = "OK"
    ch.basic_publish(exchange=queue_exchange,
                     routing_key=task_queue,
                     properties=pika.BasicProperties(correlation_id=props.correlation_id),
                     body=response)
    ch.basic_ack(delivery_tag=method.delivery_tag)


channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue=task_queue, on_message_callback=callback)

try:
    channel.start_consuming()
except Exception:
    channel.stop_consuming()
    traceback.print_exc(file=sys.stdout)
