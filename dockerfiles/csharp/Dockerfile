FROM golang:1.15beta1-alpine3.12
# grab su-exec for easy step-down from root
RUN apk add --no-cache 'su-exec>=0.2'
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
	apk add --no-cache --update\
        bash \
        git \
#Programmin language support
	#Mono for C# support
	    mono --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing && \
        apk add --no-cache --virtual=.build-dependencies ca-certificates && \
        cert-sync /etc/ssl/certs/ca-certificates.crt && \
        apk del .build-dependencies
# Sandbox workdir
WORKDIR /sandbox
# Translators configs
COPY /cmd/sandbox /sandbox/sandbox
COPY /cmd/main/languageConfig.json /sandbox/sandbox/languageConfig.json
COPY /cmd/rabbit /sandbox/rabbit
COPY /cmd/config /sandbox/config
COPY /cmd/solution /sandbox/solution

ENV GOCACHE="/tmp/.cache"

RUN go get "github.com/streadway/amqp"
RUN mkdir /sandbox/target && chmod -R 777 /sandbox/target
CMD go run /sandbox/sandbox/main.go
#test 
