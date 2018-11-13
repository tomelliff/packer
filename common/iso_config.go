package common

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"

	getter "github.com/hashicorp/go-getter"
)

// ISOConfig contains configuration for downloading ISO images.
type ISOConfig struct {
	ISOChecksum     string   `mapstructure:"iso_checksum"`
	ISOChecksumURL  string   `mapstructure:"iso_checksum_url"`
	ISOChecksumType string   `mapstructure:"iso_checksum_type"`
	ISOUrls         []string `mapstructure:"iso_urls"`
	TargetPath      string   `mapstructure:"iso_target_path"`
	TargetExtension string   `mapstructure:"iso_target_extension"`
	RawSingleISOUrl string   `mapstructure:"iso_url"`
}

func (c *ISOConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {

	if c.RawSingleISOUrl != "" {
		c.ISOUrls = append([]string{c.RawSingleISOUrl}, c.ISOUrls...)
	}
	if len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
		return
	}

	if c.ISOChecksumType == "" {
		errs = append(
			errs, errors.New("The iso_checksum_type must be specified."))
		return
	}
	c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)
	if c.ISOChecksumType != "none" {
		if c.ISOChecksum == "" && c.ISOChecksumURL == "" {
			errs = append(
				errs, errors.New("Due to large file sizes, an iso_checksum is required"))
			return warnings, errs
		}

		// If iso_checksum has no value use iso_checksum_url instead.
		if c.ISOChecksumURL != "" {
			dst := "/tmp/packer/chksm.iso" // TODO(azr): use code from #6950
			err := getter.Get(dst, c.ISOChecksumURL)
			if err != nil {
				return warnings, append(errs, fmt.Errorf("Failed to download iso checksum file: %s", err))
			}
			file, err := os.Open(dst)
			if err != nil {
				errs = append(errs, err)
				return warnings, errs
			}
			defer file.Close()
			err = c.parseCheckSumFile(bufio.NewReader(file))
			if err != nil {
				errs = append(errs, err)
				return warnings, errs
			}
		}
	}

	c.ISOChecksum = strings.ToLower(c.ISOChecksum)

	if c.TargetExtension == "" {
		c.TargetExtension = "iso"
	}
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

	return warnings, errs
}

func (c *ISOConfig) parseCheckSumFile(rd *bufio.Reader) error {
	u, err := url.Parse(c.ISOUrls[0])
	if err != nil {
		return err
	}

	checksumurl, err := url.Parse(c.ISOChecksumURL)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(u.Path)
	if err != nil {
		log.Printf("Unable to generate absolute path from provided iso_url: %s", err)
		absPath = ""
	}

	relpath, err := filepath.Rel(filepath.Dir(checksumurl.Path), absPath)
	if err != nil {
		log.Printf("Unable to determine relative pathing; continuing with abspath.")
		relpath = ""
	}

	filename := filepath.Base(u.Path)

	errNotFound := fmt.Errorf("No checksum for %q, %q or %q found at: %s",
		filename, relpath, u.Path, c.ISOChecksumURL)
	for {
		line, err := rd.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		options := []string{filename, relpath, "./" + relpath, absPath}
		if strings.ToLower(parts[0]) == c.ISOChecksumType {
			// BSD-style checksum
			for _, match := range options {
				if parts[1] == fmt.Sprintf("(%s)", match) {
					c.ISOChecksum = parts[3]
					return nil
				}
			}
		} else {
			// Standard checksum
			if parts[1][0] == '*' {
				// Binary mode
				parts[1] = parts[1][1:]
			}
			for _, match := range options {
				if parts[1] == match {
					c.ISOChecksum = parts[0]
					return nil
				}
			}
		}
	}
	return errNotFound
}
