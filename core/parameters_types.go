package core

import "github.com/yumauri/fbrcm/core/config"

type ParametersCache = config.ParametersCache

type ParametersCacheState int

const (
	ParametersCacheMissing ParametersCacheState = iota
	ParametersCacheFresh
	ParametersCacheStale
)
