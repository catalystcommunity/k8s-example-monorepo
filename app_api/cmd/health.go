package cmd

import (
	"database/sql"
	"io/fs"
	"regexp"
	"strconv"

	"github.com/catalystcommunity/app-utils-go/errorutils"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

// ExpectedVersion is the highest migration version expected to be applied
var expectedVersion = getHighestVersionFromEmbeddedMigrations()

// GetExpectedMigrationVersion returns the highest migration version that should be applied
func GetExpectedMigrationVersion() int64 {
	return expectedVersion
}
var migrationsComplete = false

// TODO: Implement a health check that is only "healthy" when DB migrations are done.

// compares the parsed maximum db migration version against the database's current migration version, returns true
// if they are equal
func migrationsAreComplete() bool {
	if migrationsComplete {
		// immediately return if we know the migrations are complete
		return true
	}
	// connect to the db
	sqldb, err := sql.Open("postgres", config.DbUri)
	if err != nil {
		errorutils.LogOnErr(nil, "error getting sql db from gorm db", err)
		return false
	}
	defer sqldb.Close()
	// get the current version from the db
	var currentVersion int64
	if currentVersion, err = goose.GetDBVersion(sqldb); err != nil {
		return false
	}
	// set the global migrationsComplete variable for reference for the next health check. This avoids needlessly
	// pinging the db on every health check
	migrationsComplete = expectedVersion == currentVersion
	if !migrationsComplete {
		// log error for visibility on readiness
		logging.Log.WithFields(logrus.Fields{"expected_version": expectedVersion, "current_version": currentVersion}).Error("readiness check failed: database migrations are not complete")
	}
	return migrationsComplete
}

// parses the embedded migrations directory for migration files and returns the highest version number
func getHighestVersionFromEmbeddedMigrations() (highestVersion int64) {
	goose.SetBaseFS(migrations)
	var files []fs.DirEntry
	var err error
	if files, err = migrations.ReadDir("migrations"); err != nil {
		errorutils.LogOnErr(nil, "error reading embedded migrations", err)
		return
	}

	pattern := regexp.MustCompile("(\\d+)")
	for _, file := range files {
		var version int64
		capture := pattern.Find([]byte(file.Name()))
		if version, err = strconv.ParseInt(string(capture), 10, 32); err != nil {
			errorutils.LogOnErr(nil, "error getting migration version from file", err)
			return
		}
		if version > highestVersion {
			highestVersion = version
		}
	}
	return
}
