package config

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
}

func LoadConfig() *DBConfig {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    port, err := strconv.Atoi(os.Getenv("DB_PORT"))
    if err != nil {
        log.Fatal("Error parsing DB_PORT:", err)
    }

    return &DBConfig{
        Host:     os.Getenv("DB_HOST"),
        Port:     port,
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
    }
}