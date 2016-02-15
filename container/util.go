package container

import (
	"fmt"
	"os"
)

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

func warn(err error) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
}
