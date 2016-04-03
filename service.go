// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

/*
Package winsvc provids easy windows service support.

Example

This a simple windows service example:

	package main

	import (
		"flag"
		"fmt"
		"log"
		"net/http"
		"os"
		"path/filepath"
		"time"

		"github.com/chai2010/winsvc"
	)

	var (
		appPath string

		flagServiceName = flag.String("service-name", "hello-winsvc", "Set service name")
		flagServiceDesc = flag.String("service-desc", "hello windows service", "Set service description")

		flagServiceInstall   = flag.Bool("service-install", false, "Install service")
		flagServiceUninstall = flag.Bool("service-remove", false, "Remove service")
		flagServiceStart     = flag.Bool("service-start", false, "Start service")
		flagServiceStop      = flag.Bool("service-stop", false, "Stop service")

		flagHelp = flag.Bool("help", false, "Show usage and exit.")
	)

	func init() {
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, `Usage:
	hello [options]...

	Options:
	`)
			flag.PrintDefaults()
			fmt.Fprintf(os.Stderr, "%s\n", `
	Example:
	# run hello server
	$ go build -o hello.exe hello.go
	$ hello.exe

	# install hello as windows service
	$ hello.exe -service-install

	# start/stop hello service
	$ hello.exe -service-start
	$ hello.exe -service-stop

	# remove hello service
	$ hello.exe -service-remove

	# help
	$ hello.exe -h

	Report bugs to <chaishushan{AT}gmail.com>.`)
		}

		// change to current dir
		var err error
		if appPath, err = winsvc.GetAppPath(); err != nil {
			log.Fatal(err)
		}
		if err := os.Chdir(filepath.Dir(appPath)); err != nil {
			log.Fatal(err)
		}
	}

	func main() {
		flag.Parse()

		// install service
		if *flagServiceInstall {
			if err := winsvc.InstallService(appPath, *flagServiceName, *flagServiceDesc); err != nil {
				log.Fatalf("installService(%s, %s): %v\n", *flagServiceName, *flagServiceDesc, err)
			}
			fmt.Printf("Done\n")
			return
		}

		// remove service
		if *flagServiceUninstall {
			if err := winsvc.RemoveService(*flagServiceName); err != nil {
				log.Fatalln("removeService:", err)
			}
			fmt.Printf("Done\n")
			return
		}

		// start service
		if *flagServiceStart {
			if err := winsvc.StartService(*flagServiceName); err != nil {
				log.Fatalln("startService:", err)
			}
			fmt.Printf("Done\n")
			return
		}

		// stop service
		if *flagServiceStop {
			if err := winsvc.StopService(*flagServiceName); err != nil {
				log.Fatalln("stopService:", err)
			}
			fmt.Printf("Done\n")
			return
		}

		// run as service
		if !winsvc.InServiceMode() {
			log.Println("main:", "runService")
			if err := winsvc.RunAsService(*flagServiceName, StartServer, StopServer, false); err != nil {
				log.Fatalf("svc.Run: %v\n", err)
			}
			return
		}

		// run as normal
		StartServer()
	}

	func StartServer() {
		log.Println("StartServer, port = 8080")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "winsrv server", time.Now())
		})
		http.ListenAndServe(":8080", nil)
	}

	func StopServer() {
		log.Println("StopServer")
	}

BUGS

Report bugs to <chaishushan@gmail.com>.

Thanks!
*/
package winsvc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func GetAppPath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("winsvc.GetAppPath: %s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("winsvc.GetAppPath: %s is directory", p)
		}
	}
	return "", err
}

func InServiceMode() bool {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("winsvc.InServiceMode: svc.IsAnInteractiveSession(): err = %v", err)
	}
	return isIntSess
}

func InstallService(appPath, name, desc string, params ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("winsvc.InstallService: service %s already exists", name)
	}
	s, err = m.CreateService(name, appPath, mgr.Config{
		DisplayName: desc,
		StartType:   windows.SERVICE_AUTO_START,
	},
		params...,
	)
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("winsvc.InstallService: InstallAsEventCreate failed, err = %v", err)
	}
	return nil
}

func RemoveService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("winsvc.RemoveService: service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("winsvc.RemoveService: eventlog.Remove failed: %v", err)
	}
	return nil
}

func StartService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("winsvc.StartService: could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("p1", "p2", "p3")
	if err != nil {
		return fmt.Errorf("winsvc.StartService: could not start service: %v", err)
	}
	return nil
}

func StopService(name string) error {
	if err := controlService(name, svc.Stop, svc.Stopped); err != nil {
		return err
	}
	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("winsvc.controlService: could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("winsvc.controlService: could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("winsvc.controlService: timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("winsvc.controlService: could not retrieve service status: %v", err)
		}
	}
	return nil
}

var elog debug.Log

func RunAsService(name string, start, stop func(), isDebug bool) (err error) {
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	run := svc.Run
	if isDebug {
		run = debug.Run
	}

	elog.Info(1, fmt.Sprintf("winsvc.RunAsService: starting %s service", name))
	if err = run(name, &winService{Start: start, Stop: stop}); err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("winsvc.RunAsService: %s service stopped", name))
	return
}

type winService struct {
	Start func()
	Stop  func()
}

func (p *winService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	elog.Info(1, "winsvc.Execute:"+"begin")
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go p.Start()

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				elog.Error(1, fmt.Sprintf("winsvc.Execute:: unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	p.Stop()

	elog.Info(1, "winsvc.Execute:"+"end")
	return
}
