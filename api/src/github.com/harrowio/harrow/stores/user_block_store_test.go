package stores_test

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_UserBlockStore_FindAllByUserUuid_returnsBlocksForUser(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbUserBlockStore(tx)
	tx.MustExec(`
          INSERT INTO user_blocks (uuid, user_uuid, reason, valid)
          SELECT
            uuid_generate_v4(),
            $1,
            'testing ' || i,
            '[1999-12-01,)'
          FROM
            generate_series(1, 2) as i
          ;`, user.Uuid)

	blocks, err := store.FindAllByUserUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 2; got != want {
		t.Errorf("len(blocks) = %d; want = %d", got, want)
	}
}

func Test_UserBlockStore_FindAllByUserUuid_doesNotReturnInvalidBlocks(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbUserBlockStore(tx)
	tx.MustExec(`
          INSERT INTO user_blocks (uuid, user_uuid, reason, valid)
          SELECT
            uuid_generate_v4(),
            $1,
            'testing ' || i,
            '(,1999-12-01]'
          FROM
            generate_series(1, 2) as i
          ;`, user.Uuid)

	blocks, err := store.FindAllByUserUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 0; got != want {
		t.Errorf("len(blocks) = %d; want = %d", got, want)
	}
}

func Test_UserBlockStore_FindAllByUserUuid_returnsNoErrorIfNoBlocksExist(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbUserBlockStore(tx)
	tx.MustExec(`
          INSERT INTO user_blocks (uuid, user_uuid, reason, valid)
          SELECT
            uuid_generate_v4(),
            $1,
            'testing ' || i,
            '[1999-11-01,1999-12-01]'
          FROM
            generate_series(1, 2) as i
          ;`, user.Uuid)

	blocks, err := store.FindAllByUserUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 0; got != want {
		t.Errorf("len(blocks) = %d; want = %d", got, want)
	}
}

func Test_UserBlockStore_Create_createsBlockThatCanBeFound(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	aDayAgo := time.Now().Add(-24 * time.Hour)
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	reasonForBlock := "testing"
	block, err := user.NewBlock(reasonForBlock)
	if err != nil {
		t.Fatal(err)
	}
	block.BlockForever(aDayAgo)

	store := stores.NewDbUserBlockStore(tx)
	if err := store.Create(block); err != nil {
		t.Fatal(err)
	}

	blocks, err := store.FindAllByUserUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 1; got != want {
		t.Fatalf("len(blocks) = %d; want = %d", got, want)
	}

	if got, want := blocks[0].Reason, reasonForBlock; got != want {
		t.Fatalf("blocks[0].Reason = %q; want %q", got, want)
	}
}
