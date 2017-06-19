package mailDispatcher

import (
	"encoding/base32"
	"fmt"
	"net/url"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/hmail"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"

	"github.com/jmoiron/sqlx"
)

type activityHandler func(logger.Logger, *config.Config, *domain.Activity, *sqlx.Tx) ([]*hmail.Mail, error)

var activityHandlers = map[string]activityHandler{
	"user.requested-password-reset":     handleUserRequestedPasswordReset,
	"user.signed-up":                    userVerificationEmail,
	"user.requested-verification-email": userVerificationEmail,
	"invitation.created":                handleInvitationCreated,
	"invitation.accepted":               handleInvitationChanged,
	"invitation.refused":                handleInvitationChanged,
}

type ActivityCreatedData struct {
	Table    string           `json:"table"`
	Activity *domain.Activity `json:"new"`
}

func handleUserRequestedPasswordReset(log logger.Logger, c *config.Config, activity *domain.Activity, tx *sqlx.Tx) ([]*hmail.Mail, error) {
	user, ok := activity.Payload.(*domain.User)
	if !ok {
		return nil, fmt.Errorf("Wrong payload type; have %T, want *domain.User", activity.Payload)
	}

	userStore := stores.NewDbUserStore(tx, c)
	user, err := userStore.FindByUuid(user.Uuid)
	if err != nil {
		return nil, err
	}
	hmac := user.HMAC([]byte(c.HttpConfig().UserHmacSecret))

	displayName := user.Name
	if displayName == "" {
		displayName = user.Email
	}
	recipient := &hmail.Recipient{
		Subject:     "Your password reset link",
		DisplayName: user.Name,
		UrlHost:     user.UrlHost,
	}

	actor := &hmail.Actor{
		DisplayName: "Someone",
	}

	action := &hmail.Action{
		DisplayName: "requested a password reset link",
	}

	escapedMac := url.QueryEscape(base32.StdEncoding.EncodeToString(hmac))
	escapedEmail := url.QueryEscape(user.Email)
	resetUri := fmt.Sprintf("#/a/reset-password?email=%s&mac=%s", escapedEmail, escapedMac)
	object := &hmail.Object{
		DisplayName: user.Email,
		Uri:         resetUri,
	}

	return []*hmail.Mail{{
		To:         []string{user.Email},
		RoutingKey: "users.requested-password-reset",
		Data: &hmail.MailContext{
			Recipient: recipient,
			Actor:     actor,
			Action:    action,
			Object:    object,
		},
	}}, nil
}

func userVerificationEmail(log logger.Logger, c *config.Config, activity *domain.Activity, tx *sqlx.Tx) ([]*hmail.Mail, error) {
	user, ok := activity.Payload.(*domain.User)
	if !ok {
		return nil, fmt.Errorf("Wrong payload type; have %T, want *domain.User", activity.Payload)
	}

	userStore := stores.NewDbUserStore(tx, c)
	user, err := userStore.FindByUuid(user.Uuid)
	if err != nil {
		return nil, err
	}

	displayName := user.Name
	if displayName == "" {
		displayName = user.Email
	}
	recipient := &hmail.Recipient{
		Subject:     "[Harrow.io] Welcome! Please verify your email address",
		DisplayName: user.Name,
		UrlHost:     user.UrlHost,
	}

	actor := &hmail.Actor{
		DisplayName: "You",
	}

	action := &hmail.Action{
		DisplayName: "signed up with",
	}

	params := url.Values{}
	params.Set("token", user.Token)
	params.Set("user", user.Uuid)

	activationUri := fmt.Sprintf("/a/verify-email?%s", params.Encode())
	object := &hmail.Object{
		DisplayName: user.Email,
		Uri:         activationUri,
	}

	return []*hmail.Mail{{
		To:         []string{user.Email},
		RoutingKey: "users.sign-up-email-verification",
		Data: &hmail.MailContext{
			Recipient: recipient,
			Actor:     actor,
			Action:    action,
			Object:    object,
		},
	}}, nil
}

