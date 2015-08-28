package stack

import (
	"PoEPricer/item"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type tokenType int

type Expression struct {
	tokens []*token
}

const (
	type_BOOL tokenType = iota
	type_VARBOOL
	type_MAPBOOL
	type_METHODBOOL
	type_FLOAT
	type_VARFLOAT
	type_MAPFLOAT
	type_METHODFLOAT
	type_STRING
	type_VARSTRING
	type_OPERATOR
	type_PARENTHESIS
)

type operator int

const (
	operator_MULTIPLY operator = iota
	operator_DIVIDE
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
	index       int
	mapKey      string
	boolValue   bool
	floatValue  float64
	stringValue string
	operator    operator
}

type stack struct {
	content []*token
}

type indexTypePair struct {
	tokenType tokenType
	mapKey    string
	index     int
}

//Maps the Names of all Methods and Variables to type_VARBOOL, type_METHODBOOL, type_VARFLOAT, type_MAPFLOAT, type_METHODFLOAT
var values map[string]indexTypePair

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

func addToList(list []string, add string) []string {
	if len(add) > 0 {
		return append(list, add)
	}
	return list
}

func peekString(str string, pos int) string {
	if len(str) > pos {
		return string(str[pos])
	} else {
		return ""
	}
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

func printTokens(list []*token) {
	for i := 0; i < len(list); i++ {
		if list[i].tokenType == type_OPERATOR {
			fmt.Print(list[i].operator, " ")
		} else {
			if list[i].tokenType == type_BOOL {
				fmt.Print(list[i].boolValue, " ")
			} else {
				fmt.Print(list[i].floatValue, " ")
			}

		}
	}
	fmt.Println()
}

func applyOperator(operator, right, left *token) *token {
	switch operator.operator {
	case operator_MULTIPLY:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue * right.floatValue}
	case operator_DIVIDE:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue / right.floatValue}
	case operator_ADD:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue + right.floatValue}
	case operator_SUBTRACT:
		return &token{tokenType: type_FLOAT, floatValue: left.floatValue - right.floatValue}
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
		if tokenTypeIsFloat(right.tokenType) {
			return &token{tokenType: type_BOOL, boolValue: left.floatValue == right.floatValue}
		} else if tokenTypeIsBool(right.tokenType) {
			return &token{tokenType: type_BOOL, boolValue: left.boolValue == right.boolValue}
		} else {
			return &token{tokenType: type_BOOL, boolValue: left.stringValue == right.stringValue}
		}
	case operator_NOTEQUAL:
		if tokenTypeIsFloat(right.tokenType) {
			return &token{tokenType: type_BOOL, boolValue: left.floatValue != right.floatValue}
		} else if tokenTypeIsBool(right.tokenType) {
			return &token{tokenType: type_BOOL, boolValue: left.boolValue != right.boolValue}
		} else {
			return &token{tokenType: type_BOOL, boolValue: left.stringValue != right.stringValue}
		}
	default:
		panic("Given operator not implemented. This should not have happened.")
	}
}

func (exp *Expression) Execute(item *item.Item) bool {
	itemValue := reflect.ValueOf(*item)
	itemPointerValue := reflect.ValueOf(item)
	//Put in the appropriate Values
	for i := 0; i < len(exp.tokens); i++ {
		if exp.tokens[i].tokenType == type_VARBOOL {
			exp.tokens[i].boolValue = itemValue.Field(exp.tokens[i].index).Bool()
		} else if exp.tokens[i].tokenType == type_VARFLOAT {
			exp.tokens[i].floatValue = itemValue.Field(exp.tokens[i].index).Float()
		} else if exp.tokens[i].tokenType == type_MAPBOOL {
			exp.tokens[i].boolValue = item.GetBoolModValue(exp.tokens[i].mapKey)
		} else if exp.tokens[i].tokenType == type_MAPFLOAT {
			exp.tokens[i].floatValue = item.GetFloatModValue(exp.tokens[i].mapKey)
		} else if exp.tokens[i].tokenType == type_METHODBOOL {
			results := itemPointerValue.Method(exp.tokens[i].index).Call(nil)
			exp.tokens[i].boolValue = results[0].Bool()
		} else if exp.tokens[i].tokenType == type_METHODFLOAT {
			results := itemPointerValue.Method(exp.tokens[i].index).Call(nil)
			exp.tokens[i].floatValue = results[0].Float()
		} else if exp.tokens[i].tokenType == type_VARSTRING {
			exp.tokens[i].stringValue = itemValue.Field(exp.tokens[i].index).String()
		}
	}
	//Execute with real numbers
	stack := newStack()
	for i := 0; i < len(exp.tokens); i++ {
		tok := exp.tokens[i]
		if tok.tokenType == type_OPERATOR {
			if tok.operator == operator_NOT {
				stack.push(&token{tokenType: type_BOOL, boolValue: !stack.pop().boolValue})
			} else {
				stack.push(applyOperator(tok, stack.pop(), stack.pop()))
			}
		} else {
			stack.push(tok)
		}
	}
	/*fmt.Print("Executing: ")
	printTokens(exp.tokens)
	fmt.Print("Result: ")
	printTokens(stack.content)*/
	return stack.pop().boolValue
}

