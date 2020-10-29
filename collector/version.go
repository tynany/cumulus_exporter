package collector

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	versionSubsystem = "version"

	versionLabels = []string{"major", "minor", "patch", "release"}
	versionDesc   = map[string]*prometheus.Desc{
		"info":  colPromDesc(versionSubsystem, "info", "Cumulus version info.", versionLabels),
		"major": colPromDesc(versionSubsystem, "major", "Cumulus major release.", nil),
		"minor": colPromDesc(versionSubsystem, "minor", "Cumulus minor release.", nil),
		"patch": colPromDesc(versionSubsystem, "patch", "Cumulus patch release.", nil),
	}

	totalVersionErrors = 0.0
	versionCollected   = false

	// collectedVersionInfo contains the cached values that this collector exports.
	collectedVersionInfo = versionInfo{}
)

func init() {
	registerCollector(versionSubsystem, enabledByDefault, NewVersionCollector)
}

// VersionCollector collects version metrics, implemented as per the Collector interface.
type VersionCollector struct{}

// NewVersionCollector returns a new VersionCollector.
func NewVersionCollector() Collector {
	return &VersionCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *VersionCollector) Get(ch chan<- prometheus.Metric) (float64, error) {

	// A simple cache is implemented as version info is static, so there is no reason to constantly read /etc/lsb-release
	if !versionCollected {
		if err := collectLSBRelease(); err != nil {
			totalVersionErrors++
			return totalVersionErrors, err
		}
	}
	labels := []string{collectedVersionInfo.majorStr, collectedVersionInfo.minorStr, collectedVersionInfo.patchStr, collectedVersionInfo.release}
	noLabels := []string{}

	newGauge(ch, versionDesc["info"], 1, labels...)
	newGauge(ch, versionDesc["major"], collectedVersionInfo.major, noLabels...)
	newGauge(ch, versionDesc["minor"], collectedVersionInfo.minor, noLabels...)
	newGauge(ch, versionDesc["patch"], collectedVersionInfo.patch, noLabels...)

	return totalVersionErrors, nil
}

func collectLSBRelease() error {
	file, err := os.Open("/etc/lsb-release")
	if err != nil {
		return err
	}
	defer file.Close()

	// Expected lines of /etc/lsb-release is in the form of:
	//   key=val
	// or
	//   key="val"
	// Regex will match on two groups -- key and val.
	reLSB := regexp.MustCompile(`(.*)="?([^"]+)`)

	// The entries map contains the keys and values from /etc/lsb-release.
	entries := make(map[string]string)

	// Iterate through each line of /etc/lsb-release
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := reLSB.FindStringSubmatch(line)

		// Expecting to find exactly 2 regex matches (i.e. the key and value).
		if len(match) != 3 {
			return fmt.Errorf("unexpected entry in /etc/lsb-release: %s", line)
		}

		// Add the found key and value to the entries map.
		entries[match[1]] = match[2]
	}

	// Expected value of DISTRIB_RELEASE is symantec versioning form:
	//   x.y.z
	// Regex will match on three groups -- x, y and z.
	reRelease := regexp.MustCompile(`(.*)\.(.*)\.(.*)`)

	// found is used to determine if a DISTRIB_RELEASE key is found in /etc/lsb-release.
	found := false

	// Iterate though all key/vals in /etc/lsb-release.
	for lsbName, lsbVal := range entries {
		if strings.ToUpper(lsbName) == "DISTRIB_RELEASE" {
			found = true

			// Populate collectedVersionInfo
			collectedVersionInfo.release = lsbVal
			match := reRelease.FindStringSubmatch(lsbVal)
			if len(match) != 4 {
				return fmt.Errorf("unexpected semantic version from DISTRIB_RELEASE in /etc/lsb-release: %s", lsbVal)
			}
			collectedVersionInfo.majorStr = match[1]
			collectedVersionInfo.major, err = strconv.ParseFloat(match[1], 64)
			if err != nil {
				return fmt.Errorf("cannot convert major version %q to float64: %v", match[1], err)
			}
			collectedVersionInfo.minorStr = match[2]
			collectedVersionInfo.minor, err = strconv.ParseFloat(match[2], 64)
			if err != nil {
				return fmt.Errorf("cannot convert minor version %q to float64: %v", match[2], err)
			}
			collectedVersionInfo.patchStr = match[3]
			collectedVersionInfo.patch, err = strconv.ParseFloat(match[3], 64)
			if err != nil {
				return fmt.Errorf("cannot convert patch version %q to float64: %v", match[3], err)
			}
			versionCollected = true

		}
	}
	if !found {
		return fmt.Errorf("cannot find DISTRIB_RELEASE in /etc/lsb-release")
	}
	return nil
}

type versionInfo struct {
	release  string
	majorStr string
	minorStr string
	patchStr string
	major    float64
	minor    float64
	patch    float64
}
