import re
from typing import *
from abc import *
from enum import Enum


class LanguageInfo:
    def __init__(
            self,
            name: str,
            extension: str,
            single_line_comment=r"//.*\n?",
            multi_line_comment=r"/\*(?:[^*]|\*(?!\/))*(?:\*/)?"
    ):
        self.name = name
        self.extension = extension
        self.single_line_comment = single_line_comment
        self.multi_line_comment = multi_line_comment


class Language(Enum):
    C = LanguageInfo("c", ".c")
    CPP = LanguageInfo("cpp", ".cpp")
    PYTHON = LanguageInfo("python", ".py", r"#.*\n?")
    JS = LanguageInfo("js", ".js")
    PHP = LanguageInfo("php", ".php")
    KOTLIN = LanguageInfo("kotlin", ".kt")


class LanguageElement(ABC):
    @abstractmethod
    def get_regex(self) -> re.Pattern:
        pass


class PrefixOperator(LanguageElement):
    def __init__(self, oper: str):
        self.oper: str = oper

    def get_regex(self):
        return re.compile(re.escape(self.oper) + """(?=['\\";\w\s\n])""")


class PostfixOperator(LanguageElement):
    def __init__(self, oper: str):
        self.oper: str = oper

    def get_regex(self):
        return re.compile("""(?<=['\\"\w\s])""" + re.escape(self.oper))


class BinaryOperator(LanguageElement):
    def __init__(self, oper: str):
        self.oper: str = oper

    def get_regex(self):
        return re.compile("""(?<=['\\"\w\s])""" + re.escape(self.oper) + """(?=['\\";\w\s\n])""")


class TernaryOperator(LanguageElement):
    def __init__(self, oper1: str, oper2: str):
        self.oper1 = oper1
        self.oper2 = oper2

    def get_regex(self):
        return re.compile(
            """(?<=['\\"\w\s])""" + re.escape(self.oper1)
            + """(?=['\\"\w\s\n](?<=['\\"\w\s]))""" + re.escape(self.oper2)
            + """(?=['\\";\w\s\n])"""
        )


class Keyword(LanguageElement):
    def __init__(self, kw: str):
        self.kw = kw

    def get_regex(self):
        return re.compile(rf"\b{self.kw}\b")


class OperatorsChecker:
    def __init__(self, lang: Union[Language, str]):
        if type(lang) == str:
            lang = Language[lang.upper()]
        self.lang: Language = lang

    def check(self, code: str, forbidden_elements: List[LanguageElement]) -> List[LanguageElement]:
        cleaned_code = self.__process_code(code)
        found_elements = []
        for element in forbidden_elements:
            if re.search(element.get_regex(), cleaned_code) is not None:
                found_elements.append(element)
        return found_elements

    def __process_code(self, code: str):
        single_line_comment = self.lang.value.single_line_comment
        multi_line_comment = self.lang.value.multi_line_comment
        string_regex = """(['"]).*?\\1"""
        unified_regex = re.compile(f"{single_line_comment}|{multi_line_comment}|{string_regex}")
        return re.sub(unified_regex, "", code)