package main

/*
#cgo LDFLAGS: -lX11
#include <X11/X.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/XKBlib.h>
#include <stdio.h>
#include <ctype.h>
Bool running;

extern void keyboardEvent(char*,int);

static void startKeyboardListener() {
	Display *display = XOpenDisplay(NULL);
    XGrabKeyboard(display, DefaultRootWindow(display), True, GrabModeAsync, GrabModeAsync, CurrentTime);
    XEvent event;
    running = True;
    while(running) {
        XNextEvent(display, &event);
        switch (event.type){
            case KeyPress: {
                int a;
                char *key = XKeysymToString(*XGetKeyboardMapping(display,event.xkey.keycode,1,&a));
                keyboardEvent(key,1);
                continue;
            }
            case KeyRelease: {
                int a;
                char *key = XKeysymToString(*XGetKeyboardMapping(display,event.xkey.keycode,1,&a));
                keyboardEvent(key,0);
                continue;
            }
        }
    }
    XUngrabKeyboard(display, CurrentTime);
    XFlush(display);
}
static void releaseKeyboard() {
	running = False;
}
*/
import "C"

var keyboardCallback func(string, bool)

func startKeyboardListen(calback func(string, bool)) {
	keyboardCallback = calback
	go C.startKeyboardListener()
}

//export keyboardEvent
func keyboardEvent(key *C.char, pressed C.int) {
	sKey := C.GoString(key)
	press := false
	if pressed == 1 {
		press = true
	}
	keyboardCallback(sKey, press)
}

func releaseKeyboard() {
	C.releaseKeyboard()
}
