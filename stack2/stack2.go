package stack2

import (
	"PoEPricer/item"
	"errors"
	//"fmt"
	"reflect"
	"strconv"
	"strings"
)

type tokenType int

const (
	type_BOOL tokenType = iota
	type_FLOAT
	type_STRING
	type_OPERATOR
	type_PARENTHESIS
)

type operator int

const (
	operator_HASAFFIX operator = iota
	operator_COULDHAVEAFFIX
	operator_MULTIPLY
	operator_DIVIDE
	operator_UNARY_PLUS
	operator_UNARY_MINUS
	operator_MAXIMUM
	operator_ADD
	operator_SUBTRACT
	operator_LESS
	operator_LESSEQUAL
	operator_GREATER
	operator_GREATEREQUAL
	operator_EQUAL
	operator_NOTEQUAL
	operator_AND
	operator_OR
	operator_NOT
)

type token struct {
	tokenType   tokenType
	isVariable  bool
	index       int
	boolValue   bool
	floatValue  float64
	stringValue string
	operator    operator
}

type expression struct {
	name       string //Set if the expression belongs to a function
	resultType tokenType
	tokens     []*token
}

type stack struct {
	content []*token
}

type Filter struct {
	floatIndices                       map[string]int
	boolIndices                        map[string]int
	stringIndices                      map[string]int
	floatIndex, boolIndex, stringIndex int
	functions                          []*expression
	parentCondition                    *condition
}

type condition struct {
	SubConditions []*condition
	lines         []string
	expression    *expression
	Warning       string
}

var globalFloatIndices map[string]int
var globalBoolIndices map[string]int
var globalStringIndices map[string]int

func peekString(str string, pos int) string {
	if len(str) > pos {
		return string(str[pos])
	} else {
		return ""
	}
}

func isAlphabetic(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNumber(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isNumeric(ch byte) bool {
	return isNumber(ch) || ch == '.'
}

func (tok *token) String() string {
	/*tokenType   tokenType
	isVariable  bool
	index       int
	boolValue   bool
	floatValue  float64
	stringValue string
	operator    operator*/
	switch tok.tokenType {
	case type_BOOL:
		return strconv.FormatBool(tok.boolValue)
	case type_FLOAT:
		return strconv.FormatFloat(tok.floatValue, 'f', 2, 64)
	case type_STRING:
		return "\"" + tok.stringValue + "\""
	case type_OPERATOR:
		switch tok.operator {
		case operator_MULTIPLY:
			return "*"
		case operator_DIVIDE:
			return "*"
		case operator_ADD:
			return "+"
		case operator_SUBTRACT:
			return "-"
		case operator_UNARY_PLUS:
			return "U+"
		case operator_UNARY_MINUS:
			return "U-"
		case operator_LESS:
			return "<"
		case operator_LESSEQUAL:
			return "<="
		case operator_GREATER:
			return ">"
		case operator_GREATEREQUAL:
			return ">="
		case operator_EQUAL:
			return "=="
		case operator_NOTEQUAL:
			return "!="
		case operator_AND:
			return "&&"
		case operator_OR:
			return "||"
		case operator_NOT:
			return "!"
		default:
			return "Unknown_Operator"
		}
	default:
		return "Unknown_Type"
	}
}

func tokenize(expression string) ([]string, error) {
	list := make([]string, 0)
	for i := 0; i < len(expression); i++ {
		switch {
		case expression[i] == '"':
			word := ""
			for {
				i++
				next := peekString(expression, i)
				if len(next) == 0 {
					return nil, errors.New("String not terminated")
				} else if next == "\\" {
					peek := peekString(expression, i+1)
					if peek == "\"" || peek == "\"" {
						word += string(expression[i+1])
						i++
					} else {
						return nil, errors.New("Unknown terminated character: \"\\" + peek + "\"")
					}
				} else if next == "\"" {
					list = append(list, "\""+word+"\"")
					break
				} else {
					word += next
				}
			}
		case expression[i] == '(' || expression[i] == ')' || expression[i] == '*' ||
			expression[i] == '/' || expression[i] == '+' || expression[i] == '-':
			list = append(list, string(expression[i]))
		case expression[i] == '|' || expression[i] == '&' || expression[i] == '=':
			if peekString(expression, i+1) == string(expression[i]) {
				list = append(list, string(expression[i])+string(expression[i]))
				i++
			} else {
				return nil, errors.New("Invalid single character \"" + string(expression[i]) + "\".")
			}
		case expression[i] == '<' || expression[i] == '>' || expression[i] == '!':
			if peekString(expression, i+1) == "=" {
				list = append(list, string(expression[i])+"=")
				i++
			} else {
				list = append(list, string(expression[i]))
			}

		case isAlphabetic(expression[i]): //Variable
			word := string(expression[i])
			for {
				next := peekString(expression, i+1)
				if len(next) == 1 && (isAlphabetic(next[0]) || isNumber(next[0])) {
					i++
					word += next
				} else {
					list = append(list, word)
					break
				}
			}
		case isNumeric(expression[i]):
			word := string(expression[i])
			for {
				next := peekString(expression, i+1)
				if len(next) == 1 && isNumeric(next[0]) {
					i++
					word += next
				} else {
					list = append(list, word)
					break
				}
			}
		case expression[i] == '$' || expression[i] == ' ':
			//do nothing
		}
	}
	return list, nil
}

func stripComment(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] == '"' { //String -> ignore all until end of string
			for {
				i++
				next := peekString(line, i)
				if next == "\\" { //Terminated character, ignore next
					i++
				} else if next == "\"" {
					break
				}
			}
		} else if line[i] == '#' {
			return line[:i]
		}
	}
	return line
}

