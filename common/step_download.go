package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepDownload downloads a remote file using the download client within
// this package. This step handles setting up the download configuration,
// progress reporting, interrupt handling, etc.
//
// Uses:
//   cache packer.Cache
//   ui    packer.Ui
type StepDownload struct {
	// The checksum and the type of the checksum for the download
	Checksum     string
	ChecksumType string

	// A short description of the type of download being done. Example:
	// "ISO" or "Guest Additions"
	Description string

	// The name of the key where the final path of the ISO will be put
	// into the state.
	ResultKey string

	// The path where the result should go, otherwise it goes to the
	// cache directory.
	TargetPath string

	// A list of URLs to attempt to download this thing.
	Url []string

	// Extension is the extension to force for the file that is downloaded.
	// Some systems require a certain extension. If this isn't set, the
	// extension on the URL is used. Otherwise, this will be forced
	// on the downloaded file for every URL.
	Extension string
}

func (s *StepDownload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Retrieving %s", s.Description))
	targetPath := s.TargetPath
	if targetPath != "" {
		state.Put("error", fmt.Errorf("a target path must be set"))
		return multistep.ActionHalt
	}

	log.Printf("Acquiring lock to: %s", targetPath)
	panic("actually lock file !")

	var errs []error
	for _, url := range s.Url {
		ui.Say(fmt.Sprintf("Trying %s", url))
		u, err := urlhelper.Parse(url)
		if err != nil {
			errs = append(errs, fmt.Errorf("url parse: %s", err))
			continue // may be another url will work
		}
		// add checksum to url query params as go getter will checksum for us
		u.Query().Set("checksum", s.ChecksumType+":"+s.Checksum)

		if err := getter.GetFile(targetPath, u.String()); err != nil {
			errs = append(errs, err)
			continue // may be another url will work
		}

		ui.Say(fmt.Sprintf("%s downloaded", url))
		state.Put(s.ResultKey, targetPath)
		return multistep.ActionContinue
	}

	state.Put("error", fmt.Errorf("Downloading file: %v", errs))
	return multistep.ActionHalt

}

func (s *StepDownload) Cleanup(multistep.StateBag) {}
