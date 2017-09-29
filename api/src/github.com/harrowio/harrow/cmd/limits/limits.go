package harrowLimits

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	redis "gopkg.in/redis.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/protos"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

const ProgramName = "limits"

type LimitCmd struct {
	db     *sqlx.DB
	c      *config.Config
	logger logger.Logger
	lsCfg  config.LimitsStoreConfig

	bhStore *stores.DbBillingHistoryStore

	limitsStore limits.LimitsStore
}

func Main() {

	var (
		logger zerolog.Logger           = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()
		c      *config.Config           = config.GetConfig()
		lsCfg  config.LimitsStoreConfig = c.LimitsStoreConfig()
	)

	flag.CommandLine.Parse(os.Args[1:])

	lc := LimitCmd{
		c:      c,
		logger: logger,
		lsCfg:  lsCfg,
	}
	go lc.Serve()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	s := <-signals
	logger.Info().Msgf("got signal %s, exiting", s)
}

func (lc *LimitCmd) Log() logger.Logger {
	if lc.logger == nil {
		lc.logger = logger.Discard
	}
	return lc.logger
}

func (lc *LimitCmd) service() (*limits.Service, *sqlx.Tx) {
	tx, err := lc.tx()
	if err != nil {
		lc.Log().Fatal().Msgf("starting tx to init service %v", err)
	}
	kv := stores.NewRedisKeyValueStore(redis.NewTCPClient(lc.c.RedisConnOpts(0)))
	orgStore := stores.NewDbOrganizationStore(tx)
	projStore := stores.NewDbProjectStore(tx)
	bpStore := stores.NewDbBillingPlanStore(tx)

	bhStore := stores.NewDbBillingHistoryStore(tx, kv)
	bh, err := bhStore.Load()
	if err != nil {
		lc.Log().Fatal().Msgf("can't load billing history, failing %v", err)
	}

	limitsStore := stores.NewDbLimitsStore(tx)
	return limits.NewService(orgStore, projStore, bpStore, bh, limitsStore), tx
}

func (lc *LimitCmd) Exceeded(ctxt context.Context, oKey *protos.OrganizationKey) (*protos.OrganizationLimitsExceeded, error) {
	s, tx := lc.service()
	defer tx.Rollback()

	start := time.Now()
	orgStore := stores.NewDbOrganizationStore(tx)
	org, err := orgStore.FindByUuid(oKey.Uuid)
	if err != nil {
		lc.Log().Info().Msgf("error looking up organization in Exceeded(), %v", err)
		return nil, err
	}
	elapsed := time.Since(start)
	lc.Log().Debug().Str("uuid", org.Uuid).Int64("db-org-store-lookup-ms", elapsed.Nanoseconds()/int64(1e6)).Msg("finished")

	start = time.Now()
	e, err := s.Exceeded(org)
	if err != nil {
		lc.Log().Info().Msgf("error calling Exceeded(), %v", err)
		return nil, err
	}
	elapsed = time.Since(start)

	lc.Log().Debug().Str("uuid", org.Uuid).Int64("limits-service-exceeded-ms", elapsed.Nanoseconds()/int64(1e6)).Msg("finished")

	return &protos.OrganizationLimitsExceeded{Exceeded: e}, nil
}

func (lc *LimitCmd) ForOrganization(ctxt context.Context, oKey *protos.OrganizationKey) (*protos.OrganizationLimits, error) {

	tx, err := lc.tx()
	if err != nil {
		lc.Log().Fatal().Msgf("starting tx to init service %v", err)
	}
	defer tx.Rollback()

	kv := stores.NewRedisKeyValueStore(redis.NewTCPClient(lc.c.RedisConnOpts(0)))

	var (
		limitsStore         limits.LimitsStore            = stores.NewDbLimitsStore(tx)
		orgStore            *stores.DbOrganizationStore   = stores.NewDbOrganizationStore(tx)
		planStore           *stores.DbBillingPlanStore    = stores.NewDbBillingPlanStore(tx)
		billingHistoryStore *stores.DbBillingHistoryStore = stores.NewDbBillingHistoryStore(tx, kv)
	)

	start := time.Now()
	org, err := orgStore.FindByUuid(oKey.Uuid)
	if err != nil {
		lc.Log().Info().Msgf("error looking up organization in Exceeded(), %v", err)
		return nil, err
	}
	elapsed := time.Since(start)
	lc.Log().Debug().Str("uuid", org.Uuid).Int64("db-org-store-lookup-ms", elapsed.Nanoseconds()/int64(1e6)).Msg("finished")

	limits, err := limitsStore.FindByOrganizationUuid(org.Uuid)
	if err != nil {
		return nil, errors.Wrap(err, "can't lookup organization limits for uuid")
	}

	bh, err := billingHistoryStore.Load()
	if err != nil {
		lc.Log().Fatal().Msgf("can't load billing history, failing %v", err)
	}

	planUuid := bh.PlanUuidFor(org.Uuid)
	plan, err := planStore.FindByUuid(planUuid)
	if err != nil {
		lc.Log().Warn().Msgf("error loading plan for organization %s: %s", org.Uuid, err)
	}

	reported := limits.Report(
		plan,
		bh.ExtraUsersFor(org.Uuid),
		bh.ExtraProjectsFor(org.Uuid),
	)

	if err != nil {
		lc.Log().Info().Msgf("error calling Exceeded(), %v", err)
		return nil, err
	}
	lc.Log().Debug().Str("uuid", org.Uuid).Int64("limits-service-for-organization-ms", elapsed.Nanoseconds()/int64(1e6)).Msg("finished")

	return &protos.OrganizationLimits{
		Projects:            int32(reported.Projects),
		Members:             int32(reported.Members),
		PrivateRepositories: int32(reported.PrivateRepositories),
		PublicRepositories:  int32(reported.PublicRepositories),
		TrialDaysLeft:       int32(reported.TrialDaysLeft),
		TrialEnabled:        reported.TrialEnabled,
	}, nil
}

func (lc *LimitCmd) Serve() {

	lc.Log().Info().Msgf("starting limits server on %s:%s", lc.lsCfg.Bind, lc.lsCfg.Port)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", lc.lsCfg.Bind, lc.lsCfg.Port))
	if err != nil {
		lc.Log().Fatal().Msgf("can't start limits server on %s:%s quitting", lc.lsCfg.Bind, lc.lsCfg.Port)
	}

	s := grpc.NewServer()
	protos.RegisterLimitsServiceServer(s, lc)
	reflection.Register(s)
	if err := s.Serve(ln); err != nil {
		lc.Log().Fatal().Msgf("failed to start limits server: %v", err)
	}

}

func (lc *LimitCmd) tx() (*sqlx.Tx, error) {
	return lc.db_conn().Beginx()
}

func (lc *LimitCmd) db_conn() *sqlx.DB {
	if lc.db == nil {
		lc.Log().Debug().Msg("initiating db connection")
		db, err := lc.c.DB()
		if err != nil {
			lc.Log().Fatal().Msgf("error opening database handle %s", err)
		}
		lc.db = db
	}
	return lc.db
}

func usage() {
	fmt.Printf("Usage:\n  %s show --organization-uuid=id\n", os.Args[0])
	os.Exit(1)
}
