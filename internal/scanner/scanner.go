package scanner

import "fmt"

func Scan(path string) error {
    fmt.Printf("scanning %s\n", path)
    return nil
}
