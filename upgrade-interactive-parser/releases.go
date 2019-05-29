package main

import (
	"github.com/Masterminds/semver"
	"io/ioutil"
	"encoding/json"
	"strings"
	"fmt"
)

type helmList struct {
	Releases []struct {
		Name string
		Chart string
	}
}

type internalRelease struct {
	Name string
	Chart string
	Repository string
	Version *semver.Version
	LatestVersion *semver.Version
}

func fetchReleases(path string) ([]internalRelease, error) {
	releases := make([]internalRelease, 0)


	file, err := ioutil.ReadFile(path)
	if err != nil {
		return releases, err
	}
	data := helmList{}
	if err := json.Unmarshal([]byte(file), &data); err != nil {
		return releases, err
	}

	for _, item := range data.Releases {
		// Chart = "installer-1.0.5"
		arr := strings.Split(item.Chart, "-")
		last := len(arr)-1
		textVersion := arr[last]
		version, err := semver.NewVersion(textVersion)
		if err != nil {
			displayWarning(fmt.Errorf("while parsing %s: %s", item.Chart, err))
			continue
		}

		release := internalRelease{
			Name: item.Name,
			Chart: strings.Join(arr[:last], "-"),
			Version: version,		
		}
		releases = append(releases, release)

	}

	return releases, nil
}