package operators

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dlclark/regexp2"
)

type pattern string

func toPattern(s string) pattern {
	return pattern(regexp2.Escape(s))
}

func (pat pattern) NotFollowedBy(pats ...string) pattern {
	return pattern(fmt.Sprintf("%s(?!%s)", pat, or(pats)))
}

func (pat pattern) NotPrecededBy(pats ...string) pattern {
	return pattern(fmt.Sprintf("(?<!%s)%s", or(pats), pat))
}

func (pat pattern) or(other pattern) pattern {
	if pat == "" {
		return other
	}
	return pattern(fmt.Sprintf("%s|%s", pat, other))
}

func or(alternatives []string) pattern {
	sb := new(strings.Builder)
	for i, alt := range alternatives {
		if len(alt) > 1 {
			sb.WriteString("(?:")
		}
		sb.WriteString(regexp2.Escape(alt))
		if len(alt) > 1 {
			sb.WriteString(")")
		}
		if i < len(alternatives)-1 {
			sb.WriteString("|")
		}
	}
	return pattern(sb.String())
}

type languageSpec struct {
	inherits          string
	singleLineComment *regexp2.Regexp
	multiLineComment  *regexp2.Regexp
	opersKwsRegexes   map[string]pattern
}

type Checker struct {
	languages map[string]languageSpec
}

func NewChecker() *Checker {
	return &Checker{
		languages: map[string]languageSpec{
			"c": languageSpec{
				inherits:          "",
				singleLineComment: regexp2.MustCompile(`//.*$`, regexp2.RE2),
				multiLineComment:  regexp2.MustCompile(`/\*(?:[^*]|\*(?!\/))*(?:\*/)?`, regexp2.RE2),
				opersKwsRegexes: map[string]pattern{
					"-": toPattern("-").NotPrecededBy("-").NotFollowedBy("=", ">", "-"),
					"+": toPattern("+").NotPrecededBy("+").NotFollowedBy("=", "+"),
					">": toPattern(">").NotPrecededBy("-", ">", ":", "%").NotFollowedBy(">", "="),
					"<": toPattern("<").NotPrecededBy("<").NotFollowedBy("<", "=", ":", "%"),
					"=": toPattern("=").NotPrecededBy(
						"-", "+", ">", "<", "=", "!", "^", "&", "|", "*", "%", "/",
					).NotFollowedBy("="),
					"&":  toPattern("&").NotPrecededBy("&").NotFollowedBy("&", "="),
					"|":  toPattern("|").NotPrecededBy("|").NotFollowedBy("|", "="),
					":":  toPattern(":").NotPrecededBy("<", "%").NotFollowedBy(">"),
					"%":  toPattern("%").NotPrecededBy("<").NotFollowedBy("=", ":", ">"),
					"#":  toPattern("#").NotPrecededBy("#").NotFollowedBy("#"),
					".":  toPattern(".").NotPrecededBy("..").NotFollowedBy(".."),
					"!":  toPattern("!").NotFollowedBy("="),
					"*":  toPattern("*").NotFollowedBy("="),
					"/":  toPattern("/").NotFollowedBy("="),
					">>": toPattern(">>").NotFollowedBy("="),
					"<<": toPattern("<<").NotFollowedBy("="),
				},
			},
			"cpp": languageSpec{
				inherits:          "c",
				singleLineComment: nil,
				multiLineComment:  nil,
				opersKwsRegexes: map[string]pattern{
					".": toPattern(".").NotPrecededBy("..").NotFollowedBy("..", "*"),
					":": toPattern(":").NotPrecededBy("<", "%", ":").NotFollowedBy(">", ":"),
					">": toPattern(">").NotPrecededBy("-", ">", ":", "%").NotFollowedBy(">", "=", "*"),
					"*": toPattern("*").NotPrecededBy(">", ".").NotFollowedBy("="),
				},
			},
			"csharp": languageSpec{
				inherits:          "c",
				singleLineComment: nil,
				multiLineComment:  nil,
				opersKwsRegexes: map[string]pattern{
					"=": toPattern("=").NotPrecededBy(
						"-", "+", ">", "<", "=", "!", "^", "&", "|", "*", "%", "/",
					).NotFollowedBy("=", ">"),
					"?": toPattern("?").NotPrecededBy("?").NotFollowedBy("?", "."),
					":": toPattern(":").NotPrecededBy(":").NotFollowedBy(":"),
					".": toPattern(".").NotPrecededBy("?", "..").NotFollowedBy(".."),
				},
			},
			"java": languageSpec{
				inherits:          "c",
				singleLineComment: nil,
				multiLineComment:  nil,
				opersKwsRegexes:   map[string]pattern{},
			},
			"kotlin": languageSpec{
				inherits:          "c",
				singleLineComment: nil,
				multiLineComment:  nil,
				opersKwsRegexes: map[string]pattern{
					"?":   toPattern("?").NotFollowedBy("."),
					".":   toPattern(".").NotPrecededBy("..", "?").NotFollowedBy(".."),
					"!in": toPattern("!in"),
					"as?": toPattern("as?"),
				},
			},
			"python": languageSpec{
				inherits:          "c",
				singleLineComment: regexp2.MustCompile(`#.*$`, regexp2.RE2),
				multiLineComment:  nil,
				opersKwsRegexes: map[string]pattern{
					"/": toPattern("/").NotPrecededBy("/").NotFollowedBy("=", "/"),
					"*": toPattern("*").NotPrecededBy("*").NotFollowedBy("=", "*"),
					"@": toPattern("@").NotFollowedBy("="),
					"=": toPattern("=").NotPrecededBy(
						"-", "+", ">", "<", "=", "!", "^", "&", "|", "*", "%", "/", "@", ":",
					).NotFollowedBy("="),
					":": toPattern(":").NotFollowedBy("="),
				},
			},
			"go": languageSpec{
				inherits:          "c",
				singleLineComment: nil,
				multiLineComment:  nil,
				opersKwsRegexes: map[string]pattern{
					"=": toPattern("=").NotPrecededBy(
						"-", "+", ">", "<", "=", "!", "^", "&", "|", "*", "%", "/", ":",
					).NotFollowedBy("="),
					":": toPattern(":").NotFollowedBy("="),
					"<": toPattern("<").NotPrecededBy("<").NotFollowedBy("<", "=", "-"),
					"-": toPattern("-").NotPrecededBy("-", "<").NotFollowedBy("=", ">", "-"),
					"&": toPattern("&").NotPrecededBy("&").NotFollowedBy("&", "=", "^"),
					"^": toPattern("^").NotPrecededBy("&").NotFollowedBy("^", "="),
				},
			},
		},
	}
}

