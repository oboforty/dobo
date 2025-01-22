package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var tomlData = `
name = "mydb"

[items]
key_type = "int32"
sort_key_type = "int32"
# "json" | "gop" | "in32" | "int64" | "float32" | "float64"
value_type = "json"

[partitions]
cluster_size = 1

[memtable]
# "redblack" | "avl" | "skiplist"
type = "redblack"

[sstable]
# "sst_io" | "parquet"
type = "sst_io"
# "summary" | "lru_cache" | "none"
lookup_aid = "lru_cache"

# TODO:
# crc
`

func git(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Simulate some work
	time.Sleep(5 * time.Second)

	select {
	case <-ctx.Done():
		fmt.Println("git: context cancelled")
	default:
		fmt.Println("git: work done")
	}
}

func main3() {
	// var wg sync.WaitGroup

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		fmt.Println("server: hello handler started")
		defer cancel()
		var wg sync.WaitGroup
		wg.Add(1)

		go git(ctx, &wg)

		wg.Wait()

		select {
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Println("server:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			// Respond to the client after the work is done
			fmt.Fprintf(w, "server: work done successfully\n")
			fmt.Println("server: handler completed successfully")
		}
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}
