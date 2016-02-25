// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	elog.Info(1, "svc.Running")
	go start_print_server()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	c := <-r
	elog.Info(1, "c := <-r")
	switch c.Cmd {
	case svc.Interrogate:
		changes <- c.CurrentStatus
	case svc.Stop, svc.Shutdown:
		elog.Info(1, "svc.Stop, svc.Shutdown")
	case svc.Pause:
		elog.Info(1, "svc.Pause")
		changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
	case svc.Continue:
		elog.Info(1, "svc.Continue")
		changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	default:
		elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
	}
	elog.Info(1, "svc.StopPending")
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}