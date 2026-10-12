package main

import (
    "fmt"
    "os"

    "github.com/leo-jcn/du-split/internal/app"
)

func main() {
    if err := app.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "du-split: %v\n", err)
        os.Exit(1)
    }
}
