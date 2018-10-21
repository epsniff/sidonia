package jstag

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

/*Simple static asset service to import github.com/lytics/jstags into Go projects
so vendoring tools can lock to a jstag checkout. */

func outPath() string {
	flg := flag.String("outpath", "", "Path to .../jstag/out/ directory containing files to be served")
	flag.Parse()

	log.Info(fmt.Sprintf("flg: %s", *flg))
	if *flg != "" {

		return *flg
	}
	log.Error("-outpath not set!")
	os.Exit(1)
	return ""
}

func ServeOut() {

	dir := outPath()
	log.Infof("Out Directory: `%s`", dir)

	log.Info("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir(dir))))

}
