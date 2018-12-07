package common

import (
	"github.com/hashicorp/packer/helper/multistep"
)

var _ multistep.Step = new(StepDownload)
