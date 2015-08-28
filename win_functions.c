#include "win_functions.h"
#include "_cgo_export.h"
#include <tlhelp32.h>

boolean UpdateLoading = FALSE;

// Shuts down the program
void Exit()
{
	//if (!ExitCalled){
		Shell_NotifyIcon(NIM_DELETE, &nIconData);
		FreeMenuIfNecessary();
		UnhookWindowsHookEx(KeyboardHook);
		UnhookWindowsHookEx(MouseHook);
	//}
	//ExitCalled = TRUE;
    PostQuitMessage(0);
}


// Uses SendInput to emulate a keyboard press
void PressKeyboardKey(char key){
	    INPUT ip;
	    ip.type = INPUT_KEYBOARD;
	    ip.ki.wVk = key;
	    ip.ki.wScan = 0;
	    ip.ki.dwFlags = 0;
	    ip.ki.time = 0;
	    ip.ki.dwExtraInfo = 0;
	    SendInput(1, &ip, sizeof(INPUT));
	    ip.ki.dwFlags =  KEYEVENTF_KEYUP;
        SendInput(1, &ip, sizeof(INPUT));	
}

// Copys an item to clipboard.
// Used instead of SendInput ctrl+c, due to it being faster in this case.
void ItemToClipboard(){
	HWND foreground = GetForegroundWindow();
	if (foreground){
		SendMessage(foreground, WM_KEYDOWN, 67, 0);
		SendMessage(foreground, WM_KEYUP, 67, 0);
	}
}

// Checks whether the Current active window is called "Path of Exile"
boolean IsPoEActive(){
	HWND foreground = GetForegroundWindow();
	if (foreground){
		char window_title[18];
		GetWindowText(foreground, window_title, 18);
		if (!strcmp(window_title , "Path of Exile")){
			return TRUE;
		}
	}
	return FALSE;
}

// Called whenever user presses a keyboard key.
// Checks whether poe is running, if so checks item on hotkey pressed and evaluates it.
LRESULT CALLBACK LowLevelKeyboardProc( int nCode, WPARAM wParam, LPARAM lParam ){
	char pressedKey;
	// Declare a pointer to the KBDLLHOOKSTRUCTdsad
	KBDLLHOOKSTRUCT *pKeyBoard = (KBDLLHOOKSTRUCT *)lParam;
	switch( wParam )
	{
	case WM_KEYDOWN: // When the key has been pressed down
		{
		//get the key code
		pressedKey = (char)pKeyBoard->vkCode;
		if ((pressedKey == -94) || (pressedKey == -93)){	//Ctrl
			CtrlPressed = TRUE;
		}
		}
		break;
	case WM_KEYUP: 
       {
		//get the key code
        pressedKey = (char)pKeyBoard->vkCode;
		if ((pressedKey == -94) || (pressedKey == -93)){	//Ctrl
			CtrlPressed = FALSE;
		}else if (pressedKey == 68){	//d){
			if (CtrlPressed && IsPoEActive()){
				POINT p;
				GetCursorPos(&p);
				//Copy item to clipboard
				ItemToClipboard();
				//Read from Clipboard
				Sleep(5);
				HANDLE h;
				if (!OpenClipboard(NULL)){
					return CallNextHookEx( NULL, nCode, wParam, lParam);
				}
				h = GetClipboardData(CF_TEXT);
				const char* output = evaluateItem((char*)h);
				const size_t len = strlen(output) + 1;
				HGLOBAL hMem =  GlobalAlloc(GMEM_MOVEABLE, len);
				memcpy(GlobalLock(hMem), output, len);
				GlobalUnlock(hMem);
				EmptyClipboard();
				SetClipboardData(CF_TEXT, hMem);
				CloseClipboard();
				if (len > 1) {
					PressKeyboardKey(VK_RETURN);
					PressKeyboardKey(65);	//ctrl+a
					PressKeyboardKey(86);	//ctrl+v
					PressKeyboardKey(VK_RETURN);
				}else if (UseDoubleclick){
					mouse_event(MOUSEEVENTF_LEFTDOWN, p.x, p.y, 0, 0);
					mouse_event(MOUSEEVENTF_LEFTUP, p.x, p.y, 0, 0);
				}
			}
		}
		}
		break;
	}
    //
	return CallNextHookEx( NULL, nCode, wParam, lParam);
}


