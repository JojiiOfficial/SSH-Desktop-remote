package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mkideal/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

type argT struct {
	cli.Helper
	Host        string `cli:"*r,host" usage:"SSH host or username@host"`
	User        string `cli:"l,user" usage:"SSH user"`
	Port        int    `cli:"p,port" usage:"SSH remote port"`
	KeyFile     string `cli:"i,identity" usage:"SSH identity file"`
	Mouse       bool   `cli:"m,mouse" usage:"boolean mirror mouse" dft:"false"`
	Quiet       bool   `cli:"q,quiet" usage:"No output" dft:"false"`
	MouseToggle bool   `cli:"t,mtggle" usage:"Switch between devices with \u0060-key" dft:"true"`
	Sensitivity int64  `cli:"s,sensitivity" usage:"Send more mouse motions. May cause lags. Lower=more sensitive" dft:"7900000"`
	ShowMouse   bool   `cli:"e,showmouse" usage:"Don't hide mouse even if unclutter is installed" dft:"false"`
}

//Enabled a struct
type Enabled struct {
	mu      sync.Mutex
	enabled bool
}

var e *Enabled
var stdin io.WriteCloser
var help = cli.HelpCommand("display help information")
var sensitivity int64
var mouseToggle, quiet, showmouse bool

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
		sensitivity = argt.Sensitivity
		mouseToggle = argt.MouseToggle
		quiet = argt.Quiet
		showmouse = argt.ShowMouse

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
				fmt.Println("")
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
		if !argt.Quiet {
			fmt.Println("Successfully connected")
			fmt.Println("Available:\n - Keyboard")
			if argt.Mouse {
				fmt.Println(" - Mouse")
			}
		}

		e = &Enabled{enabled: true}

		inKeyboard()

		if argt.Mouse {
			mouseInit()
			inMouse()
		}

		for {

		}
	},
}

func inKeyboard() {
	startKeyboardListen(func(key string, pressed bool) {
		if key == "Control_R" {
			showMouse()
			os.Exit(0)
			return
		}
		pressRemoteKey(stdin, mouseToggle, key, pressed)
	})
}

func inMouse() {
	go hideMouse()
	var nx, ny int
	lx, ly := getMousePos()
	screenX, screenY := getDisplaySize()
	minCornerDistance := 200
	if lx < minCornerDistance {
		lx = minCornerDistance
	}
	if lx >= screenX-minCornerDistance {
		lx = screenX - minCornerDistance
	}
	if ly < minCornerDistance {
		ly = minCornerDistance
	}
	if ly >= screenY-minCornerDistance {
		ly = screenY - minCornerDistance
	}
	setMousePos(lx, ly)
	lc := time.Now().UnixNano()
	startMouseListener(func(a, b, c int) {
		if !e.enabled {
			return
		}
		if a > 0 {
			remoteMouseButton(stdin, a, b)
		} else {
			nx, ny = b, c
			dx := lx - nx
			dy := ly - ny
			if (nx != lx || ny != ly) && (lc+sensitivity <= time.Now().UnixNano()) {
				setMousePos(lx, ly)
				moveRemoteMouse(stdin, dx*-1, dy*-1)
				lc = time.Now().UnixNano()
			}
		}
	})
}

var unclutterCMD *exec.Cmd

func hideMouse() {
	if !showmouse {
		unclutterCMD = exec.Command("/usr/bin/unclutter", "-idle", "0")
		unclutterCMD.Run()
	}
}

func showMouse() {
	if unclutterCMD != nil && unclutterCMD.Process != nil && !showmouse {
		unclutterCMD.Process.Kill()
	}
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

func pressRemoteKey(stdin io.WriteCloser, mouseToggle bool, key string, pressed bool) {
	if key == "grave" && mouseToggle {
		e.enabled = false
		releaseMouse()
		showMouse()
		releaseKeyboard()
		if !quiet {
			fmt.Println("Input detached")
		}
		go (func() {
			for {
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('`')
				if input == "`" {
					e.enabled = true
					if !quiet {
						fmt.Println("Input attached")
					}
					inMouse()
					inKeyboard()
				}
			}
		})()
		return
	}
	cmd := "keydown"
	if !pressed {
		cmd = "keyup"
	}
	stdin.Write([]byte("xdotool " + cmd + " " + key + "\n"))
}
