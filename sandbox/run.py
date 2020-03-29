from checker_core import run_on_simple_tests 
import os, logging, uuid
import argparse

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-lang', dest="lang", help='Programming language')
    parser.add_argument('-file_name', dest="name", help='File name')
    parser.add_argument('-task_num', dest="taskId", help='Task ID')
    parser.add_argument('-corr_id', dest="correlationID", help='Correlation ID for rabbit')
    parser.add_argument('-user_id', dest="userID", help='User ID')

    args, unknown = parser.parse_known_args()
    logging.info(args)
    task_id = str(args.taskId)
    lang = str(args.lang)
    file_name = str(args.name)
    corr_id = str(args.correlationID)
    user_id = str(args.userID)
    
    data = run_on_simple_tests(task_num=task_id, lang=lang, file_name=file_name, userId=user_id, corr_id=corr_id)
    print(data)

if __name__=="__main__":
    main()
