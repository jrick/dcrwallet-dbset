// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

var (
	dbPath  = flag.String("db", "", "path to dcrwallet db")
	nflag   = flag.Bool("n", false, "don't write new version (only logs old version)")
	version = flag.Uint("version", 0, "version to set db to")
)

func main() {
	flag.Parse()
	version := uint32(*version)
	dbPath := *dbPath
	if version == 0 && !*nflag {
		log.Fatal("-version is required and must be positive")
	}
	if dbPath == "" {
		log.Fatal("-db is required")
	}
	if _, err := os.Stat(dbPath); err != nil {
		log.Fatalf("-db: %v", err)
	}
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, version)
	err = db.Update(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte("meta"))
		if meta == nil {
			return errors.New("missing meta bucket")
		}
		vers := meta.Get([]byte("ver"))
		if len(vers) != 4 {
			log.Print("warning: missing version")
		} else {
			old := binary.BigEndian.Uint32(vers)
			log.Printf("prior version: %v", old)
		}
		if *nflag {
			return nil
		}
		log.Printf("setting db version to %v", version)
		return meta.Put([]byte("ver"), versionBytes)
	})
	if err != nil {
		log.Fatalf("update: %v", err)
	}
	err = db.Close()
	if err != nil {
		log.Fatalf("close: %v", err)
	}
}
