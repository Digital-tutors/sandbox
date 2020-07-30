FROM golang:1.15beta1-alpine3.12
# grab su-exec for easy step-down from root
RUN apk add --no-cache 'su-exec>=0.2' && \
    echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
	apk add --no-cache --update\
        bash \
        unzip \
        openjdk11 \
#Programmin language support
	# Kotlin
ENV     KOTLIN_VERSION=1.3.72 \
        KOTLIN_HOME="/usr/local/kotlin"
RUN     apk update && \
        apk add ca-certificates && \
        update-ca-certificates && \
        apk add openssl wget bash git

RUN     cd  /tmp && \
        wget -q -k "https://github.com/JetBrains/kotlin/releases/download/v${KOTLIN_VERSION}/kotlin-compiler-${KOTLIN_VERSION}.zip"  && \
        unzip "kotlin-compiler-${KOTLIN_VERSION}.zip" && \
        mkdir -p "${KOTLIN_HOME}" && \
        mv "/tmp/kotlinc/bin" "/tmp/kotlinc/lib" "${KOTLIN_HOME}" && \
        rm ${KOTLIN_HOME}/bin/*.bat && \
        chmod +x ${KOTLIN_HOME}/bin/* && \
        ln -s "${KOTLIN_HOME}/bin/"* "/usr/bin/" && \
        apk del wget ca-certificates curl openssl && \
        rm -rf /tmp/* /var/cache/apk/*
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
RUN chmod +x /sandbox/sandbox/main.go
CMD go run /sandbox/sandbox/main.go
#test 
