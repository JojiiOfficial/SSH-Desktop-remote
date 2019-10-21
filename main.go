package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/mkideal/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

type argT struct {
	cli.Helper
	Host    string `cli:"*r,host" usage:"SSH host or username@host"`
	User    string `cli:"l,user" usage:"SSH user"`
	Port    int    `cli:"p,port" usage:"SSH remote port"`
	KeyFile string `cli:"i,identity" usage:"SSH identity file"`
	Mouse   bool   `cli:"m,mouse" usage:"boolean mirror mouse" dft:"false"`
}

//Enabled a struct
type Enabled struct {
	mu      sync.Mutex
	enabled bool
}

var e *Enabled
var stdin io.WriteCloser
var help = cli.HelpCommand("display help information")

func main() {
	if err := cli.Root(child,
		cli.Tree(help),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func sshagent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

var child = &cli.Command{
	Name: "remote",
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv()
		argt := argv.(*argT)
		usern := argt.User
		host := argt.Host

		if len(usern) == 0 && !strings.Contains(host, "@") {
			user, err := user.Current()
			if err != nil {
				panic(err)
			}
			if user != nil && len(user.Username) > 0 {
				usern = user.Username
			} else {
				fmt.Println("Error getting user. Try adding the user manually: -l or --user")
				os.Exit(1)
				return nil
			}
		} else if strings.Contains(host, "@") {
			data := strings.Split(host, "@")
			usern = data[0]
			host = data[1]
		}

		if argt.Port == 0 && !strings.Contains(host, ":") {
			host += ":22"
		} else if strings.Contains(host, ":") && argt.Port != 0 {
			d := strings.Split(host, ":")[1]
			fmt.Println("Warning! Using port '" + d + "' instead of '" + strconv.Itoa(argt.Port) + "'")
		} else if argt.Port != 0 {
			host += ":" + strconv.Itoa(argt.Port)
		}

		var sshauth ssh.AuthMethod
		if len(argt.KeyFile) > 0 {
			if _, err := os.Stat(argt.KeyFile); err == nil {
				buffer, err := ioutil.ReadFile(argt.KeyFile)
				if err != nil {
					fmt.Println("Couldn't read keyfile!")
					os.Exit(1)
					return nil
				}
				key, err := ssh.ParsePrivateKey(buffer)
				if err != nil {
					fmt.Println("Couldn't read keyfile!")
					os.Exit(1)
					return nil
				}
				sshauth = ssh.PublicKeys(key)
			} else {
				fmt.Println("File does not exists!")
				os.Exit(1)
				return nil
			}
		} else {
			sshauth = sshagent()
		}

		var connection *ssh.Client
		var err error
		for i := 0; ; {
			if i >= 4 {
				fmt.Println("\nToo many attempts!")
				return nil
			}
			if sshauth == nil {
				if i > 1 {
					fmt.Print("\n" + usern + "'s password: ")
				} else {
					fmt.Print(usern + "'s password: ")
				}
				bpas, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					panic(err)
				}
				pass := string(bpas)
				sshauth = ssh.Password(pass)
			}

			connection, err = ssh.Dial("tcp", host, &ssh.ClientConfig{
				User: usern,
				Auth: []ssh.AuthMethod{
					sshauth,
				},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					return nil
				},
			})

			if err != nil {
				sshauth = nil
				i++
				continue
			} else {
				break
			}
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
		_ = stdout
		stdin, err = session.StdinPipe()

		if err != nil {
			fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		err = session.Shell()
		if err != nil {
			fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		stdin.Write([]byte("export DISPLAY=:0\n"))
		fmt.Println("\nSuccessfully connected")
		fmt.Println("Available:\n - Keyboard")
		if argt.Mouse {
			fmt.Println(" - Mouse")
		}

		e = &Enabled{enabled: true}

		go (func() {
			for {
				char, key, err := keyboard.GetSingleKey()
				if err == nil {
					if key == 3 {
						os.Exit(0)
					}
					if key != 0 {
						writeLetter(stdin, int(key))
					} else {
						writeLetter(stdin, int(char))
					}
				} else {
					panic(err)
				}
			}
		})()

		if argt.Mouse {
			mouseInit()
			var nx, ny int
			lx, ly := getMousePos()

			startMouseListener(func(a, b, c int) {
				if a > 0 {
					remoteMouseButton(stdin, a, b)
				} else {
					nx, ny = b, c
					dx := lx - nx
					dy := ly - ny
					if nx != lx && ny != ly {
						if e.enabled {
							setMousePos(lx, ly)
							moveRemoteMouse(stdin, dx*-1, dy*-1)
							time.Sleep(4 * time.Millisecond)
						}
					}
				}
			})
		}

		for {

		}
	},
}

func remoteMouseButton(stdin io.WriteCloser, button, state int) {
	sstate := "mouseup"
	if state == 1 {
		sstate = "mousedown"
	}
	cmd := "xdotool " + sstate + " " + strconv.Itoa(button) + "\n"
	stdin.Write([]byte(cmd))
}

func moveRemoteMouse(stdin io.WriteCloser, dx, dy int) {
	cmd := "xdotool mousemove_relative -- " + strconv.Itoa(dx) + " " + strconv.Itoa(dy) + "\n"
	stdin.Write([]byte(cmd))
}

func writeLetter(stdin io.WriteCloser, letter int) {
	stdin.Write([]byte("xdotool key " + convertToCommandCode(letter) + "\n"))
}

func convertToCommandCode(keycode int) string {
	switch keycode {
	case 8:
		return "Control_L+BackSpace"
	case 9:
		return "Tab"
	case 13:
		return "Return"
	case 32:
		return "space"
	case 33:
		return "exclam"
	case 34:
		return "quotedbl"
	case 35:
		return "numbersign"
	case 36:
		return "dollar"
	case 37:
		return "percent"
	case 38:
		return "ampersand"
	case 39:
		return "apostrophe"
	case 40:
		return "parenleft"
	case 41:
		return "parenright"
	case 42:
		return "asterisk"
	case 43:
		return "plus"
	case 44:
		return "comma"
	case 45:
		return "minus"
	case 46:
		return "period"
	case 47:
		return "slash"
	case 58:
		return "colon"
	case 59:
		return "semicolon"
	case 61:
		return "equal"
	case 63:
		return "question"
	case 64:
		return "at"
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
	case 96:
		return "grave"
	case 123:
		return "braceleft"
	case 124:
		return "bar"
	case 125:
		return "braceright"
	case 127:
		return "BackSpace"
	case 65514:
		return "Right"
	case 65515:
		return "Left"
	case 65516:
		return "Down"
	case 65517:
		return "Up"
	case 65522:
		return "Delete"
	}
	return string(keycode)
}
