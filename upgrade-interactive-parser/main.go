package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
)


// HELM_PATH_REPOSITORY_FILE=/Users/drago/.helm/repository/repositories.yaml

func displayWarning(err error) {
	fmt.Printf("> %s\n", err)
}

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	var err error

	flagInput := flag.String("input", "", "[required] where to pick releases list generated by helm")
	flagBind := flag.String("bind", "", "bind chart name to repository (ex.: -bind jetstack/cert-manager) and exit")
	flagOutput := flag.String("output", "", "place to store changes")
	flagUpgradeAll := flag.Bool("all", false, "upgrade all releases without prompt")

	flag.Parse()


	// cache all known charts and their versions
	repository := &Repository{}
	err = repository.Init()
	exitOnError(err)

	// store fixes and exit if according flag is set
	if *flagBind != "" {
		arr := strings.Split(*flagBind, "/")
		err := repository.Bind(arr[1], arr[0])
		exitOnError(err)
		os.Exit(0)
	}

	if *flagInput == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// fetch available releases
	releases, err := fetchReleases(*flagInput)
	exitOnError(err)

	// split releases
	releasesActual, releasesDeprecated, releasesUpgradable := splitReleases(releases, repository)

	renderDeprecatedTable(releasesDeprecated)
	renderActualTable(releasesActual)
	renderUpgradableTable(releasesUpgradable)

	
	if *flagOutput != "" {
		releasesStore := interactiveReleasesSelect(releasesUpgradable, *flagUpgradeAll)
		if err := storeDiff(releasesStore, *flagOutput); err != nil {
			panic(err)
		}
	}
	
}