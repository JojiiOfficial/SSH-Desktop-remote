# SSH-Desktop-remote
This is a small but effective tool to control a linux desktop via ssh and xdotool written in go.
It uses robotgo and [eiannone/keyboard](https://github.com/eiannone/keyboard) as client libraries for the input and translates it into xdotool commands on a remote machine. This allows you to 'mirror' your mouse and keyboard.

# Installation/Requirements
This tool uses [robotgo](https://github.com/go-vgo/robotgo) so you need to install the same tools/packages which are required by robotgo
(https://github.com/go-vgo/robotgo#requirements)
If these are installed successfully you can compile it with:
```
go get
go build -o remoteSSH
```

In addition you need to install xdotool on the remote machine:
```
apt install xdotool -y
```

# Usage
```
./remoteSSH <host> <username> <mouse/nomouse>
```
By default the mouse option is turned off. If mouse is turned on, you can toggle it afterwards with the '\`' key.
<br>If the port is not 22 you can add ':port' to the hostname-argument

# Features
- Keyboard
  - [x] Keyboard ASCII keyboard support (a-Z,0-9,!@#$%^&*(){}[]+=|\/?_-)
  - [ ] Key combinations (eg. Crtl+a)
- Mouse
  - [x] mouse movement
  - [ ] mouse buttons (left-,rightmousebutton)