func isEmptyLine(line string) bool {
	for i := 0; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			return false
		}
	}
	return true
}

func newFilter() *Filter {
	result := new(Filter)
	result.boolIndices = make(map[string]int)
	result.floatIndices = make(map[string]int)
	result.stringIndices = make(map[string]int)
	for key, value := range globalBoolIndices {
		result.boolIndices[key] = value
	}
	for key, value := range globalFloatIndices {
		result.floatIndices[key] = value
	}
	for key, value := range globalStringIndices {
		result.stringIndices[key] = value
	}
	result.boolIndex = len(globalBoolIndices)
	result.floatIndex = len(globalFloatIndices)
	result.stringIndex = len(globalStringIndices)
	return result
}

func operatorStringToToken(str string) *token {
	token := &token{tokenType: type_OPERATOR}
	switch str {
	case "*":
		token.operator = operator_MULTIPLY
	case "/":
		token.operator = operator_DIVIDE
	case "+":
		token.operator = operator_ADD
	case "-":
		token.operator = operator_SUBTRACT
	case "<":
		token.operator = operator_LESS
	case "<=":
		token.operator = operator_LESSEQUAL
	case ">":
		token.operator = operator_GREATER
	case ">=":
		token.operator = operator_GREATEREQUAL
	case "==":
		token.operator = operator_EQUAL
	case "!=":
		token.operator = operator_NOTEQUAL
	case "&&":
		token.operator = operator_AND
	case "||":
		token.operator = operator_OR
	case "!":
		token.operator = operator_NOT
	default:
		panic("Call of operatorStringToToken with invalid string. This should never have happened.")
	}
	return token
}

func (t1 *token) precedence(t2 *token) bool {
	if t2 == nil {
		return true
	} else if t2.tokenType == type_PARENTHESIS {
		return true
	} else {
		return t1.operator < t2.operator
	}
}

func newStack() *stack {
	result := new(stack)
	result.content = make([]*token, 0)
	return result
}

func (s *stack) push(elem *token) {
	s.content = append(s.content, elem)
}

func (s *stack) pop() *token {
	if s.empty() {
		return nil
	}
	result := s.content[len(s.content)-1]
	s.content = s.content[:len(s.content)-1]
	return result
}

func (s *stack) peek() *token {
	if s.empty() {
		return nil
	} else {
		return s.content[len(s.content)-1]
	}
}

func (s *stack) empty() bool {
	return len(s.content) == 0
}

