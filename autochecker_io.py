#from ... import CONFIG_PATH
import json
import os
import shutil

def get_config_path():
    return "/mnt/c/Users/mylif/source/repos/Projects/AutoChecker/config.json"

def save_code_in_file(code: str, lang: str, file_name: str, dir_path: str):
    extension = get_extension(get_config_path(), lang)
    source_file_full_name = os.path.join(dir_path, file_name + extension)
    with open(source_file_full_name, 'w+', encoding="utf-8") as f:
        f.write(code)


def get_extension(config_path: str, lang: str):
    with open(config_path, 'r', encoding="utf-8") as f:
        config = json.load(f)

    lang_config = config["lang_configs"]
    extension = lang_config[lang]["extension"]
    return extension


def delete_file(dir_path):
    try:
        shutil.rmtree(dir_path)
    except OSError as e:
        print("Error: %s - %s." % (e.filename, e.strerror))
