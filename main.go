package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/s5i/shortlink/auth"
	"github.com/s5i/shortlink/config"
	"github.com/s5i/shortlink/database"
)

var (
	dbPath       = flag.String("db_path", "shortlink.db", "Path to database file.")
	configPath   = flag.String("config_path", "shortlink.yaml", "Path to config file.")
	createConfig = flag.Bool("create_config", false, "If true, create example config file and exit.")
)

func main() {
	flag.Parse()

	if *createConfig {
		if err := config.CreateExample(*configPath); err != nil {
			log.Printf("error creating config: %v", err)
			os.Exit(3)
		}
		os.Exit(0)
	}

	cfg, err := config.Read(*configPath)
	if err != nil {
		log.Printf("error reading config: %v", err)
		os.Exit(1)
	}

	db, err := database.NewBolt(*dbPath)
	if err != nil {
		log.Printf("error opening database: %v", err)
		os.Exit(2)
	}
	defer db.Close()

	if err := db.UpdateAdmins(cfg.Admins); err != nil {
		log.Printf("failed to update admin list: %v", err)
		os.Exit(3)
	}

	mux := http.NewServeMux()

	auth := auth.New(cfg.OAuthClientID, cfg.OAuthClientSecret, cfg.JWTSecret, cfg.JWTTTL, cfg.Hostname, mux)
	mux.HandleFunc("/", GetLink(db, cfg.DefaultRedirectURL))
	mux.HandleFunc("/admin/edit", EditLink(auth, db))
	mux.HandleFunc("/admin/list", ListLinks(auth, db))
	mux.HandleFunc("/admin/users", EditUsers(auth, db))

	if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
		log.Printf("failed to listen: %v", err)
		os.Exit(4)
	}
}
