package main

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql" // import the dialect
	_ "github.com/go-sql-driver/mysql"               // mysql driver
	"github.com/google/uuid"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	selectBase = "select id, name, rtype, rstate, created_at, updated_at FROM testdata"
)

var (
	idInts  = []int{0, 1, 2, 3, 4, 5, 6, 7}
	queries = map[string]string{
		"truncate":         "truncate table testdata",
		"select_all":       selectBase,
		"select_all_index": selectBase + " ORDER BY counter",
		"select_where":     selectBase + " WHERE ",
	}
)

// Record ...
type Record struct {
	UUID      string
	Name      string
	Rtype     int
	Rstate    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func randInt() int {
	rand.Seed(time.Now().UnixNano())
	num := idInts[rand.Intn(len(idInts))]
	return num
}

func randName() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 191)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// nolint: deadcoce
func timeFromFloat64(ts float64) time.Time {
	secs := int64(ts)
	nsecs := int64((ts - float64(secs)) * 1e9)
	return time.Unix(secs, nsecs).UTC()
}

func scanResults(rows *sql.Rows) ([]*Record, error) {
	results := make([]*Record, 0)
	for rows.Next() {
		var uuid string
		var name string
		var rtype int
		var rstate int
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&uuid,
			&name,
			&rtype,
			&rstate,
			&createdAt,
			&updatedAt,
		); err != nil {
			rows.Close() // nolint: errcheck
			return nil, err
		}
		results = append(results, &Record{
			UUID:      uuid,
			Name:      name,
			Rtype:     rtype,
			Rstate:    rstate,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close() // nolint: errcheck
	return results, nil
}

func selectAll(db *sql.DB) ([]*Record, error) {
	stmt, err := db.Prepare(queries["select_all"])
	if err != nil {
		return nil, err
	}
	defer stmt.Close() // nolint: errcheck
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint: errcheck
	return scanResults(rows)
}

func selectAllIndex(db *sql.DB) ([]*Record, error) {
	stmt, err := db.Prepare(queries["select_all_index"])
	if err != nil {
		return nil, err
	}
	defer stmt.Close() // nolint: errcheck
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint: errcheck
	return scanResults(rows)
}

func selectSomeIndexFilterSQL(db *sql.DB, ints []int) ([]*Record, error) {
	placeholders := make([]string, 0)
	params := make([]interface{}, 0)
	// iterate once for rstate
	for _, i := range ints {
		params = append(params, i)
	}
	// iterate again for rtype
	for _, i := range ints {
		params = append(params, i)
	}
	for i := 1; i <= len(ints); i++ {
		placeholders = append(placeholders, "?")
	}
	query := queries["select_where"]
	query = query + "rstate IN (" + strings.Join(placeholders, ",") + ")"
	query = query + " AND "
	query = query + "rtype IN (" + strings.Join(placeholders, ",") + ") ORDER BY counter"
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func selectSomeFilterSQL(db *sql.DB, ints []int) ([]*Record, error) {
	placeholders := make([]string, 0)
	params := make([]interface{}, 0)
	// iterate once for rstate
	for _, i := range ints {
		params = append(params, i)
	}
	// iterate again for rtype
	for _, i := range ints {
		params = append(params, i)
	}
	for i := 1; i <= len(ints); i++ {
		placeholders = append(placeholders, "?")
	}
	query := queries["select_where"]
	query = query + "rstate IN (" + strings.Join(placeholders, ",") + ")"
	query = query + " AND "
	query = query + "rtype IN (" + strings.Join(placeholders, ",") + ")"
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func selectSomeIndexFilterGoqu(db *sql.DB, ints []int) ([]*Record, error) {
	dialect := goqu.Dialect("mysql")
	intints := make([]interface{}, len(ints))
	for x := range ints {
		intints[x] = ints[x]
	}
	query, params, err := dialect.From("testdata").Prepared(true).Select(
		goqu.C("id"),
		goqu.C("name"),
		goqu.C("rtype"),
		goqu.C("rstate"),
		goqu.C("created_at"),
		goqu.C("updated_at"),
	).Where(
		goqu.C("rstate").In(intints...),
		goqu.C("rtype").In(intints...),
	).Order(goqu.C("counter").Asc()).ToSQL()
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func selectSomeFilterGo(db *sql.DB, ints []int) ([]*Record, error) {
	stmt, err := db.Prepare(queries["select_all"])
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer stmt.Close() // nolint: errcheck
	defer rows.Close() // nolint: errcheck
	results, err := scanResults(rows)
	if err != nil {
		return nil, err
	}
	return filterGoResults(results, ints), nil
}

func selectSomeIndexFilterGo(db *sql.DB, ints []int) ([]*Record, error) {
	stmt, err := db.Prepare(queries["select_all_index"])
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results, err := scanResults(rows)
	if err != nil {
		return nil, err
	}

	return filterGoResults(results, ints), nil
}

func filterGoResults(recs []*Record, ints []int) []*Record {
	// some placeholders
	matchedRstate := make(map[*Record]struct{}, 0)
	matchedRtype := make(map[*Record]struct{}, 0)
	finalResults := make([]*Record, 0)

	for _, x := range ints {
		for y := range recs {
			if recs[y].Rstate == x {
				matchedRstate[recs[y]] = struct{}{}
			}
			if recs[y].Rtype == x {
				matchedRtype[recs[y]] = struct{}{}
			}
		}
	}
	// loop over the records that matched rstate
	for k := range matchedRstate {
		// see if it also matched rtype
		_, ok := matchedRtype[k]
		// if it matched rtype as well and we're using AND logic for each requirement
		// add to final list
		if ok {
			finalResults = append(finalResults, k)
			continue // move to next entry in matching rstate
		}
	}
	matchedRstate = make(map[*Record]struct{}, 0)
	matchedRtype = make(map[*Record]struct{}, 0)
	return finalResults
}

func makeRecord() Record {
	id := uuid.New().String()
	rstate := randInt()
	rtype := randInt()
	name := randName()
	rec := Record{
		UUID:   id,
		Name:   name,
		Rtype:  rtype,
		Rstate: rstate,
	}
	return rec
}

func connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/testing?parseTime=true")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Second)
	return db, nil
}

func populateDB(db *sql.DB, numRecs int) error {
	for i := 1; i <= numRecs; i++ {
		_, err := addRecord(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func truncate(db *sql.DB) error {
	_, err := db.Exec(queries["truncate"])
	return err
}

func addRecord(db *sql.DB) (int64, error) {
	query := `INSERT INTO testdata
	(id, name, rtype, rstate)
	VALUES
	(?, ?, ?, ?)
	`
	record := makeRecord()
	res, err := db.Exec(query,
		record.UUID,
		record.Name,
		record.Rtype,
		record.Rstate,
	)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}
