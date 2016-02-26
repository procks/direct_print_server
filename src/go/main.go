package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
		"usage: %s <command>\n"+
		"       where <command> is one of\n"+
		"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	const svcName = "DirectPrintServer"

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "start":
		runService(svcName, true)
		return
	case "install_s":
		err = installService(svcName, "Direct Print Server")
	case "remove_s":
		err = removeService(svcName)
	case "start_s":
		err = startService(svcName)
	case "stop_s":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	case "pause_s":
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue_s":
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}