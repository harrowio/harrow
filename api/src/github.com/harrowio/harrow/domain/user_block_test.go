package domain

import (
	"testing"
	"time"
)

func Test_UserBlock_BlockForever_setsValidFromToNow_andValidToToNil(t *testing.T) {
	user := &User{Uuid: "938cea0f-5068-4541-8263-b5313685cb7b"}
	block, err := user.NewBlock("testing")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()

	block.BlockForever(now)

	if got, want := block.ValidFrom, now; !got.Equal(want) {
		t.Errorf("block.ValidFrom = %s; want %s", got, want)
	}

	if got, want := block.ValidTo, (*time.Time)(nil); got != want {
		t.Errorf("block.ValidTo = %v; want %v", got, want)
	}
}