// Starts the listening for userinput.
void InstallHooks(){
	KeyboardHook = SetWindowsHookEx( WH_KEYBOARD_LL, LowLevelKeyboardProc, hInstance,0);
	//MouseHook = SetWindowsHookEx( WH_MOUSE_LL, LowLevelMouseProc, hInstance,0);
}

void ShowMenu(){
	POINT p;
    GetCursorPos(&p);
	if (go_NeedsReload()){
		CreateCustomMenu();
	}
	SetForegroundWindow(hwnd); // Win32 bug work-around
    TrackPopupMenu(hMenu, TPM_BOTTOMALIGN | TPM_LEFTALIGN, p.x, p.y, 0, hwnd, NULL);
}

void FreeMenuIfNecessary(){
	if (hMenu){
		DestroyMenu(hMenu);
		hMenu = NULL;
	}
}

void CreateCustomMenu(){
	FreeMenuIfNecessary();
	hMenu = CreatePopupMenu();
	HMENU hFilterSubMenu = CreatePopupMenu();
	HMENU hOptionsSubMenu = CreatePopupMenu();
    
	char *currFilter = go_GetCurrentFilterName();

    //Main Menu
    AppendMenu(hMenu, MF_STRING | MF_GRAYED, 0, VERSION);
	//AppendMenuW(hSubMenu, MF_SEPARATOR, 0, NULL);
	//AppendMenuW(hSubMenu, MF_STRING, CMD_ABOUT, L"About");
	AppendMenuW(hMenu, MF_STRING | MF_POPUP, (UINT_PTR)hFilterSubMenu, L"Filer");
	AppendMenuW(hMenu, MF_STRING | MF_POPUP, (UINT_PTR)hOptionsSubMenu, L"Options");
	AppendMenuW(hMenu, MF_SEPARATOR, 0, NULL);
    AppendMenuW(hMenu, MF_STRING, CMD_EXIT, L"Exit");

	//Sub Menu of Filters
	AppendMenuW(hFilterSubMenu, MF_STRING, CMD_FILTER_RELOAD, L"Reload current filter");
	AppendMenuW(hFilterSubMenu, MF_SEPARATOR, 0, NULL);
	int count = 0;
	char* name = go_GetNextFilterName();
	while ((strlen(name) > 0) && (count < 50)){
		if (strcmp(name , currFilter) == 0){
			AppendMenu(hFilterSubMenu, MF_STRING | MF_CHECKED, (UINT_PTR)hFilterSubMenu, name);
		}else{
			AppendMenu(hFilterSubMenu, MF_STRING, CMD_FILTERSELECT_START+count, name);
		}
		free(name);
		name = go_GetNextFilterName();
		count++;
	}
	free(name);
	AppendMenuW(hFilterSubMenu, MF_SEPARATOR, 0, NULL);
	if (strlen(currFilter) > 0){
		AppendMenuW(hFilterSubMenu, MF_STRING, CMD_FILTER_NONE, L"None");
	}else{
		AppendMenuW(hFilterSubMenu, MF_STRING | MF_CHECKED, CMD_FILTER_NONE, L"None");
	}
	free(currFilter);
	
	//Sub Menu of Options
	if (UseDoubleclick){
		AppendMenuW(hOptionsSubMenu, MF_STRING | MF_CHECKED, CMD_OPTIONS_USEDOUBLECLICK, L"Send Doubleclick");
	}else{
		AppendMenuW(hOptionsSubMenu, MF_STRING, CMD_OPTIONS_USEDOUBLECLICK, L"Send Doubleclick");
	}
	if (CheckForUpdates){
		AppendMenuW(hOptionsSubMenu, MF_STRING | MF_CHECKED, CMD_OPTIONS_CHECKUPDATES, L"Check for updates");
	}else{
		AppendMenuW(hOptionsSubMenu, MF_STRING, CMD_OPTIONS_CHECKUPDATES, L"Check for updates");
	}
	if (LaunchPoE){
		AppendMenuW(hOptionsSubMenu, MF_STRING | MF_CHECKED, CMD_OPTIONS_LAUNCHPOE, L"Launch PoE on start");
	}else{
		AppendMenuW(hOptionsSubMenu, MF_STRING, CMD_OPTIONS_LAUNCHPOE, L"Launch PoE on start");
	}
   // SetForegroundWindow(hWnd); // Win32 bug work-around
   // TrackPopupMenu(hMenu, TPM_BOTTOMALIGN | TPM_LEFTALIGN, p.x, p.y, 0, hWnd, NULL);

}

