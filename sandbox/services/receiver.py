#!/usr/bin/env python
import pika
import json

class Receiver(object):

    def __init__(self):
        self.connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='my-rabbit', port=5672))

        self.channel = self.connection.channel()

        result = self.channel.queue_declare(queue='program.results', durable=True)
        self.callback_queue = result.method.queue

        self.channel.basic_consume(
            queue=self.callback_queue,
            on_message_callback=self.on_response,
            auto_ack=True)

    def on_response(self, ch, method, props, body):
        if self.corr_id == props.correlation_id:
            self.response = body

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
        while self.response is None:
            self.connection.process_data_events()
        return str(self.response)


def send_message(result, corr_id):
    tutor_rpc = Receiver()
    response = tutor_rpc.call(result, corr_id)
    return response
