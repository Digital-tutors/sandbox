import json

def get_config_path():
    return app.dev_settings.CONFIG_PATH


def parse_config(file_name: str, lang: str, task_id: str, user_id: str, attempt: str) -> dict:
    config_path = get_config_path()
    with open(config_path, 'r', encoding="utf-8") as f:
        config = json.load(f)

    lang_config = config["lang_configs"]
    extension = lang_config[lang]["extension"]

    code_path = config["code_path"] + task_id + "_" + user_id + "/" + file_name + "_" + user_id + "_" + attempt
    source_code_path = code_path + extension
    compiler_path = lang_config[lang]["compiler_path"]

    result = {
        "code_path": code_path,
        "source_code_path": source_code_path,
        "compiler_path": compiler_path
        "lang_config": config["lang_configs"][lang]
    }
    return result