package harrowLimits

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog"
)

const ProgramName = "harrow-limits"

func Main() {

	if len(os.Args) < 2 {
		usage()
	}

	var logger zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

	organizationUuid := flag.String("organization-uuid", "", "id of the organization for which to calculate limits")
	creatorUuid := flag.String("creator-uuid", "", "id of the user making a change")
	reason := flag.String("reason", "", "reason for change")
	extraProjects := flag.Int("extra-projects", 0, "number of extra projects to grant")
	extraUsers := flag.Int("extra-users", 0, "number of extra users to grant")
	action := os.Args[1]
	flag.CommandLine.Parse(os.Args[2:])

	if organizationUuid == nil || *organizationUuid == "" {
		usage()
	}

	conf := config.GetConfig()
	db, err := conf.DB()
	if err != nil {
		logger.Fatal().Err(err)
	}

	tx := db.MustBegin()
	defer tx.Rollback()

	projectStore := stores.NewDbProjectStore(tx)
	organizationStore := stores.NewDbOrganizationStore(tx)
	billingPlanStore := stores.NewDbBillingPlanStore(tx, stores.NewBraintreeProxy())
	keyValueStore := stores.NewRedisKeyValueStore(redis.NewTCPClient(conf.RedisConnOpts(0)))
	billingEventStore := stores.NewDbBillingEventStore(tx)
	userStore := stores.NewDbUserStore(tx, conf)
	billingHistoryStore := stores.NewDbBillingHistoryStore(tx, keyValueStore)
	billingHistory, err := billingHistoryStore.Load()
	if err != nil {
		logger.Fatal().Msgf("billinghistorystore.load: %s\n", err)
	}

	organization, err := organizationStore.FindByUuid(*organizationUuid)
	if err != nil {
		logger.Fatal().Msgf("organization not found %s", err)
	}

	limitsStore := stores.NewDbLimitsStore(tx)
	limitsStore.SetLogger(logger)
	limitsService := limits.NewService(organizationStore, projectStore, billingPlanStore, billingHistory, limitsStore)

	planUuid := billingHistory.PlanUuidFor(*organizationUuid)
	plan, err := billingPlanStore.FindByUuid(planUuid)
	if err != nil {
		logger.Fatal().Msgf("billing plan store find by uuid %s", err)
	}

	lt := limitsTool{logger}

	switch action {
	case "show":
		lt.showLimits(limitsStore, organization, plan, limitsService, billingHistory)
	case "grant-extra-limits":
		err := grantExtraLimits(organization, *extraUsers, *extraProjects, *creatorUuid, *reason, billingEventStore, userStore)
		if err != nil {
			logger.Fatal().Msgf("grant extra limits error", err)
		}
		tx.Commit()
	default:
		logger.Fatal().Msgf("unknown command", action)
		usage()
	}
}

type limitsTool struct {
	log logger.Logger
}

func usage() {
	fmt.Printf("Usage:\n  %s show --organization-uuid=id\n", os.Args[0])
	os.Exit(1)
}

func (self *limitsTool) showLimits(limitsStore *stores.DbLimitsStore, organization *domain.Organization, plan *domain.BillingPlan, limitsService *limits.Service, history *domain.BillingHistory) {
	limits, err := limitsStore.FindByOrganizationUuid(organization.Uuid)
	if err != nil {
		self.log.Fatal().Msgf("limitsStore.FindByOrganizationUuid(%q): %s\n", organization.Uuid, err)
	}
	reported := limits.Report(plan, history.ExtraUsersFor(organization.Uuid), history.ExtraProjectsFor(organization.Uuid))

	exceeded, err := limitsService.Exceeded(organization)
	if err != nil {
		self.log.Fatal().Msgf("limitsService.Exceeded: %s\n", err)
	}

	output := struct {
		Limits       *domain.Limits       `json:"limits"`
		Organization *domain.Organization `json:"organization"`
		Exceeded     bool                 `json:"exceeded"`
	}{reported, organization, exceeded}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		self.log.Fatal().Msgf("json.MarshalIndent: %s\n", err)
	}

	fmt.Printf("%s\n", data)
}

func grantExtraLimits(organization *domain.Organization, extraUsers int, extraProjects int, creatorUuid string, reason string, billingEventStore *stores.DbBillingEventStore, userStore *stores.DbUserStore) error {
	_, err := userStore.FindByUuid(creatorUuid)
	if err != nil {
		return fmt.Errorf("creator: user %q not found: %s", creatorUuid, err)
	}

	reason = strings.TrimSpace(reason)
	if len(reason) == 0 {
		return fmt.Errorf("Provide a reason with the --reason flag")
	}

	if extraUsers < 0 {
		return fmt.Errorf("Cannot grant %d users (needs to be >= 0)", extraUsers)
	}

	if extraProjects < 0 {
		return fmt.Errorf("Cannot grant %d projects (needs to be >= 0)", extraProjects)
	}

	event := organization.NewBillingEvent(&domain.BillingExtraLimitsGranted{
		Users:     extraUsers,
		Projects:  extraProjects,
		GrantedBy: creatorUuid,
		Reason:    reason,
	})

	_, err = billingEventStore.Create(event)
	return err
}
