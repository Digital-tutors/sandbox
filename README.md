Модуль ```sandbox.services.operators_checker``` предназначен для нахождения в тексте программы запрещенных конструкций (операторов и ключевых слов). 

Правильность разпознавания зависит от выбранного языка программирования. Поддерживаются 6 языков: C, C++, Python, JS, PHP, Kotlin.

Пример использования - распознавание ключевых слов if и else, бинарного оператора "+", унарного префиксного оператора "++" и тернарного оператора "?:" 
в некотором коде на С.

```python
from sandbox.services.operators_checker import *

checker = OperatorsChecker(Language.C)

code = """
    void main() {
        int i = 0;
        if (true) {
          i++;
          ++i;
          1 + 2;
        }
        */ 
        if
        /*
        bool odd = i / 2 == 1 ? true : false;
        char* c = "else";
    }
"""

forbidden_elements = [
    Keyword("if"),
    Keyword("else"),
    BinaryOperator("+"),
    PrefixOperator("++"),
    TernaryOperator("?", ":")
]

result = checker.check(code, forbidden_elements)

for el in result:
    print(el)
```

Вывод:
```
Keyword(if)
BinaryOperator(+)
PrefixOperator(++)
```

Для проверки сперва создаётся объект класса ```OperatorsChecker```, в конструкторе указывается язык проверяемой программы 
либо в виде строки (```"c"```, ```"cpp"```, ```"python"```, ```"js"```, ```"php"```, ```"kotlin"```), 
либо в виде значения перечисления ```Language``` 
(```Language.C```, ```Language.CPP```, ```Language.PYTHON```, ```Language.JS```, ```Language.PHP```, ```Language.KOTLIN```).

Непосредственно проверка осуществляется методом ```checker.check(code, forbidden_elements)```. 
Первый аргумент - текст проверяемой программы в виде строки. 
Второй аргумент - список элементов языка, которые необходимо найти. Они представляют собой экземпляры следующих классов:
- ```PrefixOperator(oper)``` - унарный префиксный оператор
- ```PostfixOperator(oper)``` - унарный постфиксный оператор
- ```BinaryOperator(oper)``` - бинарный оператор
- ```TernaryOperator(oper1, oper2)``` - тернарный оператор
- ```PrefixOperator(kw)``` - ключ. слово

Аргументы ```oper``` и ```kw``` конструкторов обозначают строковое представление ключевого слова / оператора. Исключение - тернарный оператор. 
Там ```oper1```и ```oper2``` - "части" оператора (например, для тернарного оператора "?:" это "?" и ":").

Метод возвращает список тех элементов исхдного списка, которые были найдены в тексте. 
Если ничего не найдено или исходный список пуст, то возвращает, соответственно, пустой список.

При проверке игнорируются комментарии и строковые литералы.
