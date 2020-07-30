package operators

import (
	"../solution"
	"fmt"
	"log"
	"regexp"
)

type LanguageElement interface {
	getRegex()  (*regexp.Regexp, error)
}

type LanguageInfo struct {
	Name string
	Extension string
	SingleLineComment string
	MultiLineComment string
}

type Checker struct {
	LanguageInfo LanguageInfo
}


type PrefixOperator struct {
	Operator string
}

func (prefix *PrefixOperator) getRegex() (*regexp.Regexp, error) {
	return regexp.Compile(regexp.QuoteMeta(prefix.Operator) + `(?=['\\";\w\s\n])`)
}

type PostfixOperator struct {
	Operator string
}

func (prefix *PostfixOperator) getRegex() (*regexp.Regexp, error) {
	return regexp.Compile( `(?<=['\\"\w\s])` + regexp.QuoteMeta(prefix.Operator))
}

type BinaryOperator struct {
	Operator string
}

func (binary *BinaryOperator) getRegex() (*regexp.Regexp, error) {
	return regexp.Compile( `(?<=['\\"\w\s])` + regexp.QuoteMeta(binary.Operator) + `(?=['\\";\w\s\n])`)
}

type TernaryOperator struct {
	OperatorFirst string
	OperatorSecond string
}

func (ternary *TernaryOperator) getRegex() (*regexp.Regexp, error) {
	return regexp.Compile( `(?<=['\\"\w\s])` + regexp.QuoteMeta(ternary.OperatorFirst) + `(?=['\\"\w\s\n](?<=['\\"\w\s]))` + regexp.QuoteMeta(ternary.OperatorSecond) + `(?=['\\";\w\s\n])`)
}

type Keyword struct {
	Keyword string
}

func (keyWord *Keyword) getRegex() (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf("\b%v\b", keyWord.Keyword))
}

func NewChecker(userSolution solution.Solution, configurationFilePath string) *Checker {

	languageConfiguration := solution.GetConfiguration(configurationFilePath)

	return &Checker{
		LanguageInfo: LanguageInfo {
			Name: userSolution.Language,
			Extension: languageConfiguration.LangConfigs[userSolution.Language].SourceExtension,
			SingleLineComment: languageConfiguration.LangConfigs[userSolution.Language].Comments.SingleLineComment,
			MultiLineComment: languageConfiguration.LangConfigs[userSolution.Language].Comments.MultiLineComment,

		},
	}
}

func (checker *Checker) Check(userSolution solution.Solution, task solution.Task) {

	//var foundedElements []LanguageElement
	//cleanedCode := checker.prepareCode(userSolution.SourceCode)
	//
	//for index, value := range task.Options.Constructions {
	//
	//}

}

func (checker *Checker) prepareCode(code string) string {
	singleLineComment := checker.LanguageInfo.SingleLineComment
	multiLineComment := checker.LanguageInfo.MultiLineComment

	stringRegex := `(['"]).*?\\1`

	unifiedRegex, err := regexp.Compile(fmt.Sprintf("%v|%v|%v", singleLineComment,multiLineComment, stringRegex))
	if err != nil {
		log.Printf("Preparation error %v", err)
	}

	return unifiedRegex.ReplaceAllLiteralString(code, "")
}