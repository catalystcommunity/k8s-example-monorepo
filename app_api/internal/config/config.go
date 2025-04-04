package config

import (
	"github.com/catalystcommunity/app-utils-go/env"
)

var (
	// DbUri is the database connection string
	DbUri string
	
	// Port is the HTTP server port
	Port int
	
	// CommitOnSuccess determines if transactions should be committed on successful responses (2xx status)
	// Default is true, but can be set to false for testing environments
	CommitOnSuccess = env.GetEnvAsBoolOrDefault("COMMIT_ON_SUCCESS", "true")
)
