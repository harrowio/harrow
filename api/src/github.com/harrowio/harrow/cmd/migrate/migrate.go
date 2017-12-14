//go:generate go-bindata -pkg migrate -o compiled-migrations.go  -prefix ../.. ../../db/migrations/...
package migrate

import (
	"flag"
	"os"

	"github.com/harrowio/harrow/config"
	"github.com/rs/zerolog"
	migrate "github.com/rubenv/sql-migrate"
)

const ProgramName = "migrate"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()
	downN := flag.Int("down", 0, "migrate n migrations down instead of all the way up")
	flag.Parse()

	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("error opening database handle: %s\n", err)
	}

	log.Info().Msgf("looking for migrations to apply (env=%s)", c.Environment())
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "db/migrations",
	}
	var n int
	if *downN > 0 {
		log.Info().Msgf("migrating down, max %d\n", *downN)
		n, err = migrate.ExecMax(db.DB, "postgres", migrations, migrate.Down, *downN)
	} else {
		log.Info().Msgf("migrating up")
		n, err = migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	}

	if err != nil {
		log.Info().Msgf("error while running migrations: %s", err)
		// if pqErr, ok := err.(pq.Error); ok {
		// 	log.Warn().Msg(pqErr.Message)
		// 	log.Fatal().Msg(pqErr.Table)
		// }
	}

	log.Info().Msgf("applied %d migrations.\n", n)
}
