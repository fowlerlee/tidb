package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func main() {
	// Initialize Viper
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %w", err))
	}

	// Retrieve environment variables
	dbHost := viper.GetString("DATABASE_HOST")
	dbPort := viper.GetString("DATABASE_PORT")
	dbUser := viper.GetString("DATABASE_USER")
	viper.GetString("DATABASE_PASSWORD")
	dbName := viper.GetString("DATABASE_NAME")

	// Example usage
	fmt.Printf("Connecting to database %s at %s:%s with user %s\n", dbName, dbHost, dbPort, dbUser)
	// ...existing code...
}