func (exp *expression) setReturnType() error {
	stack := newStack()
	/*for i := 0; i < len(exp.tokens); i++ {
		t := exp.tokens[i]
		if t.tokenType == type_BOOL {
			fmt.Print("Bool ")
		} else if t.tokenType == type_FLOAT {
			fmt.Print("Float ")
		} else if t.tokenType == type_STRING {
			fmt.Print("String ")
		} else if t.tokenType == type_OPERATOR {
			fmt.Print("Operator_", t.operator, " ")
		} else {
			fmt.Print("Unknown ")
		}
	}
	fmt.Println()*/
	for i := 0; i < len(exp.tokens); i++ {
		tok := exp.tokens[i]
		if tok.tokenType == type_OPERATOR {
			if tok.operator == operator_NOT || tok.operator == operator_UNARY_MINUS || tok.operator == operator_UNARY_PLUS ||
				tok.operator == operator_HASAFFIX || tok.operator == operator_COULDHAVEAFFIX {
				val := stack.pop()
				if val == nil {
					return errors.New("Operator ! without an argument.")
				}
				switch tok.operator {
				case operator_NOT:
					if val.tokenType == type_BOOL {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("! operator needs a boolean value.")
					}
				case operator_UNARY_MINUS, operator_UNARY_PLUS:
					if val.tokenType == type_FLOAT {
						stack.push(&token{tokenType: type_FLOAT})
					} else {
						return errors.New("+ and - operators needs a float value.")
					}
				case operator_HASAFFIX, operator_COULDHAVEAFFIX:
					if val.tokenType == type_STRING {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("Affix evaluation needs a string argument.")
					}
				}
			} else {
				right := stack.pop()
				left := stack.pop()
				if right == nil || left == nil {
					return errors.New("Operator has insufficient arguments.")
				}
				switch tok.operator {
				case operator_MULTIPLY, operator_DIVIDE, operator_ADD, operator_SUBTRACT:
					if right.tokenType == type_FLOAT && left.tokenType == type_FLOAT {
						stack.push(&token{tokenType: type_FLOAT})
					} else {
						return errors.New("+,-,*,/ operators need two floating point values.")
					}
				case operator_MAXIMUM:
					if right.tokenType == type_FLOAT && left.tokenType == type_FLOAT {
						stack.push(&token{tokenType: type_FLOAT})
					} else {
						return errors.New("Maximum operator needs two floating point values.")
					}
				case operator_LESS, operator_LESSEQUAL, operator_GREATER, operator_GREATEREQUAL:
					if right.tokenType == type_FLOAT && left.tokenType == type_FLOAT {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("<,<=,>,>= operators need two floating point values.")
					}
				case operator_AND, operator_OR:
					if right.tokenType == type_BOOL && left.tokenType == type_BOOL {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("&& and || operators need two boolean values.")
					}
				case operator_EQUAL, operator_NOTEQUAL:
					if right.tokenType == type_BOOL && left.tokenType == type_BOOL {
						stack.push(&token{tokenType: type_BOOL})
					} else if right.tokenType == type_FLOAT && left.tokenType == type_FLOAT {
						stack.push(&token{tokenType: type_BOOL})
					} else if right.tokenType == type_STRING && left.tokenType == type_STRING {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("Comparison between different types (float, boolean, string) is not allowed.")
					}
				}
			}

		} else {
			stack.push(tok)
		}
	}
	result := stack.pop()
	if result != nil && (result.tokenType == type_BOOL || result.tokenType == type_FLOAT || result.tokenType == type_STRING) && stack.empty() {
		exp.resultType = result.tokenType
		return nil
	} else {
		return errors.New("Expression must compute to a single value")
	}
}

