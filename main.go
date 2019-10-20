package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/go-vgo/robotgo"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//Enabled a struct
type Enabled struct {
	mu      sync.Mutex
	enabled bool
}

var e *Enabled

var mouse = false
var host, user, pass string

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Usage: " + args[0] + " <host:port> <username> <mouse/nomouse>")
		return
	}

	host = args[1]

	if !strings.Contains(host, ":") {
		host += ":22"
	}

	user = args[2]

	fmt.Print(user + "'s password: ")
	bpas, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	pass := string(bpas)
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	connection, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		panic(err)
	}

	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		panic(err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	go io.Copy(os.Stdout, stdout)
	stdin, err := session.StdinPipe()

	if err != nil {
		fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	err = session.Shell()
	if err != nil {
		fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	stdin.Write([]byte("export DISPLAY=:0\n"))

	e = &Enabled{enabled: true}

	go (func() {
		for {
			char, key, err := keyboard.GetSingleKey()
			if err == nil {
				if key == 3 {
					os.Exit(0)
				}
				if key != 0 {
					fmt.Println(key)
					writeLetter(stdin, int(key))
				} else {
					fmt.Println(char)
					writeLetter(stdin, int(char))
				}
			} else {
				panic(err)
			}
		}
	})()

	if len(args) > 3 {
		mouse = args[3] == "mouse"
	}

	if mouse {
		lx, ly := robotgo.GetMousePos()
		go (func() {
			for {
				nx, ny := robotgo.GetMousePos()
				dx := lx - nx
				dy := ly - ny
				if dy != 0 && dx != 0 {
					if e.enabled {
						moveRemoteMouse(stdin, dx*-1, dy*-1)
					}
				}
				lx, ly = nx, ny
				time.Sleep(5 * time.Millisecond)
			}
		})()
	}

	for {
		lmb := robotgo.AddEvent("mleft")
		if lmb {
			remoteMouseButton(stdin, 1)
		}
	}
}

func remoteMouseButton(stdin io.WriteCloser, button int) {
	cmd := "xdotool mousedown " + strconv.Itoa(button) + ";" +
		"sleep 0.2" + ";" +
		"xdotool mouseup " + strconv.Itoa(button) + "\n"
	fmt.Print(cmd)
	stdin.Write([]byte(cmd))
}

func moveRemoteMouse(stdin io.WriteCloser, dx, dy int) {
	cmd := "xdotool mousemove_relative -- " + strconv.Itoa(dx) + " " + strconv.Itoa(dy) + "\n"
	//fmt.Print(cmd)
	stdin.Write([]byte(cmd))
}

func writeLetter(stdin io.WriteCloser, letter int) {
	stdin.Write([]byte("xdotool key " + convertToCommandCode(letter) + "\n"))
}

func convertToCommandCode(keycode int) string {
	switch keycode {
	case 32:
		return "space"
	case 13:
		return "Return"
	case 127:
		return "BackSpace"
	case 65522:
		return "Delete"
	case 65515:
		return "Left"
	case 65514:
		return "Right"
	case 65517:
		return "Up"
	case 65516:
		return "Down"
	case 96:
		return "grave"
	case 34:
		return "quotedbl"
	case 39:
		return "apostrophe"
	case 36:
		return "dollar"
	case 37:
		return "percent"
	case 38:
		return "ampersand"
	case 40:
		return "parenleft"
	case 41:
		return "parenright"
	case 42:
		return "asterisk"
	case 43:
		return "plus"
	case 45:
		return "minus"
	case 47:
		return "slash"
	case 61:
		return "equal"
	case 63:
		return "question"
	case 94:
		return "asciicircum"
	case 91:
		return "bracketleft"
	case 92:
		return "backslash"
	case 93:
		return "bracketright"
	case 95:
		return "underscore"
	case 123:
		return "braceleft"
	case 124:
		return "bar"
	case 125:
		return "braceright"
	case 8:
		return "Control_L+BackSpace"
	case 9:
		return "Tab"
	case 64:
		return "at"
	case 35:
		return "numbersign"
	case 33:
		return "exclam"
	}
	return string(keycode)
}
