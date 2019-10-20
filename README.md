# SSH-Desktop-remote
This is a small but effective tool to control the keyboard and mouse of a linux desktop environment via SSH.
It uses [robotgo](https://github.com/go-vgo/robotgo) and [eiannone/keyboard](https://github.com/eiannone/keyboard) as client libraries for the input, translates this into xdotool commands and executes them on a remote machine. This allows you to 'mirror' your mouse and keyboard input.

# Installation/Requirements
This tool uses [robotgo](https://github.com/go-vgo/robotgo) so you need to install the same tools/packages which are required by robotgo
(https://github.com/go-vgo/robotgo#requirements)
If these are installed successfully you can compile this tool with:
```
go get
go build -o remoteSSH
```

In addition you need to install xdotool on the remote machine if it isn't already installed. 
```
apt install xdotool -y
```

# Usage
```
./remoteSSH -r user@host                    # opens keyboard remote only
./remoteSSH -r user@host -m                 # opens mouse and keyboard remote
./remoteSSH -r user@host -m -i key -p 123   # opens remote with ssh privatekey auth on port 123
```
By default the mouse option is turned off. If mouse is turned on, you can toggle it afterwards with the '\`' key.
<br>If the port is not 22 you can add ':port' to the hostname-argument

# Features
- SSH
  - [x] SSH password auth
  - [X] Privatekey auth (only without passphrase)
  - [x] SSHAgent login
  - [ ] Knownhost check/handling
- Keyboard
  - [x] Keyboard ASCII keyboard support (a-Z,0-9,!@#$%^&*(){}[]+=|\/?_-)
  - [ ] Key combinations (eg. Crtl+a)
- Mouse
  - [x] Mouse movement
  - [x] Simple mouse leftclick
  - [ ] Full mousebutton support (left- ,rightmousebutton, scrolling, ect)
