package main

/*
	#include "win_functions.h"
	#include <windows.h>
	#include <stdlib.h>
*/
import "C"
import (
	"PoEPricer/item"
	//	"PoEPricer/stack"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	MB_OK              = 0x00000000
	MB_ICONEXCLAMATION = 0x00000030
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
)

var lastFilerReload time.Time
var currentFilterIteration = 0

//export evaluateItem
func evaluateItem(cstr *C.char) *C.char {
	input := C.GoString(cstr)
	result := ""
	it, err := item.ParseItem(input)
	if err != nil {
		fmt.Println("Error: ", err)
		log.Println("Parsing Error: ", err)
		result = "Error on parsing item."
	} else if filter2 != nil {
		/*for i := 0; i < len(subFilterList); i++ {
			if subFilterList[i].expression.Execute(it) {
				result = subFilterList[i].message
				break
			}
		}*/
		result = filter2.Execute(it)

	}
	if len(result) > 0 {
		result = messagePrefix + result
	}
	return C.CString(result)
}

func MessageBox(text string, caption string, style int) int {
	cText := C.CString(text)
	cCaption := C.CString(caption)
	result := C.MessageBox(nil, (*C.CHAR)(unsafe.Pointer(cText)), (*C.CHAR)(unsafe.Pointer(cCaption)), C.UINT(style))
	C.free(unsafe.Pointer(cText))
	C.free(unsafe.Pointer(cCaption))
	return int(result)
}

/*func onError(err error) {

}*/

//export go_LoadFilters
/*func go_LoadFilters() **C.char {
	result := make([]*C.char, 0)
	result = append(result, C.CString(currentFilter))
	fileinfo, err := ioutil.ReadDir("")
	if err != nil {
		onError(err)
		return nil
	}
	for i := 0; i < len(fileinfo); i++ {
		if strings.HasSuffix(fileinfo[i].Name(), ".filter") {
			result = append(result, C.CString(fileinfo[i].Name()))
		}
	}
	lastFilerReload = time.Now()
	return &result[0]
}*/

func getWorkingDir() string {
	currDir, err := os.Getwd()
	if err != nil {
		MessageBox("Could not find out working directory: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
		os.Exit(1)
	}
	return currDir
}

//export go_ReloadFilter
func go_ReloadFilter() {
	err := loadFilter2(currentFilter)
	if err != nil {
		currentFilter = ""
		//subFilterList = make([]*SubFilter, 0)
		filter2 = nil
		MessageBox("Could not reload filter: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
	} else {
		MessageBox("Successfully reloaded filter.", "Info", MB_OK|MB_ICONINFORMATION)
	}
	filterListDirty = true
}

//export go_LoadFilterByIndex
func go_LoadFilterByIndex(index int) {
	if index < 0 { //Don't load a filter
		options["currentFilter"] = ""
		writeOptions()
		currentFilter = ""
		//subFilterList = make([]*SubFilter, 0)
		filter2 = nil
	} else {
		err := loadFilter2(filterNames[index])
		if err != nil {
			MessageBox("Could not load filter: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
		} else {
			MessageBox("Successfully loaded filter.", "Info", MB_OK|MB_ICONINFORMATION)
		}
	}
	filterListDirty = true
}

//export go_GetNextFilterName
func go_GetNextFilterName() *C.char {
	var result string
	if currentFilterIteration < len(filterNames) {
		result = filterNames[currentFilterIteration]
		currentFilterIteration++
	}
	return C.CString(result)
}

//export go_GetCurrentFilterName
func go_GetCurrentFilterName() *C.char {
	currentFilterIteration = 0
	filterNames = make([]string, 0)
	fileinfo, err := ioutil.ReadDir(getWorkingDir())
	if err != nil {
		MessageBox("Could not open current directory to look for filter files: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
		os.Exit(1)
	}
	for i := 0; i < len(fileinfo); i++ {
		if strings.HasSuffix(fileinfo[i].Name(), ".filter") {
			filterNames = append(filterNames, fileinfo[i].Name())
		}
	}
	lastFilerReload = time.Now()
	filterListDirty = false
	return C.CString(currentFilter)
}

//export go_NeedsReload
func go_NeedsReload() bool {
	if filterListDirty {
		return true
	}
	fileinfo, err := ioutil.ReadDir(getWorkingDir())
	if err != nil {
		MessageBox("Could not open current directory to look for filter files: "+err.Error(), "Error", MB_OK|MB_ICONERROR)
		os.Exit(1)
	}
	for i := 0; i < len(fileinfo); i++ {
		file := fileinfo[i]
		if strings.HasSuffix(file.Name(), ".filter") && file.ModTime().After(lastFilerReload) {
			return true
		}
	}
	return false
}

//export go_SetUseDoubleclick
func go_SetUseDoubleclick(use bool) {
	options["useDoubleclick"] = strconv.FormatBool(use)
	writeOptions()
	filterListDirty = true
}

//export go_SetCheckForUpdates
func go_SetCheckForUpdates(update bool) {
	options["checkForUpdates"] = strconv.FormatBool(update)
	writeOptions()
	filterListDirty = true
}

//export go_SetLaunchPoE
func go_SetLaunchPoE(launch bool) {
	options["launchPoEOnStart"] = strconv.FormatBool(launch)
	writeOptions()
	filterListDirty = true
}

//export go_Tick
func go_Tick() {
	fmt.Println("Tick")
}

func checkUpdate() {
	client := &http.Client{}
	resp, err := client.Get("http://poe.melanite.net/PoEPricer_Version.php")
	if err != nil {
		log.Println("Could not connect to update server: ", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Println("Received bad status code from update server: ", resp.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(resp.Body) //data should be a version string
	newVersion, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		log.Println("Could not parse Version received from update server: \"", data, "\"")
		return
	}
	if newVersion > VERSION {
		C.c_AskForUpdate(C.CString("There is a new version (" + string(data) + ") available.\n\nDownload it now?"))
	}
}

func bToi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func loop() {
	C.c_Loop(C.CString("PoEPricer "+strconv.FormatFloat(VERSION, 'f', 2, 64)), C.boolean(bToi(checkUseDoubleclick())),
		C.boolean(bToi(wantUpdateCheck())), C.boolean(bToi(launchPoE())))
}
