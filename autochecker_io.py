import json
import os
import shutil
from dotenv import load_dotenv

def get_config_path():
    dotenv_path = os.path.join(os.path.dirname(__file__), ".env")
    load_dotenv(dotenv_path)

    config_path = os.getenv("CONFIG_PATH")
    return config_path

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
