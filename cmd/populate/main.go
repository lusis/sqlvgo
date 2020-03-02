package main

import (
	"log"
	"os"
	"strconv"

	"github.com/lusis/sqlvgo"
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
	_, err := sqlvgo.Populate(recsToPopulate, true)
	if err != nil {
		log.Fatal(err.Error())
	}
}
