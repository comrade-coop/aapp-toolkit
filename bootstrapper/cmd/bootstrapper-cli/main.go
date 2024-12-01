package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Bootstrapper CLI")
    if len(os.Args) < 2 {
        fmt.Println("Usage: bootstrapper-cli <command>")
        os.Exit(1)
    }
}
