package uuidhelper

import (
	gouuid "github.com/nu7hatch/gouuid"
)

func MustNewV4() string {
	if newUuid, err := gouuid.NewV4(); err != nil {
		panic(err)
	} else {
		return newUuid.String()
	}
}

func IsValid(uuid string) bool {
	_, err := gouuid.ParseHex(uuid)
	return err == nil
}