func (filter *Filter) compileExpression(exp string) (*expression, error) {
	list, err := tokenize(exp)
	if err != nil {
		return nil, err
	}
	output := make([]*token, 0)
	operators := newStack()
	nextIsUnary := true
	for i := 0; i < len(list); i++ {
		switch list[i] {
		case "HasAffix":
			token := &token{tokenType: type_OPERATOR, operator: operator_HASAFFIX}
			operators.push(token)
		case "CouldHaveAffix":
			token := &token{tokenType: type_OPERATOR, operator: operator_COULDHAVEAFFIX}
			operators.push(token)
		case "GetMaximum":
			token := &token{tokenType: type_OPERATOR, operator: operator_MAXIMUM}
			operators.push(token)
		case "+", "-":
			if nextIsUnary {
				token := &token{tokenType: type_OPERATOR}
				if list[i] == "+" {
					token.operator = operator_UNARY_PLUS
				} else {
					token.operator = operator_UNARY_MINUS
				}
				operators.push(token)
				continue
			}
			fallthrough
		case "*", "/", "<", "<=", ">", ">=", "==", "!=", "&&", "||", "!":
			token := operatorStringToToken(list[i])
			for !token.precedence(operators.peek()) {
				output = append(output, operators.pop())
			}
			operators.push(token)
			nextIsUnary = true
		case "(":
			token := &token{tokenType: type_PARENTHESIS}
			operators.push(token)
			nextIsUnary = true
		case ")":
			for {
				if operators.empty() {
					return nil, errors.New("Mismatched parenthesis, missing a \"(\".")
				}
				token := operators.pop()
				if token.tokenType == type_PARENTHESIS {
					break
				} else {
					output = append(output, token)
				}
			}
			nextIsUnary = false
		default: //Either a string, a number, or a variable/method
			nextIsUnary = false
			f, err := strconv.ParseFloat(list[i], 64)
			if err == nil {
				token := &token{tokenType: type_FLOAT, floatValue: f}
				output = append(output, token)
			} else if strings.HasPrefix(list[i], "\"") && strings.HasSuffix(list[i], "\"") {
				line := list[i]
				token := &token{tokenType: type_STRING, stringValue: line[1 : len(line)-1]}
				output = append(output, token)
			} else {
				//Var, Map or Method
				sanitized := list[i]
				if strings.HasPrefix(sanitized, "$") && strings.HasSuffix(sanitized, "$") && len(sanitized) >= 2 {
					sanitized = sanitized[1 : len(sanitized)-1]
				}
				index, ok := filter.boolIndices[sanitized]
				if ok {
					token := &token{tokenType: type_BOOL, isVariable: true, index: index}
					output = append(output, token)
				} else {
					index, ok := filter.floatIndices[sanitized]
					if ok {
						token := &token{tokenType: type_FLOAT, isVariable: true, index: index}
						output = append(output, token)
					} else {
						index, ok := filter.stringIndices[sanitized]
						if ok {
							token := &token{tokenType: type_STRING, isVariable: true, index: index}
							output = append(output, token)
						} else {
							return nil, errors.New("\"" + list[i] + "\" is not a valid identifier.")
						}
					}
				}
			}
		}
	}
	for !operators.empty() {
		token := operators.pop()
		if token.tokenType == type_PARENTHESIS {
			return nil, errors.New("Mismatched parenthesis, missing a \")\".")
		} else {
			output = append(output, token)
		}
	}
	var result expression
	result.tokens = output
	err = result.setReturnType()
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (filter *Filter) nameAlreadyUsed(name string) bool {
	_, ok1 := filter.boolIndices[name]
	_, ok2 := filter.floatIndices[name]
	_, ok3 := filter.stringIndices[name]
	return ok1 || ok2 || ok3
}

func (filter *Filter) addFunction(line string) error {
	content := strings.SplitN(line[9:], " ", 2)
	if len(content) != 2 {
		return errors.New("Invalid Function.")
	}
	if strings.HasPrefix(content[0], "$") && strings.HasSuffix(content[0], "$") && len(content[0]) >= 2 {
		content[0] = content[0][1 : len(content[0])-1] //Legacy version where variables where enclosed
	}
	if filter.nameAlreadyUsed(content[0]) {
		return errors.New("Function has a duplicate name.")
	}
	exp, err := filter.compileExpression(content[1])
	if err != nil {
		return err
	}
	if exp.resultType == type_BOOL {
		filter.boolIndices[content[0]] = filter.boolIndex
		filter.boolIndex++
	} else if exp.resultType == type_FLOAT {
		filter.floatIndices[content[0]] = filter.floatIndex
		filter.floatIndex++
	} else if exp.resultType == type_STRING {
		filter.stringIndices[content[0]] = filter.stringIndex
		filter.stringIndex++
	}
	exp.name = content[0]
	filter.functions = append(filter.functions, exp)
	return nil
}

func newCondition() *condition {
	result := new(condition)
	result.lines = make([]string, 0)
	result.SubConditions = make([]*condition, 0)
	return result
}

func (filter *Filter) passLine(parent *condition, line string) error {
	if strings.HasPrefix(line, "Condition ") {
		if parent.Warning != "" {
			return errors.New("Unexpected Subcondition, there is already a warning defined.")
		}
		child := newCondition()
		//content := strings.SplitN(line[10:], " ", 2)
		line = line[10:]
		exp, err := filter.compileExpression(line)
		if err != nil {
			return err
		}
		if exp.resultType != type_BOOL {
			return errors.New("Conditions must compute to a boolean value.")
		}
		child.expression = exp
		parent.SubConditions = append(parent.SubConditions, child)
	} else if strings.HasPrefix(line, "Warn ") {
		if len(parent.SubConditions) == 0 {
			return errors.New("Warn unexpectedly idented.")
		} else {
			warnedCondition := parent.SubConditions[len(parent.SubConditions)-1]
			if warnedCondition.Warning != "" {
				return errors.New("Unexpected Warn, there is already a warning defined for the current condition.")
			} else if len(warnedCondition.SubConditions) > 0 {
				return errors.New("Unexpected Warn, condition already has subconditions.")
			} else {
				warnedCondition.Warning = line[5:]
			}
		}
	} else if strings.HasPrefix(line, "\t") {
		if len(parent.SubConditions) == 0 {
			return errors.New("Invalid line. Indented lines must have a parent condition.")
		} else {
			return filter.passLine(parent.SubConditions[len(parent.SubConditions)-1], line[1:])
		}
	} else {
		return errors.New("Invalid line.")
	}
	return nil
}

func Compile(filterText string) (*Filter, error) {
	result := newFilter()
	filterText = strings.Replace(filterText, "\r", "", -1)
	lines := strings.Split(filterText, "\n")
	parentCondition := newCondition()
	for i := 0; i < len(lines); i++ {
		line := stripComment(lines[i])
		if isEmptyLine(line) {
			continue
		}
		if strings.HasPrefix(line, "Function ") {
			err := result.addFunction(line)
			if err != nil {
				return nil, errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": " + err.Error())
			}
		} else {
			err := result.passLine(parentCondition, line)
			if err != nil {
				return nil, errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": " + err.Error())
			}
		}
	}
	result.parentCondition = parentCondition
	/*	list, err := tokenize(filterText) //list now conatins a slice, where each element is a token in string form
		if err != nil {
			return nil, err
		}*/
	return result, nil
}

