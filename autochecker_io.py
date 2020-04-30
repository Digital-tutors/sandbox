#from ... import CONFIG_PATH
import json
import os

CONFIG_PATH = "/home/radmirkashapov/Projects/AutoChecker/config.json"

def save_code_in_file(code: str, lang: str, file_name: str, dir_path: str):
    extension = get_extension(CONFIG_PATH, lang)
    source_file_full_name = os.path.join(dir_path, file_name + extension)
    with open(source_file_full_name, 'w+', encoding="utf-8") as f:
        f.write(code)


def get_extension(config_path: str, lang: str):
    with open(config_path, 'r', encoding="utf-8") as f:
        config = json.load(f)

    lang_config = config["lang_configs"]
    extension = lang_config[lang]["extension"]
    return extension
