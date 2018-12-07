package packer

import (
	"io"
	"sync"

	"github.com/cheggaaa/pb"
	getter "github.com/hashicorp/go-getter"
)

func defaultProgressbarConfigFn(bar *pb.ProgressBar) {
	bar.SetUnits(pb.U_BYTES)
}

var defaultUiProgressBar = uiProgressBar{}

// uiProgressBar is a self managed progress bar singleton
// decorate your struct with a *uiProgressBar to
// give it TrackProgress capabilities.
// In TrackProgress if uiProgressBar is nil
// defaultUiProgressBar will be used as
// the progress bar.
type uiProgressBar struct {
	l  sync.Mutex
	pb *getter.CheggaaaProgressBar
}

func (p *uiProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	if p == nil {
		p = &defaultUiProgressBar
	}
	p.l.Lock()
	defer p.l.Unlock()

	if p.pb == nil {
		p.pb = &getter.CheggaaaProgressBar{}
	}
	return p.pb.TrackProgress(src, currentSize, totalSize, stream)
}
