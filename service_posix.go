// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package winsvc

import (
	"fmt"
	"os"
	"path/filepath"
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
	panic("winsvc: only support windows!")
}
func InstallService(appPath, name, desc string, params ...string) error {
	panic("winsvc: only support windows!")
}
func RemoveService(name string) error {
	panic("winsvc: only support windows!")
}
func RunAsService(name string, start, stop func(), isDebug bool) (err error) {
	panic("winsvc: only support windows!")
}
func StartService(name string) error {
	panic("winsvc: only support windows!")
}
func StopService(name string) error {
	panic("winsvc: only support windows!")
}
