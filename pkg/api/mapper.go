/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package api

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/sets"
)

var RESTMapper meta.RESTMapper

func init() {
	RESTMapper = meta.MultiRESTMapper{}
}

func RegisterRESTMapper(m meta.RESTMapper) {
	RESTMapper = append(RESTMapper.(meta.MultiRESTMapper), m)
}

func NewDefaultRESTMapper(group string, groupVersionStrings []string, interfacesFunc meta.VersionInterfacesFunc,
	importPathPrefix string, ignoredKinds, rootScoped sets.String) *meta.DefaultRESTMapper {

	mapper := meta.NewDefaultRESTMapper(group, groupVersionStrings, interfacesFunc)
	// enumerate all supported versions, get the kinds, and register with the mapper how to address
	// our resources.
	for _, gvString := range groupVersionStrings {
		gv, err := unversioned.ParseGroupVersion(gvString)
		// TODO stop panicing when the types are fixed
		if err != nil {
			panic(err)
		}
		if gv.Group != group {
			panic(fmt.Sprintf("%q does not match the expect %q", gv.Group, group))
		}

		for kind, oType := range Scheme.KnownTypes(gv.String()) {
			// TODO: Remove import path prefix check.
			// We check the import path prefix because we currently stuff both "api" and "extensions" objects
			// into the same group within Scheme since Scheme has no notion of groups yet.
			if !strings.HasPrefix(oType.PkgPath(), importPathPrefix) || ignoredKinds.Has(kind) {
				continue
			}
			scope := meta.RESTScopeNamespace
			if rootScoped.Has(kind) {
				scope = meta.RESTScopeRoot
			}
			mapper.Add(scope, kind, gv.String(), false)
		}
	}
	return mapper
}
