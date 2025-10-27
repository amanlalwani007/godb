package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"godb/kv"
)

func help() {
	fmt.Println("commands:")
	fmt.Println("  set <key> <value>")
	fmt.Println("  get <key>")
	fmt.Println("  del <key>")
	fmt.Println("  compact")
	fmt.Println("  exit")
}

func main() {
	db, err := kv.NewKV("db.log")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	fmt.Println("godb CLI — simple append-only log backed KV")
	help()
	in := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			fmt.Print("> ")
			continue
		}
		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])
		switch cmd {
		case "help":
			help()
		case "set":
			if len(parts) < 3 {
				fmt.Println("usage: set <key> <value>")
			} else {
				key := parts[1]
				value := strings.Join(parts[2:], " ")
				if err := db.Set(key, []byte(value)); err != nil {
					fmt.Printf("set error: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}
		case "get":
			if len(parts) != 2 {
				fmt.Println("usage: get <key>")
			} else {
				key := parts[1]
				if val, ok := db.Get(key); ok {
					fmt.Printf("%s\n", string(val))
				} else {
					fmt.Println("(nil)")
				}
			}
		case "del":
			if len(parts) != 2 {
				fmt.Println("usage: del <key>")
			} else {
				key := parts[1]
				if err := db.Del(key); err != nil {
					fmt.Printf("del error: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}
		case "compact":
			fmt.Println("Compacting log...")
			if err := db.Compact(); err != nil {
				fmt.Printf("compact error: %v\n", err)
			} else {
				fmt.Println("Compact done.")
			}
		case "exit", "quit":
			fmt.Println("bye")
			return
		default:
			fmt.Println("unknown command:", cmd)
			help()
		}
		fmt.Print("> ")
	}
	if err := in.Err(); err != nil {
		log.Printf("input error: %v", err)
	}
}
