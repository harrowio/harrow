package stores_test

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupOAuthTokenStoreTest(t *testing.T) (*sqlx.Tx, *domain.User) {

	tx := helpers.GetDbTx(t)

	u := helpers.MustCreateUser(t, tx, &domain.User{
		Uuid:     "11111111-1111-4111-a111-111111111111",
		Name:     "Anne Mustermann",
		Email:    "anne@musterma.nn",
		Password: "changeme123",
	})

	return tx, u

}

func Test_OAuthTokenStore_SucessfullyCreating(t *testing.T) {

	tx, u := setupOAuthTokenStoreTest(t)
	defer tx.Rollback()

	token := &domain.OAuthToken{
		Uuid:        "44444444-3333-4333-a333-333333333333",
		UserUuid:    u.Uuid,
		Provider:    "github",
		Scope:       "comma,separated,list",
		AccessToken: "8c6f78ac23b4a7b8c0182d",
		TokenType:   "bearer",
	}

	savedToken := helpers.MustCreateOAuthToken(t, tx, token)

	if savedToken.Uuid != token.Uuid {
		t.Fatalf("Expected savedToken.Uuid != op.Uuid, got %v != %v", savedToken.Uuid, token.Uuid)
	}

}

func Test_OAuthTokenStore_SucessfullyFindingByUserUuidAndProvider(t *testing.T) {

	tx, u := setupOAuthTokenStoreTest(t)
	defer tx.Rollback()

	token := &domain.OAuthToken{
		Uuid:        "44444444-3333-4333-a333-333333333333",
		UserUuid:    u.Uuid,
		Provider:    "github",
		Scope:       "comma,separated,list",
		AccessToken: "8c6f78ac23b4a7b8c0182d",
		TokenType:   "bearer",
	}

	helpers.MustCreateOAuthToken(t, tx, token)

	userProviderToken, err := stores.NewDbOAuthTokenStore(tx).FindByProviderAndUserUuid("github", u.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if userProviderToken.Uuid != token.Uuid {
		t.Fatalf("Expected userProviderToken.Uuid != op.Uuid, got %v != %v", userProviderToken.Uuid, token.Uuid)
	}

}
