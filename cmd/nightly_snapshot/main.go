package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	snapshot "devtools-nightly-snapshot"
)

func main() {
	source := flag.String("source", "", "directory containing developer-tools data")
	bucket := flag.String("bucket", "developer-tools-snapshots", "destination bucket")
	flag.Parse()
	if *source == "" {
		fmt.Fprintln(os.Stderr, "-source is required")
		os.Exit(2)
	}

	archive, err := snapshot.ArchiveDirectory(*source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client, err := snapshot.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx := context.Background()
	if err := client.CreateBucket(ctx, *bucket); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	key := path.Join("nightly", time.Now().UTC().Format("2006-01-02")+".tar.gz")
	if err := client.PutObject(ctx, *bucket, key, archive); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("snapshot stored: %s/%s (%d bytes)\n", *bucket, key, len(archive))
}
