# SSH-Desktop-remote
This is a small but effective tool to control a linux desktop via ssh and xdotool written in go.
It uses robotgo and eiannone/keyboard as client libraries for the input and translates it into xdotool commands on a remote machine. This allows you to 'mirror' your mouse and keyboard.

# Installation/Requirements
This tool uses "robotgo" so you need to install the same tools/packages which are required by robotgo
[https://github.com/go-vgo/robotgo#requirements]
If these are installed successfully you can compile it with 
```
go get
go build
```

# Usage
```
./remoteSSH <host:port> <username> <mouse/nomouse>
```
by default the mouse option is turned off. If mouse is turned on, you can toggle it afterwards with the \` key
