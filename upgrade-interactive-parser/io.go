package main

import (
	"fmt"
	"os"
	"github.com/olekukonko/tablewriter"
	"github.com/Masterminds/semver"
	"gopkg.in/AlecAivazis/survey.v1"
)

func renderDeprecatedTable(releases []internalRelease) {
	if len(releases) == 0 {
		return
	}
	tableItems := make([][]string, 0)
	for _, release := range releases {
		row := []string{
			release.Name,
			release.Chart,
		}
		tableItems = append(tableItems, row)
	}
	fmt.Println("Releases won't be checked as they exist in more than one repo:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "CHART"})
	table.AppendBulk(tableItems) 
	table.Render()
}

func renderActualTable(releases []internalRelease) {
	if len(releases) == 0 {
		return
	}
	tableItems := make([][]string, 0)
	for _, release := range releases {
		row := []string{
			release.Name,
			fmt.Sprintf("%s/%s", release.Repository, release.Chart),
			release.Version.String(),
		}
		tableItems = append(tableItems, row)
	}
	fmt.Println("Releases no need to be upgraded:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "CHART", "VERSION"})
	table.AppendBulk(tableItems) 
	table.Render()
}

func renderUpgradableTable(releases []internalRelease) {
	if len(releases) == 0 {
		fmt.Println("Everything is upgraded, well done :-)")
		return
	}
	tableItems := make([][]string, 0)
	for _, release := range releases {
		row := []string{
			release.Name,
			fmt.Sprintf("%s/%s", release.Repository, release.Chart),
			fmt.Sprintf("%s => %s", release.Version.String(), release.LatestVersion.String()),
		}
		tableItems = append(tableItems, row)
	}
	fmt.Println("Releases to upgrade:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "CHART", "VERSION"})
	table.AppendBulk(tableItems) 
	table.Render()
}

// stores changes in format readable by shell script
func storeDiff(releases []internalRelease, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, release := range releases {
		// if len(filterNames) != 0 && 
		_, err := file.WriteString(fmt.Sprintf("%s %s/%s %s\n",
			release.Name,
			release.Repository,
			release.Chart,
			release.LatestVersion.String(),
		))
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Diff saved to %s\n", path)
	return nil
}

func mapReleasesByNames(releases []internalRelease, names []string) []internalRelease {
	result := []internalRelease{}
	for _, release := range releases {
		for _, name := range names {
			if name == release.Name {
				result = append(result, release)
			}
		}
	}
	return result
}


func interactiveReleasesSelect(releases []internalRelease, upgradeAll bool) ([]internalRelease) {
	if len(releases) == 0 || upgradeAll {
		return releases
	}
	
	selectResult:= []string{}
	selectOptions := []string{}
	for _, release := range releases {
		selectOptions = append(selectOptions, release.Name)
	}
	if upgradeAll == true {
		return mapReleasesByNames(releases, selectOptions)
	}
	prompt := &survey.MultiSelect{
		Message: "What should we upgrade:",
		Options: selectOptions,
		PageSize: 10,
	}
	survey.AskOne(prompt, &selectResult, nil)
	return mapReleasesByNames(releases, selectResult)
}

func splitReleases(releases []internalRelease, repository *Repository) ([]internalRelease, []internalRelease, []internalRelease) {

	releasesActual := make([]internalRelease, 0)
	releasesDeprecated := make([]internalRelease, 0)
	releasesUpgradable := make([]internalRelease, 0)
	

	constraint, err := semver.NewConstraint("x.x.x")
	if err != nil {
		panic(err)
	}
	for _, release := range releases {
		// unable to find repository for this release
		if repository.isChartDeprecated(release.Chart) {
			releasesDeprecated = append(releasesDeprecated, release)
			continue
		}
		latestVersion, err := repository.Latest(release.Chart, constraint, release.Version)
		if (err != nil) {
			displayWarning(fmt.Errorf("while comparing %s: %s", release.Chart, err))
			continue
		}
		release.Repository = repository.nameByChart(release.Chart)
		release.LatestVersion = latestVersion
		if latestVersion == nil {
			// no need to update
			releasesActual = append(releasesDeprecated, release)
			continue
		}
		releasesUpgradable = append(releasesUpgradable, release)

	}
	return releasesActual, releasesDeprecated, releasesUpgradable
}