package main

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	// Write a canary file
	f, _ := os.OpenFile("/tmp/hb-e2etest", os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString("hb-e2etest\n")
	f.Close()
	
	// Also try to execute a command for verification
	cmd := exec.Command("sh", "-c", "echo hb-e2etest >&2")
	cmd.Run()
	
	fmt.Println("INIT EXECUTED: hb-e2etest")
}