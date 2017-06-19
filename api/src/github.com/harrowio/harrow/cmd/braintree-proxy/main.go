package braintreeProxy

import (
	"net/http"
	"os"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/services/braintree"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog"
)

const ProgramName = "braintree-proxy"

func Main() {

	var logger zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

	cfg := config.GetConfig().Braintree()
	client := cfg.NewClient()
	cache := braintree.NewFileCache(cfg.CacheFile)
	proxy := braintree.NewProxy(cache, stores.NewBraintreeAPI(client))
	logger.Fatal().Err(http.ListenAndServe("localhost:10002", proxy))
}