func applyTwoArgumentOperator(operator, right, left *token) *token {
	switch operator.operator {
	case operator_MULTIPLY:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue * right.floatValue}
	case operator_DIVIDE:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue / right.floatValue}
	case operator_ADD:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue + right.floatValue}
	case operator_SUBTRACT:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue - right.floatValue}
	case operator_MAXIMUM:
		if left.floatValue > right.floatValue {
			return left
		} else {
			return right
		}
	case operator_AND:
		return &token{tokenType: type_BOOL, boolValue: left.boolValue && right.boolValue}
	case operator_OR:
		return &token{tokenType: type_BOOL, boolValue: left.boolValue || right.boolValue}
	case operator_LESS:
		return &token{tokenType: type_BOOL, boolValue: left.floatValue < right.floatValue}
	case operator_LESSEQUAL:
		return &token{tokenType: type_BOOL, boolValue: left.floatValue <= right.floatValue}
	case operator_GREATER:
		return &token{tokenType: type_BOOL, boolValue: left.floatValue > right.floatValue}
	case operator_GREATEREQUAL:
		return &token{tokenType: type_BOOL, boolValue: left.floatValue >= right.floatValue}
	case operator_EQUAL:
		if right.tokenType == type_FLOAT {
			return &token{tokenType: type_BOOL, boolValue: left.floatValue == right.floatValue}
		} else if right.tokenType == type_BOOL {
			return &token{tokenType: type_BOOL, boolValue: left.boolValue == right.boolValue}
		} else {
			return &token{tokenType: type_BOOL, boolValue: left.stringValue == right.stringValue}
		}
	case operator_NOTEQUAL:
		if right.tokenType == type_FLOAT {
			return &token{tokenType: type_BOOL, boolValue: left.floatValue != right.floatValue}
		} else if right.tokenType == type_BOOL {
			return &token{tokenType: type_BOOL, boolValue: left.boolValue != right.boolValue}
		} else {
			return &token{tokenType: type_BOOL, boolValue: left.stringValue != right.stringValue}
		}
	default:
		panic("Given operator not implemented. This should not have happened.")
	}
}

