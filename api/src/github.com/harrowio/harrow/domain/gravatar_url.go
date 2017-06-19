package domain

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// gravatarUrl represents a valid url for the use with the Gravatar
// image service.
type gravatarUrl string

// newGravatarUrl returns a Gravatar url for displaying the picture
// associated with emailAddress.  The provided email address is normalized
// by converting it to lower case.
func newGravatarUrl(emailAddress string) gravatarUrl {
	normalizedEmailAddress := strings.ToLower(emailAddress)
	hash := md5.Sum([]byte(normalizedEmailAddress))
	encoded := hex.EncodeToString(hash[:])
	return gravatarUrl("https://secure.gravatar.com/avatar/" + encoded)
}

func (gravatar gravatarUrl) String() string {
	return string(gravatar)
}
