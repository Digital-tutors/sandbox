FROM openjdk:8-jre-alpine
# grab su-exec for easy step-down from root
RUN apk add --no-cache 'su-exec>=0.2' && \
    echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
	apk add --no-cache --update\
        bash \
        unzip \
#Programmin language support
    #Python 3
        python3 \
	    py2-pip \
	    python3-dev
	# Kotlin
ENV     KOTLIN_VERSION=1.3.72 \
        KOTLIN_HOME=/usr/local/kotlin
RUN     apk update && \
        apk add ca-certificates && \
        update-ca-certificates && \
        apk add openssl wget bash

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
RUN pip install --no-cache --upgrade pip
WORKDIR /sandbox
# Translators configs
COPY /sandbox /sandbox
RUN python3 -m pip install  --trusted-host pypi.python.org --trusted-host files.pythonhosted.org --trusted-host pypi.org  -r /sandbox/resources/requirements.txt
RUN	chmod +x /sandbox/run.py
ENV COMPILER="not"
ENV FILE_NAME="file"
ENV TASK_ID=0
ENV CORR_ID=0
ENV USER_ID=0
ENV IS_TEST_CREATION="True"
ENV SOLUTION_ID=0
CMD python3 -u /sandbox/run.py -lang $COMPILER -file_name $FILE_NAME  -task_num $TASK_ID -corr_id $CORR_ID -user_id $USER_ID -is_test_creation $IS_TEST_CREATION -solution_id $SOLUTION_ID
#test 