func (ch Checker) prepareCode(sourceCode string, lang string) (string, error) {
	singleLineComment, err := ch.getSingleLineComment(lang)
	multiLineComment, err := ch.getMultiLineComment(lang)

	stringRegex := `"(?:[^"\\]|\\"|\\)*"|'(?:[^'\\]|\\'|\\)*'`

	unifiedRegex, err := regexp2.Compile(fmt.Sprintf("%v|%v|%v", singleLineComment, multiLineComment, stringRegex), regexp2.RE2|regexp2.Multiline)
	if err != nil {
		log.Printf("Preparation error %v", err)
	}

	return unifiedRegex.Replace(sourceCode, "", -1, -1)
}

func (ch Checker) getOpersKwsPattern(key string, lang string) (pattern, error) {
	currectLangSpec := ch.languages[lang]
	var pat pattern
	var keyFound bool
	for {
		pat, keyFound = currectLangSpec.opersKwsRegexes[key]
		if !keyFound && currectLangSpec.inherits != "" {
			currectLangSpec = ch.languages[currectLangSpec.inherits]
		} else {
			break
		}
	}
	if keyFound {
		return pat, nil
	}
	return pat, errors.New("oper/kw pattern not found")
}

func (ch Checker) getSingleLineComment(lang string) (*regexp2.Regexp, error) {
	currectLangSpec := ch.languages[lang]
	var re *regexp2.Regexp
	for {
		re = currectLangSpec.singleLineComment
		if re == nil && currectLangSpec.inherits != "" {
			currectLangSpec = ch.languages[currectLangSpec.inherits]
		} else {
			break
		}
	}
	if re != nil {
		return re, nil
	}
	return re, errors.New("single line comment pattern not found")
}

func (ch Checker) getMultiLineComment(lang string) (*regexp2.Regexp, error) {
	currectLangSpec := ch.languages[lang]
	var re *regexp2.Regexp
	for {
		re = currectLangSpec.multiLineComment
		if re == nil && currectLangSpec.inherits != "" {
			currectLangSpec = ch.languages[currectLangSpec.inherits]
		} else {
			break
		}
	}
	if re != nil {
		return re, nil
	}
	return re, errors.New("multi line comment pattern not found")
}

func isKeyword(s string) (bool, error) {
	return regexp2.MustCompile(`\b[a-zA-Z_]+\b`, regexp2.RE2).MatchString(s)
}

func (ch Checker) Check(sourceCode string, constructions []string, lang string) bool {
	preparedCode, _ := ch.prepareCode(sourceCode, lang)
	wholePattern := pattern("")
	for _, constr := range constructions {
		constrPattern, err := ch.getOpersKwsPattern(constr, lang)
		if err != nil {
			if isKw, _ := isKeyword(constr); isKw {
				constrPattern = pattern(fmt.Sprintf(`\b%s\b`, toPattern(constr)))
			} else {
				constrPattern = toPattern(constr)
			}
		}
		wholePattern = wholePattern.or(constrPattern)
	}
	if wholePattern == "" {
		return false
	}
	wholeRe := regexp2.MustCompile(string(wholePattern), regexp2.RE2)
	result, _ := wholeRe.FindStringMatch(preparedCode)
	return result != nil
}
