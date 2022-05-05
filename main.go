package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joshuarubin/lifecycle"
)

func main() {
	ctx := lifecycle.New(context.Background())

	if err := newApp().run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
