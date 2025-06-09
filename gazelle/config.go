package gazelle

import (
	"github.com/goniz/gazelle-foreign-cc/common"
)

// Re-export common types and functions for backward compatibility
type CMakeConfig = common.CMakeConfig
type CMakeTarget = common.CMakeTarget
type CMakeConfigureFile = common.CMakeConfigureFile

// Re-export common functions for backward compatibility
var NewCMakeConfig = common.NewCMakeConfig
var GetCMakeConfig = common.GetCMakeConfig

// Note: In modern Gazelle, configuration is handled through the Language interface
// methods rather than a separate Configurer registration.
