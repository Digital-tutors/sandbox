from services import checker
import os
import argparse

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-lang', dest="lang", help='Programming language')
    parser.add_argument('-file_name', dest="name", help='File name')
    parser.add_argument('-task_num', dest="taskId", help='Task ID')
    parser.add_argument('-corr_id', dest="correlationID", help='Correlation ID for rabbit')
    parser.add_argument('-user_id', dest="userID", help='User ID')
    parser.add_argument('-is_test_creation', dest="is_test_creation", help='Checking whay we run: task solver or test generator')
    parser.add_argument('-solution_id', dest="solution_id", help='Solution ID')

    args = parser.parse_args()
    task_id = str(args.taskId)
    lang = str(args.lang)
    file_name = str(args.name)
    corr_id = str(args.correlationID)
    user_id = str(args.userID)
    solution_id = str(args.solution_id)
    is_test_creation = True if args.is_test_creation == "True" else False
    if not is_test_creation:
        sandbox = checker.Checker(task_id=task_id, lang=lang, file_name=file_name, user_id=user_id, corr_id=corr_id, solution_id=solution_id)

if __name__=="__main__":
    main()