package usecase

import common "github.com/scanoss/papi/api/commonv2"

type Error struct {
	Code    int
	Status  common.StatusCode
	Message string
	Error   error
}
