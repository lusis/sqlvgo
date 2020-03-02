package main

import (
	"database/sql"
	"fmt"
	"testing"
)

var testResults []*Record

var testWantedInts = []int{1, 3, 5, 7}

var testRecCounts = []int{50, 100, 1000, 5000, 10000, 50000, 100000}

func testSetup(db *sql.DB, numRecords int) error {
	if err := truncate(db); err != nil {
		return err
	}
	if err := populateDB(db, numRecords); err != nil {
		return err
	}
	return nil
}

func flush(db *sql.DB) error {
	_, err := db.Exec("flush tables")
	return err
}

func BenchmarkSelectAll(b *testing.B) {
	b.StopTimer()
	db, err := connect()
	if err != nil {
		b.Fatalf(err.Error())
	}
	defer db.Close()
	for _, numRecords := range testRecCounts {
		err := testSetup(db, numRecords)
		if err != nil {
			b.Fatal(err.Error())
		}

		b.Run(fmt.Sprintf("RecordCount_%d", numRecords), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				res, err := selectAll(db)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				testResults = res
				flush(db)
			}
		})
	}
}
func BenchmarkSelectAllIndex(b *testing.B) {
	b.StopTimer()
	db, err := connect()
	if err != nil {
		b.Fatalf(err.Error())
	}
	defer db.Close()
	for _, numRecords := range testRecCounts {
		err := testSetup(db, numRecords)
		if err != nil {
			b.Fatal(err.Error())
		}
		b.Run(fmt.Sprintf("RecordCount_%d", numRecords), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				res, err := selectAllIndex(db)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				testResults = res
				flush(db)
			}
		})
	}
}

func BenchmarkSelectSomeIndexed(b *testing.B) {
	b.StopTimer()
	db, err := connect()
	if err != nil {
		b.Fatalf(err.Error())
	}
	defer db.Close()
	for _, numRecords := range testRecCounts {
		// populate db once for each record count iteration
		err := testSetup(db, numRecords)
		if err != nil {
			b.Fatal(err.Error())
		}
		var sqlRecords int
		var goRecords int
		var goquRecords int

		b.Run(fmt.Sprintf("SomeIndexedSQL/RecordCount_%d", numRecords), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				res, err := selectSomeIndexFilterSQL(db, testWantedInts)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				sqlRecords = len(res)
				testResults = res
				flush(db)
			}
		})

		b.Run(fmt.Sprintf("SomeIndexedGoqu/RecordCount_%d", numRecords), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				res, err := selectSomeIndexFilterGoqu(db, testWantedInts)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				goquRecords = len(res)
				testResults = res
				flush(db)
			}
		})

		b.Run(fmt.Sprintf("SomeIndexedGo/RecordCount_%d", numRecords), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				res, err := selectSomeIndexFilterGo(db, testWantedInts)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				testResults = res
				flush(db)
				goRecords = len(res)
			}
		})
		// some sanity checks
		if sqlRecords != goquRecords {
			b.Fatalf("sql version returned %d but goqu version returned %d. counts should match", sqlRecords, goquRecords)
		}
		if goRecords != sqlRecords {
			b.Fatalf("sql version returned %d but go version returned %d. counts should match", sqlRecords, goRecords)
		}
	}
}

func Benchmark100kSelectSomeIndexed(b *testing.B) {
	b.StopTimer()
	db, err := connect()
	if err != nil {
		b.Fatalf(err.Error())
	}
	defer db.Close()
	var sqlRecords int
	var goRecords int
	var goquRecords int

	b.Run("SQL", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			b.StartTimer()
			res, err := selectSomeIndexFilterSQL(db, testWantedInts)
			b.StopTimer()
			if err != nil {
				b.Fatal(err)
			}
			sqlRecords = len(res)
			testResults = res
			flush(db)
		}
	})

	b.Run("Goqu", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			b.StartTimer()
			res, err := selectSomeIndexFilterGoqu(db, testWantedInts)
			if err != nil {
				b.Fatal(err)
			}
			b.StopTimer()
			goquRecords = len(res)
			testResults = res
			flush(db)
		}
	})

	b.Run("Go", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			b.StartTimer()
			res, err := selectSomeIndexFilterGo(db, testWantedInts)
			b.StopTimer()
			if err != nil {
				b.Fatal(err)
			}
			testResults = res
			flush(db)
			goRecords = len(res)
		}
	})
	if goRecords != sqlRecords {
		b.Fatalf("sql version returned %d but go version returned %d. counts should match", sqlRecords, goRecords)
	}
	if sqlRecords != goquRecords {
		b.Fatalf("sql version returned %d but goqu version returned %d. counts should match", sqlRecords, goquRecords)
	}
}
