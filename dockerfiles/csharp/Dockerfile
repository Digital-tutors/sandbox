FROM alpine:3.5
# grab su-exec for easy step-down from root
RUN apk add --no-cache 'su-exec>=0.2'
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
	apk add --no-cache --update\
        bash \
#Programmin language support
    #Python 3
        python3 \
	    py2-pip \
	    python3-dev
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
