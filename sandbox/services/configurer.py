import json
from dotenv import load_dotenv
import os

def get_config_path():
    dotenv_path = os.path.join(os.path.dirname(__file__), ".env")
    load_dotenv(dotenv_path)
    config_path = os.getenv("CONFIG_PATH")
    return config_path

def parse_config(file_name: str, lang: str) -> dict:
    config_path = get_config_path()
    with open(config_path, 'r', encoding="utf-8") as f:
        config = json.load(f)
    lang_config = config["lang_configs"]

    extension = lang_config[lang]["extension"]

    code_path = config["code_path"] + file_name + "/" + file_name
    source_code_path = code_path + extension
    compiler_path = lang_config[lang]["compiler_path"]

    result = {
        "code_path": code_path,
        "source_code_path": source_code_path,
        "compiler_path": compiler_path,
        "lang_config": config["lang_configs"][lang]
    }
    return result
