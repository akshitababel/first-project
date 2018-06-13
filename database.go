/*
Package data - Handles functions related to data source access e.g. cache, databases
*/
package data

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// SportsDb is the pointer to the iSports database resource.
var SportsDb *sql.DB
var SportsDbAU *sql.DB

// UserDb is the pointer to the iSports Users database resource.
var UserDb *sql.DB

// LogDb is the pointer to the iSports Logs database resource.
var LogDb *sql.DB

// GeniusStatsDb is the pointer to the geniusstats database resource.
var GeniusStatsDb *sql.DB

// AEST : set dafault time zone
var AEST *time.Location

// InitDB initialises the database pools with
func InitDB(host, port, user, password, hostAU, passAU string) (sportsDb, sportsDbAU, userDb *sql.DB, logDb *sql.DB, geniusStatsDb *sql.DB, err error) {
	SportsDb, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/isports")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err = SportsDb.Ping(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	SportsDbAU, err = sql.Open("mysql", user+":"+passAU+"@tcp("+hostAU+":"+port+")/isports")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err = SportsDbAU.Ping(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	UserDb, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/isports_users")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err = UserDb.Ping(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	LogDb, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/isports_logs")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err = LogDb.Ping(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	GeniusStatsDb, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/geniusstats")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err = GeniusStatsDb.Ping(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	AEST, _ = time.LoadLocation("Australia/Melbourne")
	return SportsDb, SportsDbAU, UserDb, LogDb, GeniusStatsDb, nil
}

// StringArray returns a cleaned up string array version of the supplied sql.Rows data.
func StringArray(rows *sql.Rows) (results [][]string) {
	colNames, err := rows.Columns()
	if err != nil {
		return nil
	}

	interfaceCols := make([]interface{}, len(colNames))
	stringCols := make([]sql.NullString, len(colNames))
	for i, _ := range stringCols {
		interfaceCols[i] = &stringCols[i]
	}

	for rows.Next() {
		err := rows.Scan(interfaceCols...)
		if err != nil {
			return nil
		}
		rowData := make([]string, len(colNames))
		for i := 0; i < len(colNames); i++ {
			rowData[i] = interfaceCols[i].(*sql.NullString).String
		}
		results = append(results, rowData)
	}
	return results
}
