package domain

import "github.com/craigmj/gototp"

// TotpTokenWithClock implements a TotpToken which does not depend on
// the time package but uses the domain clock instead.
type TotpTokenWithClock struct {
	*gototp.TOTP
}

func NewTotpTokenWithClock(secretB32 string) (*TotpTokenWithClock, error) {
	otp, err := gototp.New(secretB32)
	if err != nil {
		return nil, err
	}

	return &TotpTokenWithClock{TOTP: otp}, nil
}

func (totp *TotpTokenWithClock) FromNow(periods int64) int32 {
	period := (Clock.Now().Unix() / int64(totp.Period)) + int64(periods)
	return totp.ForPeriod(period)
}
