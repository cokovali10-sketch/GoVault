package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cokovali10-sketch/GoVault/vault"
)

func main() {
	file := flag.String("file", ".govault.db", "Vault file path")
	pass := flag.String("pass", "", "Password (required)")
	flag.Parse()

	if *pass == "" {
		fmt.Fprintln(os.Stderr, "Error: -pass is required")
		flag.Usage()
		os.Exit(1)
	}

	v := vault.NewVault(*file, *pass)
	if err := v.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	switch args[0] {
	case "set":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: govault set <key> <value> [tags...]")
			os.Exit(1)
		}
		tags := args[3:]
		if err := v.Set(args[1], []byte(args[2]), tags...); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := v.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  Set '%s' (%d bytes)\n", args[1], len(args[2]))

	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: govault get <key>")
			os.Exit(1)
		}
		val, err := v.Get(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(val))

	case "rm", "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: govault rm <key>")
			os.Exit(1)
		}
		if v.Delete(args[1]) {
			v.Save()
			fmt.Printf("  Deleted '%s'\n", args[1])
		} else {
			fmt.Fprintf(os.Stderr, "Key '%s' not found\n", args[1])
		}

	case "list", "ls":
		keys := v.List()
		if len(keys) == 0 {
			fmt.Println("  Vault is empty.")
		} else {
			for _, k := range keys {
				entry, _ := v.GetEntry(k)
				tags := ""
				if len(entry.Tags) > 0 {
					tags = " [" + strings.Join(entry.Tags, ", ") + "]"
				}
				fmt.Printf("  %-30s  %d bytes%s\n", k, len(entry.Value), tags)
			}
		}

	case "search":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: govault search <query>")
			os.Exit(1)
		}
		results := v.Search(args[1])
		if len(results) == 0 {
			fmt.Println("  No results.")
		} else {
			for _, k := range results {
				fmt.Printf("  %s\n", k)
			}
		}

	case "stats":
		stats := v.Stats()
		fmt.Printf("  Entries:    %v\n", stats["entries"])
		fmt.Printf("  Total size: %v bytes\n", stats["total_size"])
		fmt.Printf("  Created:    %v\n", stats["created_at"])
		fmt.Printf("  Updated:    %v\n", stats["updated_at"])

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`GoVault — Encrypted local storage

Usage:
  govault -pass <password> set <key> <value> [tags...]
  govault -pass <password> get <key>
  govault -pass <password> rm <key>
  govault -pass <password> list
  govault -pass <password> search <query>
  govault -pass <password> stats`)
}
