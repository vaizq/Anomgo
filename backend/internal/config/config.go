package config

import (
	"flag"
	"os"
)

var (
	Addr          string
	InternalAddr  string
	DSN           string
	MoneropayURL  string
	CssDir        string
	UploadDir     string
	StaticDir     string
	PgpPrivateKey string
)

func Parse() {
	flag.StringVar(&Addr, "addr", "localhost:4000", "http address for server to listen")
	flag.StringVar(&CssDir, "css-dir", "./ui/css/", "directory from where css files are served")
	flag.StringVar(&UploadDir, "upload-dir", "./uploads/", "directory where uploaded images are stored")
	flag.StringVar(&StaticDir, "static-dir", "./static/", "directory where static files are stored")
	flag.StringVar(&DSN, "dsn", os.Getenv("DSN"), "postgres data source name")
	flag.StringVar(&InternalAddr, "internal-addr", "0.0.0.0:4420", "internal address to listen")
	flag.StringVar(&MoneropayURL, "moneropay-url", "http://localhost:5000", "moneropay url")
	flag.StringVar(&PgpPrivateKey, "PGP-private-key-file", os.Getenv("PGP-private-key-file"), "pgp private key file")
	flag.Parse()
}
