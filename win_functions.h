#ifndef __WIN_FUNCS_FOR_GO__
#define __WIN_FUNCS_FOR_GO__

#include <stdlib.h>
#include <Windows.h>


	#define WM_POEPRICERMESSAGE (WM_USER + 1)
	#define CMD_EXIT 1001
	#define CMD_ABOUT 1002
	#define CMD_FILTER_NONE 1003
	#define CMD_FILTER_RELOAD 1004
	#define CMD_OPTIONS_USEDOUBLECLICK 1005
	#define CMD_OPTIONS_CHECKUPDATES 1006
	#define CMD_OPTIONS_LAUNCHPOE 1007
	#define CMD_FILTERSELECT_START 1010	//Numbers after are reserved for each filter
	
	HINSTANCE hInstance;
	NOTIFYICONDATA nIconData;
	HHOOK KeyboardHook;
	HHOOK MouseHook;
	HWND hwnd;
	HMENU hMenu;
	boolean CtrlPressed;
	boolean UseDoubleclick;
	boolean CheckForUpdates;
	//boolean ExitCalled;
	boolean LaunchPoE;
	char* VERSION;
	HANDLE PoEHandle;
	
	LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
	void FreeMenuIfNecessary();
	void CreateCustomMenu();


   //char* cstrings = "C stringd example";
	void c_Loop(char*, boolean, boolean, boolean);
	void c_MessageBox(char*, int);
	void c_AskForUpdate(char*);
	
#endif