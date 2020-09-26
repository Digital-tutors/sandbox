FROM golang:1.15beta1-alpine3.12
# grab su-exec for easy step-down from root
RUN apk add --no-cache 'su-exec>=0.2'
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
	apk add --no-cache --update\
        bash \
        git \
#Programmin language support
    #gcc (ASM32, ASM64, C)
        binutils \
		cmake \
		make \
		libgcc \
		musl-dev \
		gcc
# Sandbox workdir
WORKDIR $GOPATH/src/sandbox
# Translators configs
COPY /cmd/sandbox $GOPATH/src/sandbox/cmd/sandbox
COPY /cmd/main/languageConfig.json $GOPATH/src/sandbox/cmd/sandbox/languageConfig.json
COPY /cmd/rabbit/publishResult.go $GOPATH/src/sandbox/cmd/rabbit/publishResult.go
COPY /cmd/config $GOPATH/src/sandbox/cmd/config
COPY /cmd/solution $GOPATH/src/sandbox/cmd/solution

ENV GOCACHE="/tmp/.cache"

RUN go get "github.com/streadway/amqp"
RUN mkdir /target && chmod -R 777 /target
CMD go run $GOPATH/src/sandbox/cmd/sandbox/main.go
#test 
