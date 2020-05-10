#!/usr/bin/env student-python
import pika
import json

class Receiver(object):

    def __init__(self):
        self.connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='my-rabbit', port=5672))

        self.channel = self.connection.channel()

        result = self.channel.queue_declare(queue='program.results', durable=True)
        self.channel.queue_bind(exchange='program',
                           queue='program.results',
                           routing_key='program.results')
        self.callback_queue = result.method.queue


    def call(self, mssg, corr_id):
        self.response = None
        self.corr_id = corr_id
        self.channel.basic_publish(
            exchange='program',
            routing_key='program.results',
            properties=pika.BasicProperties(
                reply_to=self.callback_queue,
                correlation_id=self.corr_id,
                delivery_mode=2
            ),
            body=json.dumps(mssg))


def send_message(result, corr_id):
    receiver = Receiver()
    receiver.call(result, corr_id)