func applySingleArgumentOperator(operator, arg *token, possibleAffixes []item.AffixList) *token {
	switch operator.operator {
	case operator_NOT:
		arg.boolValue = !arg.boolValue
		return arg
	case operator_UNARY_MINUS:
		arg.floatValue = -arg.floatValue
		return arg
	case operator_UNARY_PLUS:
		return arg
	case operator_HASAFFIX:
		if len(possibleAffixes) == 0 {
			return &token{tokenType: type_BOOL, boolValue: false}
		}
		for i := 0; i < len(possibleAffixes); i++ {
			found := false
			for j := 0; j < len(possibleAffixes[i]); j++ {
				affix := possibleAffixes[i][j]
				if affix.Name == arg.stringValue {
					found = true
					break
				}
			}
			if !found {
				return &token{tokenType: type_BOOL, boolValue: false}
			}
		}
		return &token{tokenType: type_BOOL, boolValue: true}
	case operator_COULDHAVEAFFIX:
		for i := 0; i < len(possibleAffixes); i++ {
			for j := 0; j < len(possibleAffixes[i]); j++ {
				affix := possibleAffixes[i][j]
				if affix.Name == arg.stringValue {
					return &token{tokenType: type_BOOL, boolValue: true}
				}
			}
		}
		return &token{tokenType: type_BOOL, boolValue: false}
	default:
		panic("Given operator not implemented. This should not have happened.")
	}
}

func (exp *expression) evaluate(boolValues []bool, floatValues []float64, stringValues []string, affixPossibilities []item.AffixList) *token {
	for i := 0; i < len(exp.tokens); i++ {
		tok := exp.tokens[i]
		if tok.isVariable {
			if tok.tokenType == type_BOOL {
				tok.boolValue = boolValues[tok.index]
			} else if tok.tokenType == type_FLOAT {
				tok.floatValue = floatValues[tok.index]
			} else if tok.tokenType == type_STRING {
				tok.stringValue = stringValues[tok.index]
			}
		}
	}
	//fmt.Println(exp.tokens)
	stack := newStack()
	for i := 0; i < len(exp.tokens); i++ {
		tok := exp.tokens[i]
		if tok.tokenType == type_OPERATOR {
			switch tok.operator {
			case operator_NOT, operator_UNARY_MINUS, operator_UNARY_PLUS, operator_HASAFFIX, operator_COULDHAVEAFFIX:
				stack.push(applySingleArgumentOperator(tok, stack.pop(), affixPossibilities))
			default:
				stack.push(applyTwoArgumentOperator(tok, stack.pop(), stack.pop()))
			}
		} else {
			stack.push(tok)
		}
	}
	return stack.pop()
}

func (cond *condition) getWarning(boolValues []bool, floatValues []float64, stringValues []string, affixPossibilities []item.AffixList) string {
	if cond.expression == nil || cond.expression.evaluate(boolValues, floatValues, stringValues, affixPossibilities).boolValue {
		if cond.Warning != "" {
			return cond.Warning
		} else {
			for i := 0; i < len(cond.SubConditions); i++ {
				subWarn := cond.SubConditions[i].getWarning(boolValues, floatValues, stringValues, affixPossibilities)
				if subWarn != "" {
					return subWarn
				}
			}
		}
	}
	return ""
}

