package main

/*
#cgo LDFLAGS: -lX11
#include <stdio.h>
#include <X11/Xlib.h>

extern void onMouseEvent(int, int);

Window root_window;
unsigned int mask;
Display *display;
XEvent xevent;

static void init() {
	display = XOpenDisplay(NULL);
}

static void startMouseEventListener(){
	Display *display = XOpenDisplay(NULL);
    Window window;
    XEvent evt;

    if(display == NULL){
		return;
	}

    XAllowEvents(display, AsyncBoth, CurrentTime);
    window = DefaultRootWindow(display);
	XGrabPointer(display, root_window, 1, PointerMotionMask | ButtonReleaseMask | ButtonPressMask, GrabModeAsync, GrabModeAsync, None, None, CurrentTime);

    while(True) {
		XNextEvent(display, &evt);
		if (evt.type == ButtonPress) {
			onMouseEvent(evt.xbutton.button, 1);
		}
		if (evt.type == ButtonRelease) {
			onMouseEvent(evt.xbutton.button, 0);
		}
    }
}

static int* getMouse () {
	if (display == NULL){
		printf("You need to run init first!");
		static int r[2];
		return r;
	}
	int root_x, root_y;
	XQueryPointer(display, DefaultRootWindow(display), &root_window, &root_window, &root_x, &root_y, &root_x, &root_y, &mask);
	static int  r[2];
	r[0] = root_x;
	r[1] = root_y;
	return r;
}

static void setMousePos(int x, int y) {
	if (display == NULL){
		printf("You need to run init first!");
		return;
	}
	root_window = XRootWindow(display, 0);
	XSelectInput(display, root_window, KeyReleaseMask);
	XWarpPointer(display, None, root_window, 0, 0, 0, 0, x, y);
	XFlush(display);
}
*/
import "C"
import (
	"unsafe"
)

var callback func(int, int)

//export onMouseEvent
func onMouseEvent(a, b C.int) {
	button := (int)(a)
	state := (int)(b)
	callback(button, state)
}

func mouseInit() {
	C.init()
}

func startMouseListener(call func(int, int)) {
	callback = call
	go C.startMouseEventListener()
}

func setMousePos(x, y int) {
	C.setMousePos((C.int)(x), (C.int)(y))
}

func getMousePos() (int, int) {
	var theCArray *C.int = C.getMouse()
	length := 2
	slice := (*[1 << 28]C.int)(unsafe.Pointer(theCArray))[:length:length]
	return (int)(slice[0]), (int)(slice[1])
}
