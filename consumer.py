#!/usr/bin/env python
import pika, traceback, sys, json
import os, logging
from DockerSandbox import DockerSandbox
from logger import initialize_logger
import pathlib
import pickle
import stat

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.queue_declare(queue='program.tasks', durable=True)
print(' [*] Waiting for messages. To exit press CTRL+C')

def callback(ch, method, props, body):
    body = json.loads(body.decode('utf-8'))
    print(type(body))
    print(" [x] Received %r" % body)
    lang = str(body[1]).lower()
    task_num = body[0]
    code = body[2]
    # id задачи + id пользователя + номер попытки
    __code_file_name = body[2]
    #checker_core.save_code_in_file()
    
    dir_path = "/home/radmirkashapov/Projects/Test/target/{dir}/".format(dir=__code_file_name)
    if(not os.path.exists(dir_path)):
        pathlib.Path("/home/radmirkashapov/Projects/Test/target/{dir}/".format(dir=__code_file_name)).mkdir(parents=True, exist_ok=True)

    
    docker_Sandbox = DockerSandbox(compiler_name = "python3", path = "~/Projects/Test/target", file_name=__code_file_name, corr_id=body[4], user_id=body[5])
    docker_Sandbox.execute()
    logging.info(res)
    print("OK")
    response="OK"
    ch.basic_publish(exchange='program',
                     routing_key='program.tasks',
                     properties=pika.BasicProperties(correlation_id = props.correlation_id),
                     body=response)
    ch.basic_ack(delivery_tag=method.delivery_tag)


channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue='program.tasks', on_message_callback=callback)


try:
    channel.start_consuming()
except Exception:
    channel.stop_consuming()
    traceback.print_exc(file=sys.stdout)