func (filter *Filter) Execute(it *item.Item) string {
	boolValues := make([]bool, filter.boolIndex)
	floatValues := make([]float64, filter.floatIndex)
	stringValues := make([]string, filter.stringIndex)
	boolIndex := 0
	floatIndex := 0
	stringIndex := 0
	//***Plug in all the data from an item***
	itemValue := reflect.ValueOf(*it)
	numFields := itemValue.Type().NumField()
	//Variables
	for i := 0; i < numFields; i++ {
		if itemValue.Type().Field(i).Type.Name() == "bool" {
			boolValues[boolIndex] = itemValue.Field(i).Bool()
			boolIndex++
		} else if itemValue.Type().Field(i).Type.Name() == "float64" {
			floatValues[floatIndex] = itemValue.Field(i).Float()
			floatIndex++
		} else if itemValue.Type().Field(i).Type.Name() == "string" {
			stringValues[stringIndex] = itemValue.Field(i).String()
			stringIndex++
		}
	}
	//Methods
	itemValue = reflect.ValueOf(it)
	numMethods := itemValue.Type().NumMethod()
	for i := 0; i < numMethods; i++ {
		if itemValue.Type().Method(i).Type.String() == "func(*item.Item) bool" {
			results := itemValue.Method(i).Call(nil)
			boolValues[boolIndex] = results[0].Bool()
			boolIndex++
		} else if itemValue.Type().Method(i).Type.String() == "func(*item.Item) float64" {
			results := itemValue.Method(i).Call(nil)
			floatValues[floatIndex] = results[0].Float()
			floatIndex++
		} else if itemValue.Type().Method(i).Type.String() == "func(*item.Item) string" {
			results := itemValue.Method(i).Call(nil)
			stringValues[stringIndex] = results[0].String()
			stringIndex++
		}
	}
	//MapKeys
	for _, description := range item.ItemBoolMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		index, _ := globalBoolIndices[description]
		boolValues[index] = it.GetBoolModValue(description)
	}
	for _, description := range item.ItemFloatMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		index, _ := globalFloatIndices[description]
		floatValues[index] = it.GetFloatModValue(description)
	}
	// Functions
	for i := 0; i < len(filter.functions); i++ {
		f := filter.functions[i]
		tok := f.evaluate(boolValues, floatValues, stringValues, it.PossibleAffixes)
		//fmt.Println("Function: ", f.name, ", Values: ", tok.boolValue, " / ", tok.floatValue, " / ", tok.stringValue)
		if tok.tokenType == type_BOOL {
			index, _ := filter.boolIndices[f.name]
			boolValues[index] = tok.boolValue
		} else if tok.tokenType == type_FLOAT {
			index, _ := filter.floatIndices[f.name]
			floatValues[index] = tok.floatValue
		} else if tok.tokenType == type_STRING {
			index, _ := filter.stringIndices[f.name]
			stringValues[index] = tok.stringValue
		} else {
			panic("Unknown token type returned from function.")
		}
	}
	return filter.parentCondition.getWarning(boolValues, floatValues, stringValues, it.PossibleAffixes)

}

func init() {
	globalBoolIndices = make(map[string]int)
	globalFloatIndices = make(map[string]int)
	globalStringIndices = make(map[string]int)
	//Setup Methods and Variables of item.Item
	var it item.Item
	v := reflect.ValueOf(it)
	//Variables
	numFields := v.Type().NumField()
	for i := 0; i < numFields; i++ {
		if v.Type().Field(i).Type.Name() == "bool" {
			globalBoolIndices[v.Type().Field(i).Name] = len(globalBoolIndices)
		} else if v.Type().Field(i).Type.Name() == "float64" {
			globalFloatIndices[v.Type().Field(i).Name] = len(globalFloatIndices)
		} else if v.Type().Field(i).Type.Name() == "string" {
			globalStringIndices[v.Type().Field(i).Name] = len(globalStringIndices)
		}
	}
	//Methods
	v = reflect.ValueOf(&it)
	numMethods := v.Type().NumMethod()
	for i := 0; i < numMethods; i++ {
		if v.Type().Method(i).Type.String() == "func(*item.Item) bool" {
			globalBoolIndices[v.Type().Method(i).Name] = len(globalBoolIndices)
		} else if v.Type().Method(i).Type.String() == "func(*item.Item) float64" {
			globalFloatIndices[v.Type().Method(i).Name] = len(globalFloatIndices)
		} else if v.Type().Method(i).Type.String() == "func(*item.Item) string" {
			globalStringIndices[v.Type().Method(i).Name] = len(globalStringIndices)
		}
	}
	//MapKeys
	for _, key := range item.ItemBoolMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		globalBoolIndices[key] = len(globalBoolIndices)
	}
	for _, key := range item.ItemFloatMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		globalFloatIndices[key] = len(globalFloatIndices)
	}
}
