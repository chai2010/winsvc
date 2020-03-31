- *赞助 BTC: 1Cbd6oGAUUyBi7X7MaR4np4nTmQZXVgkCW*
- *赞助 ETH: 0x623A3C3a72186A6336C79b18Ac1eD36e1c71A8a6*
- *Go语言付费QQ群: 1055927514*

----

# Windows service

[![GoDoc](https://godoc.org/github.com/chai2010/winsvc?status.svg)](https://godoc.org/github.com/chai2010/winsvc)

## Install

`go get github.com/chai2010/winsvc`

## Example

```Go
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
```

BUGS
====

Report bugs to <chaishushan@gmail.com>.

Thanks!
