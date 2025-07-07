package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/s5i/goutil/authn"
	"github.com/s5i/goutil/version"
	"github.com/s5i/shortlink/database"
)

var (
	configPath = flag.String("config_path", "shortlink.yaml", "Path to config file.")
	fVersion   = flag.Bool("version", false, "When true, print version and exit.")
)

func main() {
	flag.Parse()

	if *fVersion {
		fmt.Fprintln(os.Stderr, version.Get())
		os.Exit(0)
	}

	cfg, err := ReadConfig(*configPath)
	if err != nil {
		log.Printf("error reading config: %v", err)
		os.Exit(1)
	}

	db, err := database.NewBolt(cfg.DatabasePath)
	if err != nil {
		log.Printf("error opening database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.UpdateAdmins(cfg.Admins); err != nil {
		log.Printf("failed to update admin list: %v", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	authn, err := authn.New(
		authn.OptClientID(cfg.OAuthClientID),
		authn.OptClientSecret(cfg.OAuthClientSecret),
		authn.OptJWTTTL(cfg.JWTTTL),
		authn.OptHostname(cfg.Hostname),
		authn.OptMux(mux),
		authn.OptCallbackPath("/auth/callback"),
	)
	if err != nil {
		log.Printf("failed to create authn: %v", err)
		os.Exit(1)
	}

	mux.HandleFunc("/", GetLink(db, cfg.DefaultRedirectURL))
	mux.HandleFunc("/admin/edit", EditLink(authn, db))
	mux.HandleFunc("/admin/list", ListLinks(authn, db))
	mux.HandleFunc("/admin/users", EditUsers(authn, db))

	if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
		log.Printf("failed to listen: %v", err)
		os.Exit(1)
	}
}
