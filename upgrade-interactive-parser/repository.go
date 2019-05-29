package main

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"errors"
	"github.com/Masterminds/semver"
	"sort"
	"sync"
	"os"
)

// describes a structure of helm repository file
type helmRepositories struct {
	APIVersion string	`yaml:"apiVersion"`
	Repositories []struct {
		Name string
		Cache string
	}
}

// describes a structure of helm package file
type helmPackages struct {
	APIVersion string	`yaml:"apiVersion"`
	Entries map[string][]struct {
		Deprecated bool
		AppVersion string
		Home string
		Version string
	}
}

type internalChart struct {
	// chart name
	name string
	// name of the repository selected chart belongs to
	repoName string
	// list of available chart versions
	versions []*semver.Version
}

type Repository struct {
	sync.Mutex
	bindingPath string
	deprecated map[string]bool
	cache map[string]internalChart
	bindings map[string]string
}

// tells if this chart is market as non-handled by the tool
func (r *Repository) isChartDeprecated(name string) bool {
	_, exists := r.deprecated[name]
	return exists
}

// return repository name binded to this chart if it is binded
func (r *Repository) isChartBinded(name string) string {
	repo, exists := r.bindings[name]
	if exists {
		return repo
	} else {
		return ""
	}
}

func (r *Repository) isChartExists(name string) bool {
	_, exists := r.cache[name]
	return exists
}

func (r *Repository) deprecateChart(name string) {
	r.Lock()
	r.deprecated[name] = true
	delete(r.cache, name)
	r.Unlock()
}

func (r *Repository) restoreBindings() error {
	r.bindings = make(map[string]string)
	buf, err := ioutil.ReadFile(r.bindingPath)
	if err != nil {

		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := yaml.Unmarshal(buf, r.bindings); err != nil {
		return err
	}

	return nil
}

func (r *Repository) storeBindings() error {
	buf, err := yaml.Marshal(r.bindings)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(r.bindingPath, buf, 0644); err != nil {
		return err
	}
	return nil
}

func (r *Repository) nameByChart(name string) string {
	chart, exists := r.cache[name]
	if !exists {
		return ""
	} else {
		return chart.repoName
	}
}


func (r *Repository) Bind(chart, helmRepo string) error {
	r.Lock()
	r.bindings[chart] = helmRepo
	r.Unlock()
	return r.storeBindings()
}

func (r *Repository) Latest(name string, constraint *semver.Constraints, currentVersion *semver.Version) (*semver.Version, error) {
	if !r.isChartExists(name) {
		return nil, errors.New("chart does not exist")
	}
	if r.isChartDeprecated(name) {
		return nil, errors.New("chart is deprecated")
	}
	for _, version := range r.cache[name].versions {
		if currentVersion.String() == version.String() {
			// no need to update
			return nil, nil
		}
		if constraint.Check(version) {
			return version, nil
		}
	}
	return currentVersion, nil
}

func (r *Repository) Init() error {

	helmPluginDir := os.Getenv("HELM_PLUGIN_DIR")
	if helmPluginDir != "" {
		r.bindingPath = fmt.Sprintf("%s/bindings.yml", helmPluginDir)
	} else {
		r.bindingPath = fmt.Sprintf("%s/helm-bindings.yml", os.Getenv("HOME"))
	}

	if err := r.restoreBindings(); err != nil {
		return err
	}

	r.cache = make(map[string]internalChart)
	r.deprecated = make(map[string]bool)

	// get pathes to repositories caches
	repoContent, repoErr := ioutil.ReadFile("/Users/drago/.helm/repository/repositories.yaml")
	if repoErr != nil {
		return repoErr
	}
	repositoryList := helmRepositories{}
	if err := yaml.Unmarshal([]byte(repoContent), &repositoryList); err != nil {
		return err
	}
	if repositoryList.APIVersion != "v1" {
		return errors.New("found repositories is not compatible with v1")
	}
	// extract known packages and create own cache with it
	for _, repository := range repositoryList.Repositories {
		content, err := ioutil.ReadFile(repository.Cache)
		if err != nil {
			return repoErr
		}
		packageList := helmPackages{}
		if err := yaml.Unmarshal([]byte(content), &packageList); err != nil {
			return err
		}
		if packageList.APIVersion != "v1" {
			return errors.New("found package list is not compatible with v1")
		}
		
		for name, chart := range packageList.Entries {
			
			// chart is marked as deprecated, it means that
			// few same charts are exist in different repository and we
			// do not know which one should be upgraded
			if r.isChartDeprecated(name) {
				continue
			}

			// chart is alredy added in the cache
			// now we need to understand: is this chart is deprecated (see above)
			// or it is marked as binded to one repo by the user
			if r.isChartExists(name) {
				repoName := r.isChartBinded(name)
				// chart is not binded by the user - we can deprecate it
				if repoName == "" {
					r.deprecateChart(name)
					continue
				}
				// chart is binded to another repo - we can skip this one
				if repoName != repository.Name {
					continue
				}
				// looks like this chart was assigned to the wrong repo - we can remove it from the cache so the correct one will be added
				r.Lock()
				delete(r.cache, name)
				r.Unlock()
			}

			// collect all available chart versions
			versions := make([]*semver.Version, 0)
			for _, chartVersioned := range chart {
				
				version, err := semver.NewVersion(chartVersioned.Version)
				if err != nil {
					displayWarning(fmt.Errorf("while parsing %s: %s", chartVersioned.Version, err))
					continue
				}
				versions = append(versions, version)
			}

			sort.Sort(sort.Reverse(semver.Collection(versions)))
			r.Lock()
			r.cache[name] = internalChart{
				name: name,
				repoName: repository.Name,
				versions: versions,
			}
			r.Unlock()
		}
	}

	return nil
}