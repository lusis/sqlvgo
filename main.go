package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	var recsToPopulate int
	if x := os.Getenv("NUM_RECORDS"); len(x) == 0 {
		recsToPopulate = 100000
	} else {
		y, err := strconv.Atoi(x)
		if err != nil {
			log.Fatalf("invalid number: %s", x)
		}
		recsToPopulate = y
	}
	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := truncate(db); err != nil {
		log.Fatal(err.Error())
	}
	if err := populateDB(db, recsToPopulate); err != nil {
		log.Fatal(err.Error())
	}
}
