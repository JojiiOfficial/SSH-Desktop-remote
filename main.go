package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	_ "github.com/eiannone/keyboard"
	"github.com/go-vgo/robotgo"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

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

	go (func() {
		for {
			char, key, err := keyboard.GetSingleKey()
			if err == nil {
				if key == 3 {
					return
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
		enabled := true
		go (func() {
			for {
				if !enabled {
					continue
				}
				nx, ny := robotgo.GetMousePos()
				dx := lx - nx
				dy := ly - ny
				if dy != 0 && dx != 0 {
					fmt.Println("X: " + strconv.Itoa(dx) + " Y: " + strconv.Itoa(dy))
					moveRemoteMouse(stdin, dx*-1, dy*-1)

				}
				lx, ly = nx, ny
				time.Sleep(5 * time.Millisecond)
			}
		})()

		go (func() {
			for {
				keve := robotgo.AddEvent("`")
				if keve {
					enabled = !enabled
					if enabled {
						lx, ly = robotgo.GetMousePos()
					}
				}
			}
		})()
	}

	for {

	}
}

func moveRemoteMouse(stdin io.WriteCloser, dx, dy int) {
	cmd := "xdotool mousemove_relative -- " + strconv.Itoa(dx) + " " + strconv.Itoa(dy) + "\n"
	fmt.Println(cmd)
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
	case 8:
		return "Control_L+BackSpace"
	}
	return string(keycode)
}
