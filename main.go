// main.go
package main

import (
	//	"PoEPricer/item"
	"PoEPricer/stack"
	"PoEPricer/stack2"
	"errors"
	//	"fmt"
	"io/ioutil"
	"log"

	"os"
	"strconv"
	"strings"
	//	"time"
)

type SubFilter struct {
	expression *stack.Expression
	message    string
}

const VERSION = 1.01

var currentFilter string
var filter2 *stack2.Filter
var subFilterList []*SubFilter
var filterNames []string
var options map[string]string
var filterListDirty bool
var messagePrefix string

func expandFunctions(in string, definitions map[string]string) string {
	for key, value := range definitions {
		in = strings.Replace(in, key, "("+value+")", -1)
	}
	return in
}

func loadFilter(name string) error {
	definitions := make(map[string]string)
	filters := make([]*SubFilter, 0)
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	dataStr := strings.Replace(string(data), "\r", "", -1)
	lines := strings.Split(dataStr, "\n")
	conditionNext := true
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "Function ") {
			content := strings.SplitN(line[9:], " ", 2)
			if len(content) != 2 {
				return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": Invalid Function.")
			}
			if !strings.HasPrefix(content[0], "$") || !strings.HasSuffix(content[0], "$") {
				return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": Function name must be enclosed in '$'.")
			}
			definitions[content[0]] = expandFunctions(content[1], definitions)
		} else if strings.HasPrefix(line, "Condition ") {
			if !conditionNext {
				return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": Expected 'Warn', found 'Condition'.")
			}
			sanitized := expandFunctions(line[10:], definitions)
			conditionNext = false
			filter := new(SubFilter)
			exp, err := stack.Compile(sanitized)
			if err != nil {
				return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": " + err.Error())
			}
			filter.expression = exp
			filter.message = "Message not set."
			filters = append(filters, filter)
		} else if strings.HasPrefix(line, "Warn ") {
			if conditionNext {
				return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": Expected 'Condition', found 'Warn'.")
			}
			conditionNext = true
			filters[len(filters)-1].message = line[5:]
		} else if !strings.HasPrefix(line, "#") && len(line) > 0 {
			return errors.New("Line " + strconv.FormatInt(int64(i+1), 10) + ": Comment lines must start with a '#', Conditions with 'Condition ', Messages with 'Warn ' and Functions with 'Function '.")
		}
	}
	if !conditionNext {
		return errors.New("Expected 'Warn', found end of the file.")
	} else { //All went well, update options file
		options["currentFilter"] = name
		writeOptions()

		currentFilter = name
		subFilterList = filters
	}
	return nil
}

func loadFilter2(name string) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	dataStr := strings.Replace(string(data), "\r", "", -1)
	filter2, err = stack2.Compile(dataStr)
	if err != nil {
		options["currentFilter"] = ""
		writeOptions()
		currentFilter = ""
	} else {
		options["currentFilter"] = name
		writeOptions()
		currentFilter = name
	}
	return err
}

func writeOptions() {
	txt := ""
	for key, value := range options {
		txt += key + "=" + value + "\n"
	}
	err := ioutil.WriteFile("options.txt", []byte(txt), 0666)
	if err != nil {
		MessageBox("Could not write to options.txt: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
	}
}

func readOptions() {
	options = make(map[string]string)
	data, err := ioutil.ReadFile("options.txt")
	if err != nil {
		MessageBox(err.Error(), "Error", MB_OK|MB_ICONERROR)
		os.Exit(1)
	}
	dataStr := strings.Replace(string(data), "\r", "", -1)
	lines := strings.Split(dataStr, "\n")
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			continue //skip empty lines
		}
		content := strings.Split(lines[i], "=")
		if len(content) != 2 {
			MessageBox("Invalid line in options.txt: "+lines[i], "Error", MB_OK|MB_ICONERROR)
			os.Exit(1)
		}
		options[content[0]] = content[1]
	}
	neededParams := []string{"currentFilter"}
	for i := 0; i < len(neededParams); i++ {
		_, ok := options[neededParams[i]]
		if !ok {
			MessageBox("options.txt misses required parameter "+neededParams[i], "Error", MB_OK|MB_ICONERROR)
			os.Exit(1)
		}
	}
}

/*func testFilter2() {
	text := `Rarity: Unique
Asphyxia's Wrath
Two-Point Arrow Quiver
--------
Requirements:
Level: 10
--------
Item Level: 73
--------
26% increased Accuracy Rating
--------
10% increased Attack Speed
+39% to Cold Resistance
36% increased Chill Duration on Enemies
20% of Physical Damage Converted to Cold Damage
7% chance to Freeze
Culling Strike
Curses on Slain Enemies are transferred to a nearby Enemy
--------
Mist of breath
Icing to lips and throat
As the warm ones choke and fall
Upon the frozen wasteland.
--------
Corrupted`
	it, err := item.ParseItem(text)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		ITERATIONS := 1
		t := time.Now()
		for i := 0; i < ITERATIONS; i++ {
			filter2.Execute(it)
		}
		fmt.Println("Time: ", time.Since(t))
		fmt.Println("Warn: ", filter2.Execute(it))

	}
}*/

func checkUseDoubleclick() bool {
	use := true
	str, ok := options["useDoubleclick"]
	if !ok {
		options["useDoubleclick"] = "true"
		writeOptions()
	} else {
		b, err := strconv.ParseBool(str)
		if err != nil {
			options["useDoubleclick"] = "true"
			writeOptions()
		} else {
			use = b
		}
	}
	return use
}

func wantUpdateCheck() bool {
	update := true
	str, ok := options["checkForUpdates"]
	if !ok {
		options["checkForUpdates"] = "true"
		writeOptions()
	} else {
		b, err := strconv.ParseBool(str)
		if err != nil {
			options["checkForUpdates"] = "true"
			writeOptions()
		} else {
			update = b
		}
	}
	return update
}

func launchPoE() bool {
	launch := false
	str, ok := options["launchPoEOnStart"]
	if !ok {
		options["launchPoEOnStart"] = "false"
		writeOptions()
	} else {
		b, err := strconv.ParseBool(str)
		if err != nil {
			options["launchPoEOnStart"] = "false"
			writeOptions()
		} else {
			launch = b
		}
	}
	return launch
}

func setMessagePrefix() {
	pre, ok := options["messagePrefix"]
	if !ok {
		messagePrefix = "@# "
		options["messagePrefix"] = messagePrefix
		writeOptions()
	} else {
		messagePrefix = pre
	}
}

func main() {
	f, _ := os.OpenFile("log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	log.SetOutput(f)
	readOptions()
	defaultFilter, ok := options["currentFilter"]
	if !ok {
		MessageBox("options.txt misses required parameter currentFilter", "Error", MB_OK|MB_ICONERROR)
		os.Exit(1)
	}
	if defaultFilter != "" {
		err := loadFilter2(defaultFilter)
		if err != nil {
			filter2 = nil
			MessageBox("Could not load filter: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
		}
	}
	setMessagePrefix()
	//testFilter2()
	if wantUpdateCheck() {
		checkUpdate()
	}
	loop()
}
