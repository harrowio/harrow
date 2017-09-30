package projector

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/clock"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/stores"
)

const ProgramName = "projector"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	listen := flag.String("listen", "0.0.0.0:8888", "interface and post to listen on")
	storage := flag.String("storage", "memory://", "where to store data; use bolt://filename for persistent storage")
	flag.Parse()

	structuredLog := NewStructuredLog(100, clock.System)
	index := (Index)(nil)

	storageURL, err := url.Parse(*storage)
	if err != nil {
		log.Fatal().Msgf("storage: " + err.Error())
	}
	switch storageURL.Scheme {
	case "memory":
		log.Info().Msg("storage=memory")
		index = NewInMemoryIndex()
	case "bolt":
		path := fmt.Sprintf("%s%s", storageURL.Host, storageURL.Path)
		dbFile, err := filepath.Abs(path)
		if err != nil {
			log.Fatal().Msgf("storage=bolt err=%s", err)
		}

		log.Info().Msgf("storage=%s", dbFile)
		boltdb, err := bolt.Open(dbFile, 0644, nil)
		if err != nil {
			log.Fatal().Msgf("storage: bolt: " + err.Error())
		}
		index = NewBoltDBIndex(boltdb)
	default:
		log.Fatal().Msgf("unknown storage driver: " + storageURL.String())
	}

	projector := NewProjector(index, log)
	db, err := config.GetConfig().DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	updateIndex := func() {
		tx := db.MustBegin()
		defer tx.Rollback()
		activityStore := stores.NewDbActivityStore(tx)
		projector.Update(activityStore)
	}

	log.Info().Msgf("subscribed to %v", projector.SubscribedTo())
	http.HandleFunc("/_/log", func(w http.ResponseWriter, req *http.Request) {
		format := "text"
		if req.FormValue("format") != "" {
			format = "format"
		} else {
			if strings.Contains(req.Header.Get("Accept"), "text/html") {
				format = "html"
			}
		}

		switch format {
		case "text":
			w.Header().Set("Content-Type", "text/plain")
			structuredLog.RenderText(w)
		case "html":
			w.Header().Set("Content-Type", "text/html")
			structuredLog.RenderHTML(w)
		}
	})
	http.HandleFunc("/organizations/", func(w http.ResponseWriter, req *http.Request) {
		updateIndex()
		segments := strings.Split(
			strings.TrimPrefix(req.URL.Path, "/organizations/"),
			"/",
		)

		organizationUUID := segments[0]
		organization := &Organization{}
		response := &Response{}
		index.Update(func(tx IndexTransaction) error {
			if err := tx.Get(organizationUUID, organization); err != nil {
				response.Error = err.Error()
			} else {
				response.Subject = organization
			}
			return nil
		})

		response.ServeHTTP(w, req)
	})

	http.HandleFunc("/project-cards/", func(w http.ResponseWriter, req *http.Request) {
		updateIndex()
		projectUuid := strings.TrimPrefix(req.URL.Path, "/project-cards/")
		card := &ProjectCard{}
		response := Response{}

		index.Update(func(tx IndexTransaction) error {
			if err := tx.Get("project-card:"+projectUuid, card); err != nil {
				response.Error = err.Error()
			} else {
				response.Subject = card
			}
			return nil
		})

		response.ServeHTTP(w, req)
	})

	go updateIndex()
	log.Info().Msgf("Listening on %s", *listen)
	log.Fatal().Err(http.ListenAndServe(*listen, nil))
}