func tokenTypeIsFloat(t tokenType) bool {
	return t == type_FLOAT || t == type_VARFLOAT || t == type_MAPFLOAT || t == type_METHODFLOAT
}

func tokenTypeIsBool(t tokenType) bool {
	return t == type_BOOL || t == type_VARBOOL || t == type_METHODBOOL || t == type_MAPBOOL
}

func tokenTypeIsString(t tokenType) bool {
	return t == type_STRING || t == type_VARSTRING
}

func (exp *Expression) check() error {
	stack := newStack()
	for i := 0; i < len(exp.tokens); i++ {
		tok := exp.tokens[i]
		if tok.tokenType == type_OPERATOR {
			if tok.operator == operator_NOT { //special case, needs only one argument
				val := stack.pop()
				if val == nil {
					return errors.New("Operator ! without an argument.")
				} else if val.tokenType == type_BOOL || val.tokenType == type_VARBOOL || val.tokenType == type_METHODBOOL {
					stack.push(&token{tokenType: type_BOOL})
				} else {
					return errors.New("! operator needs a boolean value.")
				}
			} else {
				right := stack.pop()
				left := stack.pop()
				if right == nil || left == nil {
					return errors.New("Operator has insufficient arguments.")
				}
				switch tok.operator {
				case operator_MULTIPLY, operator_DIVIDE, operator_ADD, operator_SUBTRACT:
					if tokenTypeIsFloat(right.tokenType) && tokenTypeIsFloat(left.tokenType) {
						stack.push(&token{tokenType: type_FLOAT})
					} else {
						return errors.New("+,-,*,/ operators need two floating point values.")
					}
				case operator_LESS, operator_LESSEQUAL, operator_GREATER, operator_GREATEREQUAL:
					if tokenTypeIsFloat(right.tokenType) && tokenTypeIsFloat(left.tokenType) {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("<,<=,>,>= operators need two floating point values.")
					}
				case operator_AND, operator_OR:
					if tokenTypeIsBool(right.tokenType) && tokenTypeIsBool(left.tokenType) {
						stack.push(&token{tokenType: type_BOOL})
					} else {
						return errors.New("&& and || operators need two boolean values.")
					}
				case operator_EQUAL, operator_NOTEQUAL:
					if tokenTypeIsBool(right.tokenType) && tokenTypeIsBool(left.tokenType) {
						stack.push(&token{tokenType: type_BOOL})
					} else if tokenTypeIsFloat(right.tokenType) && tokenTypeIsFloat(left.tokenType) {
						stack.push(&token{tokenType: type_BOOL})
					} else if tokenTypeIsString(right.tokenType) && tokenTypeIsString(left.tokenType) {
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
	if result != nil && (result.tokenType == type_BOOL || result.tokenType == type_VARBOOL || result.tokenType == type_METHODBOOL) && stack.empty() {
		return nil
	} else {
		return errors.New("Expression must compute to a single boolean value")
	}
}

func shuntingYard(list []string) (*Expression, error) {
	output := make([]*token, 0)
	operators := newStack()
	for i := 0; i < len(list); i++ {
		switch list[i] {
		case "*", "/", "+", "-", "<", "<=", ">", ">=", "==", "!=", "&&", "||", "!":
			token := operatorStringToToken(list[i])
			for !token.precedence(operators.peek()) {
				output = append(output, operators.pop())
			}
			operators.push(token)
		case "(":
			token := &token{tokenType: type_PARENTHESIS}
			operators.push(token)
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
		default: //Either a string, a number, or a variable/method
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
				indexTypePair, ok := values[sanitized]
				if !ok {
					return nil, errors.New("\"" + list[i] + "\" is not a valid identifier.")
				} else {
					token := &token{tokenType: indexTypePair.tokenType, index: indexTypePair.index, mapKey: indexTypePair.mapKey}
					output = append(output, token)
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
	var exp Expression
	exp.tokens = output
	return &exp, nil
}

func Compile(expression string) (*Expression, error) {
	//First split it up in multiple tokens
	list := make([]string, 0)
	current := ""
	for i := 0; i < len(expression); i++ {
		switch expression[i] {
		case '"':
			if len(current) != 0 {
				return nil, errors.New("Unexpected '\"'")
			}
			start := i
			for {
				i++
				next := peekString(expression, i)
				if len(next) == 0 {
					return nil, errors.New("String not terminated")
				} else if next == "\"" {
					list = append(list, expression[start:i+1])
					break
				}
			}
		case '(', ')', '*', '/', '+', '-':
			list = addToList(list, current)
			current = ""
			list = append(list, string(expression[i]))
		case '|', '&', '=':
			list = addToList(list, current)
			current = ""
			if peekString(expression, i+1) == string(expression[i]) {
				list = append(list, string(expression[i])+string(expression[i]))
				i++
			} else {
				return nil, errors.New("Invalid single character \"" + string(expression[i]) + "\".")
			}
		case '<', '>':
			list = addToList(list, current)
			current = ""
			if peekString(expression, i+1) == "=" {
				list = append(list, string(expression[i])+"=")
				i++
			} else {
				list = append(list, string(expression[i]))
			}
		case '!':
			list = addToList(list, current)
			current = ""
			if peekString(expression, i+1) == "=" {
				list = append(list, "!=")
				i++
			} else {
				list = append(list, "!")
			}
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G',
			'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
			'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
			'X', 'Y', 'Z', '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', '.', '$', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k',
			'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
			current += string(expression[i])
		case ' ':
			list = addToList(list, current)
			current = ""
		default:
			fmt.Println("Here ", i)
			return nil, errors.New("Invalid Character \"" + string(expression[i]) + "\".")
		}
	}
	list = addToList(list, current)
	current = ""
	return shuntingYard(list)
}

func init() {
	//Setup Methods and Variables of item.Item
	var it item.Item
	values = make(map[string]indexTypePair)
	v := reflect.ValueOf(it)
	//Variables
	numFields := v.Type().NumField()
	for i := 0; i < numFields; i++ {
		if v.Type().Field(i).Type.Name() == "bool" {
			values[v.Type().Field(i).Name] = indexTypePair{tokenType: type_VARBOOL, index: i}
		} else if v.Type().Field(i).Type.Name() == "float64" {
			values[v.Type().Field(i).Name] = indexTypePair{tokenType: type_VARFLOAT, index: i}
		} else if v.Type().Field(i).Type.Name() == "string" {
			values[v.Type().Field(i).Name] = indexTypePair{tokenType: type_VARSTRING, index: i}
		}
	}
	//Methods
	v = reflect.ValueOf(&it)
	numMethods := v.Type().NumMethod()
	for i := 0; i < numMethods; i++ {
		if v.Type().Method(i).Type.String() == "func(*item.Item) bool" {
			values[v.Type().Method(i).Name] = indexTypePair{tokenType: type_METHODBOOL, index: i}
		} else if v.Type().Method(i).Type.String() == "func(*item.Item) bool" {
			values[v.Type().Method(i).Name] = indexTypePair{tokenType: type_METHODFLOAT, index: i}
		}
	}
	//MapKeys
	for _, key := range item.ItemFloatMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		values[key] = indexTypePair{tokenType: type_MAPFLOAT, mapKey: key}
	}
	for _, key := range item.ItemBoolMods { //The values in item.ItemMods are the keys to the item modifiers in the actual item
		values[key] = indexTypePair{tokenType: type_MAPBOOL, mapKey: key}
	}
}
