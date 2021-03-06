package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	kf "github.com/lsds/KungFu/srcs/go/kungfu"
	"github.com/lsds/KungFu/srcs/go/utils"
)

var (
	runFor     = flag.Duration("run-for", 30*time.Second, "")
	errorAfter = flag.Duration("error-after", 5*time.Second, "")
)

func main() {
	flag.Parse()
	kungfu, err := kf.New()
	if err != nil {
		utils.ExitErr(err)
	}
	kungfu.Start()
	defer kungfu.Close()
	rank := kungfu.CurrentSession().Rank()
	fmt.Printf("OK, rank=%d.\n", rank)
	fmt.Fprintf(os.Stderr, "Err, rank=%d!\n", rank)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if rank == 0 {
		ctx, cancel = context.WithTimeout(ctx, *errorAfter)
	}
	done := time.After(*runFor)
	select {
	case <-ctx.Done():
		os.Exit(1)
	case <-done:
		return
	}
}