// Creates a Window and sets global hwnd to it.
void InitializeWindow(){
    // Register the window class.
    LPCTSTR  CLASS_NAME = "PoEPricer Window Class";
    
    WNDCLASS wc = { };

    wc.lpfnWndProc   = WindowProc;
    wc.hInstance     = hInstance;
    wc.lpszClassName = CLASS_NAME;

    RegisterClass(&wc);

    // Create the window.

    hwnd = CreateWindowEx(
        0,                              // Optional window styles.
        CLASS_NAME,                     // Window class
        "PoEPricer",    				// Window text
        WS_OVERLAPPEDWINDOW,            // Window style

        // Size and position
        //CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT,
		0,0,0,0,
        NULL,       // Parent window    
        NULL,       // Menu
        hInstance,  // Instance handle
        NULL        // Additional application data
        );

    if (hwnd == NULL)
    {
        Exit();
    }
    ShowWindow(hwnd, FALSE);

}


// Set in Initialize Window, receives WindowMessages (like when to show popup menu)
LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam){
    switch (uMsg)
    {
    case WM_DESTROY:
        Exit();
		return 0;
	case WM_COMMAND:
		if( wParam == CMD_EXIT){
			Exit();
			return 0;
		}else if( wParam == CMD_FILTER_NONE){
			go_LoadFilterByIndex(-1);
		}else if( wParam == CMD_FILTER_RELOAD){
			go_ReloadFilter();
		}else if( wParam == CMD_OPTIONS_USEDOUBLECLICK){
			UseDoubleclick = !UseDoubleclick;
			go_SetUseDoubleclick(UseDoubleclick);
		}else if( wParam == CMD_OPTIONS_CHECKUPDATES){
			CheckForUpdates = !CheckForUpdates;
			go_SetCheckForUpdates(CheckForUpdates);
		}else if( wParam == CMD_OPTIONS_LAUNCHPOE){
			LaunchPoE = !LaunchPoE;
			go_SetLaunchPoE(LaunchPoE);
		}else if(wParam >= CMD_FILTERSELECT_START && wParam < CMD_FILTERSELECT_START+50){
			go_LoadFilterByIndex(wParam - CMD_FILTERSELECT_START);
		}
		 /*switch (LOWORD(wParam))
            {
                case CMD_EXIT:
                    
			}*/
	case WM_POEPRICERMESSAGE:
		if (lParam == WM_RBUTTONUP){
			ShowMenu();
            return 0;
		}
		break;
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}


// Creates the window needed for popup menu, then registers a tray icon, as well as
// making sure to receive WM_POEPRICERMESSAGE whenever icon is rightclicked.
void ShowTrayIcon(){
	//Get a window for our tray
	InitializeWindow(); // Exits if it fails
	//Load icon image
	HICON hIcon;
	hIcon = LoadImage(NULL, "data/icon.ico", IMAGE_ICON, 64, 64, LR_LOADFROMFILE);
	//hIcon = LoadImage(NULL, IDI_INFORMATION, IMAGE_ICON, 16, 16, LR_SHARED);
	//Initialize Icon
	nIconData.cbSize = sizeof(NOTIFYICONDATA);
    nIconData.hWnd = hwnd;
    nIconData.uID = 100;
    nIconData.uCallbackMessage = WM_POEPRICERMESSAGE;
    nIconData.hIcon = hIcon;
	strcpy(nIconData.szTip, "PoEPricer");
    nIconData.uFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP;
	Shell_NotifyIcon(NIM_ADD, &nIconData);
}



void c_AskForUpdate(char* msg){
	int answer = MessageBox(NULL, msg, "Update available", MB_YESNO|MB_ICONINFORMATION);
	free(msg);
	if (answer == IDYES){
		UpdateLoading = TRUE;
		ShellExecute(NULL, "open", "http://poe.melanite.net", NULL, NULL, SW_SHOWNORMAL);
	}
}


void launchPoE(){
	DWORD dwType = REG_SZ;
	HKEY hKey = 0;
	char directory[1024];
	DWORD value_length = 1024;
	const char* subkey = "Software\\GrindingGearGames\\Path of Exile";
	RegOpenKey(HKEY_CURRENT_USER,subkey,&hKey);
	if (RegQueryValueEx(hKey, "InstallLocation", NULL, &dwType, (LPBYTE)&directory, &value_length) != ERROR_SUCCESS){
			MessageBox(NULL, "Could not determine Path of Exile's install location.\nDisabling autostart.", "Error", MB_OK|MB_ICONERROR);
			LaunchPoE = FALSE;
			go_SetLaunchPoE(LaunchPoE);
			return;
	}
	char executable[1024];
	strcpy(executable, directory);
	strcat(executable, "\\PathOfExile.exe");
	STARTUPINFO si;
    PROCESS_INFORMATION pi;
    ZeroMemory( &si, sizeof(si) );
	ZeroMemory( &pi, sizeof(pi) );
	si.cb = sizeof(si);
	if (!CreateProcess( executable,   // No module name (use command line)
        NULL ,       // Command line
        NULL,           // Process handle not inheritable
        NULL,           // Thread handle not inheritable
        FALSE,          // Set handle inheritance to FALSE
        0,              // No creation flags
        NULL,           // Use parent's environment block
        directory, // Use given starting directory 
        &si,            // Pointer to STARTUPINFO structure
        &pi )          // Pointer to PROCESS_INFORMATION structure
        ){
			MessageBox(NULL, "Could not start Path of Exile.\nDisabling autostart.", "Error", MB_OK|MB_ICONERROR);
			LaunchPoE = FALSE;
			go_SetLaunchPoE(LaunchPoE);
			return;
		}
	PoEHandle = pi.hProcess;
}

DWORD FindProcessId(char* processName){
    // strip path

    char* p = strrchr(processName, '\\');
    if(p)
        processName = p+1;

    PROCESSENTRY32 processInfo;
    processInfo.dwSize = sizeof(processInfo);

    HANDLE processesSnapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if ( processesSnapshot == INVALID_HANDLE_VALUE )
        return 0;

    Process32First(processesSnapshot, &processInfo);
    if ( !strcmp(processName, processInfo.szExeFile) )
    {
        CloseHandle(processesSnapshot);
        return processInfo.th32ProcessID;
    }

    while ( Process32Next(processesSnapshot, &processInfo) )
    {
        if ( !strcmp(processName, processInfo.szExeFile) )
        {
          CloseHandle(processesSnapshot);
          return processInfo.th32ProcessID;
        }
    }

    CloseHandle(processesSnapshot);
    return 0;
}

void c_Loop(char* version, boolean useDoubleclick, boolean checkUpdates, boolean launchPoEBool){
	VERSION = version;
	UseDoubleclick = useDoubleclick;
	CheckForUpdates = checkUpdates;
	LaunchPoE = launchPoEBool;
	PoEHandle = NULL;
	if(LaunchPoE && FindProcessId("PathOfExile.exe") == 0 && !UpdateLoading){
		launchPoE();
	}
	//Initialize Variables
	//ExitCalled = FALSE;
	hInstance = GetModuleHandle(NULL);
	CtrlPressed = FALSE;
	//Install Hooks
	InstallHooks();
	//c_LoadFilters();
	//CreateCustomMenu();	//Created automatically when needed
	ShowTrayIcon();
    MSG msg = {0};
	if(PoEHandle != NULL){
		SetTimer(NULL, 1, 1000, NULL);
	}
    while (GetMessage(&msg, NULL, 0, 0) != 0)
    {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
		if (PoEHandle != NULL){
			DWORD dwExitCode;
			GetExitCodeProcess(PoEHandle,&dwExitCode);
			if(dwExitCode != STILL_ACTIVE) {
				Exit();
			}
		}		
    }
	Exit();
	
}