func handleInvitationCreated(log logger.Logger, c *config.Config, activity *domain.Activity, tx *sqlx.Tx) ([]*hmail.Mail, error) {
	invitation, ok := activity.Payload.(*domain.Invitation)
	if !ok {
		return nil, fmt.Errorf("Wrong payload type; have %T, want *domain.Invitation", activity.Payload)
	}

	projectStore := stores.NewDbProjectStore(tx)
	invitationStore := stores.NewDbInvitationStore(tx)
	userStore := stores.NewDbUserStore(tx, c)

	invitation, err := invitationStore.FindByUuid(invitation.Uuid)
	if err != nil {
		log.Info().Msgf("handleInvitationCreated: invitationStore.FindByUuid(%q): %s\n", invitation.CreatorUuid, err)
		return nil, ErrStorage
	}

	inviter, err := userStore.FindByUuid(invitation.CreatorUuid)
	if err != nil {
		log.Info().Msgf("handleInvitationCreated: userStore.FindByUuid(%q): %s\n", invitation.CreatorUuid, err)
		return nil, ErrStorage
	}
	project, err := projectStore.FindByUuid(invitation.ProjectUuid)
	if err != nil {
		log.Info().Msgf("handleInvitationCreated: projectStore.FindByUuid(%q): %s\n", invitation.ProjectUuid, err)
		return nil, ErrStorage
	}

	// TODO(dh): if the invitee is not on Harrow yet, we just assume
	// that the invitee will have the same UrlHost as the inviter,
	// after sign-up.
	urlHost := inviter.UrlHost
	if invitee, err := userStore.FindByUuid(invitation.InviteeUuid); err == nil {
		urlHost = invitee.UrlHost
	} else {
		if _, ok := err.(*domain.NotFoundError); !ok {
			log.Info().Msgf("handleInvitationCreated: userStore.FindByUuid(%q): %s\n", invitation.InviteeUuid, err)
			return nil, ErrStorage
		}
	}

	recipient := &hmail.Recipient{
		Subject:     "Harrow: Project invitation",
		DisplayName: invitation.RecipientName,

		UrlHost: urlHost,
	}

	actor := &hmail.Actor{
		DisplayName: inviter.Name,
	}

	action := &hmail.Action{
		DisplayName: "sent you an invitation",
		Description: invitation.Message,
	}

	object := &hmail.Object{
		DisplayName: project.Name,
		Uri:         invitation.CallToActionPath(userStore),
	}

	return []*hmail.Mail{{
		To:         []string{invitation.Email},
		RoutingKey: "invitations.created",
		Data: &hmail.MailContext{
			Recipient: recipient,
			Actor:     actor,
			Action:    action,
			Object:    object,
		},
	}}, nil
}

func handleInvitationChanged(log logger.Logger, c *config.Config, activity *domain.Activity, tx *sqlx.Tx) ([]*hmail.Mail, error) {
	invitation, ok := activity.Payload.(*domain.Invitation)
	if !ok {
		return nil, fmt.Errorf("Wrong payload type; have %T, want *domain.Invitation", activity.Payload)
	}

	projectStore := stores.NewDbProjectStore(tx)
	userStore := stores.NewDbUserStore(tx, c)
	invitationStore := stores.NewDbInvitationStore(tx)

	invitation, err := invitationStore.FindByUuid(invitation.Uuid)
	if err != nil {
		log.Info().Msgf("handleInvitationChanged: invitationStore.FindByUuid(%q): %s\n", invitation.Uuid, err)
		return nil, ErrStorage
	}

	// We only want to send out mails when a user accepts or refuses
	// an invitation.
	if invitation.IsOpen() {
		return nil, nil
	}

	inviter, err := userStore.FindByUuid(invitation.CreatorUuid)
	if err != nil {
		log.Info().Msgf("handleInvitationChanged: userStore.FindByUuid(%q): %s\n", invitation.CreatorUuid, err)
		return nil, ErrStorage
	}
	project, err := projectStore.FindByUuid(invitation.ProjectUuid)
	if err != nil {
		log.Info().Msgf("handleInvitationChanged: projectStore.FindByUuid(%q): %s\n", invitation.ProjectUuid, err)
		return nil, ErrStorage
	}

	verb := ""
	if invitation.IsAccepted() {
		verb = "accepted"
	} else {
		verb = "refused"
	}

	recipient := &hmail.Recipient{
		Subject:     "Harrow: Project invitation " + verb,
		DisplayName: inviter.Name,

		UrlHost: inviter.UrlHost,
	}

	actor := &hmail.Actor{
		DisplayName: invitation.RecipientName,
	}

	action := &hmail.Action{
		DisplayName: verb,
	}

	object := &hmail.Object{
		DisplayName: fmt.Sprintf("your invitation to %q", project.Name),
		Uri:         fmt.Sprintf("/a/projects/%s", invitation.ProjectUuid),
	}

	return []*hmail.Mail{{
		To:         []string{inviter.Email},
		RoutingKey: "invitations.changed",
		Data: &hmail.MailContext{
			Recipient: recipient,
			Actor:     actor,
			Action:    action,
			Object:    object,
		},
	}}, nil
}
