package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ahui2016/txt/mydb"
	"github.com/ahui2016/txt/util"
)

const (
	AppConfigFolder = "github-ahui2016/txt"
	dbFileName      = "db-txt.bolt"
)

var (
	db       = new(mydb.DB)
	addr     = flag.String("addr", "127.0.0.1:8000", "Local IP address. Example: 127.0.0.1:8000")
	debug    = flag.Bool("debug", false, "Switch to debug mode.")
	demo     = flag.Bool("demo", false, "Set this flag for demo.")
	dbFolder = flag.String("db", "", "Specify a folder for the database.")
)

func init() {
	flag.Parse()
	dbPath := getDBPath()
	fmt.Println("[Database]", dbPath)

	util.Panic(db.Open(dbPath))
}

func getDBPath() string {
	if *dbFolder != "" {
		folder, err := filepath.Abs(*dbFolder)
		util.Panic(err)
		if util.PathIsNotExist(folder) {
			log.Fatal("Not Found: " + folder)
		}
		return filepath.Join(folder, dbFileName)
	}
	userConfigDir, err := os.UserConfigDir()
	util.Panic(err)
	configFolder := filepath.Join(userConfigDir, AppConfigFolder)
	util.Panic(os.MkdirAll(configFolder, 0740))
	return filepath.Join(configFolder, dbFileName)
}
