/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generators

import (
	"io"
	"sort"
	"text/template"

	"fmt"

	"k8s.io/gengo/generator"
)

type apiGenerator struct {
	generator.DefaultGen
	apis *sortAPIs
}

type sortAPIs struct {
	*APIs
	ByVersion map[string][]string
}

var _ generator.Generator = &apiGenerator{}

func CreateApisGenerator(apis *APIs, filename string) generator.Generator {
	sortapis := sortAPIs{
		apis,
		sortGroupVersion(apis.Groups),
	}

	return &apiGenerator{
		generator.DefaultGen{OptionalName: filename},
		&sortapis,
	}
}

// sortGroupVersion always sort the apis reverse, so the newest apis will be preferred
func sortGroupVersion(groups map[string]*APIGroup) map[string][]string {
	byVersion := make(map[string][]string)
	for gName, g := range groups {
		if _, ok := byVersion[gName]; !ok {
			byVersion[gName] = []string{}
		}
		for version, _ := range g.Versions {
			byVersion[gName] = append(byVersion[gName], version)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(byVersion[gName])))
	}
	fmt.Println("Apis will be generate as the following priority", byVersion)
	return byVersion
}

func (d *apiGenerator) Imports(c *generator.Context) []string {
	imports := []string{
		"sigs.k8s.io/apiserver-builder-alpha/pkg/builders",
		"k8s.io/apimachinery/pkg/runtime",
	}
	for _, group := range d.apis.Groups {
		imports = append(imports, group.PkgPath)
		for _, version := range group.Versions {
			imports = append(imports, fmt.Sprintf(
				"%s%s \"%s\"", group.Group, version.Version, version.Pkg.Path))
		}
	}
	for _, group := range d.apis.Groups {
		imports = append(imports, fmt.Sprintf(
			"_ \"%s/install\" // Install the %s group", group.Pkg.Path, group.Group))
	}

	return imports
}

func (d *apiGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("apis-template").Parse(APIsTemplate))
	err := temp.Execute(w, d.apis)
	if err != nil {
		return err
	}
	return err
}

var APIsTemplate = `
var (
	localSchemeBuilder = runtime.SchemeBuilder{
{{ range $group := .Groups -}}
	{{ range $version := $group.Versions -}}
		{{ $group.Group }}{{ $version.Version }}.AddToScheme,
	{{ end -}}
{{ end -}}
	}
	AddToScheme = localSchemeBuilder.AddToScheme
)

// GetAllApiBuilders returns all known APIGroupBuilders
// so they can be registered with the apiserver
func GetAllApiBuilders() []*builders.APIGroupBuilder {
	return []*builders.APIGroupBuilder{
		{{ range $group := .Groups -}}
		Get{{ $group.GroupTitle }}APIBuilder(),
		{{ end -}}
	}
}

{{ range $group := .Groups -}}
var {{ $group.Group }}ApiGroup = builders.NewApiGroupBuilder(
	"{{ $group.Group }}.{{ $group.Domain }}",
	"{{ $group.PkgPath}}").
	WithUnVersionedApi({{ $group.Group }}.ApiVersion).
	WithVersionedApis(
		{{ range $version := (index $.ByVersion $group.Group) -}}
		{{ $group.Group }}{{ $version }}.ApiVersion,
		{{ end -}}
	).
	WithRootScopedKinds(
		{{ range $version := $group.Versions -}}
		{{ range $res := $version.Resources -}}
		{{ if $res.NonNamespaced -}}
		"{{ $res.Kind }}",
		{{ end -}}
		{{ end -}}
		{{ end -}}
	)

func Get{{ $group.GroupTitle }}APIBuilder() *builders.APIGroupBuilder {
	return {{ $group.Group }}ApiGroup
}
{{ end -}}
`
