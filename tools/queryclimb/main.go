package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("queryclimb", flag.ContinueOnError)
	root := flags.String("root", ".", "repository to read; never modified except by explicit sqlc command")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() < 1 {
		return errors.New("command required: inventory, init, clone, seed, inspect, baseline, gate, " +
			"selfcheck, sqlc, attempt, record, dashboard, fidelity or exec")
	}
	rootPath, err := filepath.Abs(*root)
	if err != nil {
		return err
	}

	command := flags.Arg(0)
	lock, err := os.OpenFile("/tmp/tadoku-query-climb.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	lockMode := syscall.LOCK_EX
	if command == "dashboard" {
		lockMode |= syscall.LOCK_NB
	}
	if err := syscall.Flock(int(lock.Fd()), lockMode); err != nil {
		return errors.New("another harness operation holds the measurement/build/load lock")
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	rest := flags.Args()[1:]
	switch command {
	case "attempt":
		return attempt(ctx, rootPath, rest)
	case "record":
		return recordAttempt(rest)
	case "dashboard":
		return dashboard(rest)
	case "fidelity":
		return fidelity(rest)
	case "gate":
		return gate(rest)
	case "clone":
		if len(rest) != 2 {
			return errors.New("clone requires template and new name")
		}
		return clone(ctx, rest[0], rest[1])
	case "exec":
		if len(rest) == 0 {
			return errors.New("exec requires a command")
		}
		return shell(ctx, rootPath, rest[0], rest[1:]...)
	case "inventory":
		queries, err := inventory(rootPath)
		if err != nil {
			return err
		}
		return output(queries)
	case "init":
		return initialize(ctx, rootPath, rest)
	case "seed":
		return seed(ctx, rest)
	case "inspect":
		return inspect(ctx, rootPath, rest)
	case "baseline":
		return baseline(ctx, rootPath, rest)
	case "selfcheck":
		return selfcheck(ctx, rootPath, rest)
	case "sqlc":
		return sqlcCheck(ctx, rootPath)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func output(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
