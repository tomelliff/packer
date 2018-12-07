package packer

import (
	"time"

	"github.com/cheggaaa/pb"
)

func speedyProgressBar(bar *pb.ProgressBar) {
	bar.SetUnits(pb.U_BYTES)
	bar.SetRefreshRate(1 * time.Millisecond)
	bar.NotPrint = true
	bar.Format("[\x00=\x00>\x00-\x00]")
}
