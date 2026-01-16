// Copyright 2024-2025 The kpt and Nephio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kptfile

import (
	"fmt"
	"strings"

	kptfileapi "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	statusFieldName     = "status"
	conditionsFieldName = "conditions"
)

var (
	BoolToConditionStatus = map[bool]kptfileapi.ConditionStatus{
		true:  kptfileapi.ConditionTrue,
		false: kptfileapi.ConditionFalse,
	}
)

// Kptfile provides an API to manipulate the Kptfile of a kpt package
type Kptfile struct {
	fn.KubeObject
}

// NewFromKubeObjectList creates a Kptfile object by finding it in the given KubeObjects list
func NewFromKubeObjectList(objs fn.KubeObjects) (*Kptfile, error) {
	ko := objs.GetRootKptfile()
	if ko == nil {
		return nil, fmt.Errorf("the Kptfile object is missing from the package")
	}
	return &Kptfile{KubeObject: *ko}, nil
}

// NewFromPackage creates a Kptfile object from the resource (YAML) files of a package
func NewFromPackage(resources map[string]string) (*Kptfile, error) {
	kptfileStr, found := resources[kptfileapi.KptFileName]
	if !found {
		return nil, fmt.Errorf("file %q is missing from the package", kptfileapi.KptFileName)
	}

	kos, err := fn.ReadKubeObjectsFromFile(kptfileapi.KptFileName, kptfileStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse file %q from package: %w", kptfileapi.KptFileName, err)
	}
	return NewFromKubeObjectList(kos)
}

func (kf *Kptfile) WriteToPackage(resources map[string]string) error {
	if kf == nil {
		return fmt.Errorf("attempt to write empty Kptfile to the package")
	}
	kptfileStr, err := fn.WriteKubeObjectsToString(fn.KubeObjects{&kf.KubeObject})
	if err != nil {
		return err
	}
	resources[kptfileapi.KptFileName] = kptfileStr
	return nil
}

func (kf *Kptfile) String() string {
	if kf == nil {
		return ""
	}
	kptfileStr, _ := fn.WriteKubeObjectsToString(fn.KubeObjects{&kf.KubeObject})
	return kptfileStr
}

// Status returns with the Status field of the Kptfile as a SubObject
// If the Status field doesn't exist, it is added.
func (kf *Kptfile) Status() *fn.SubObject {
	return kf.UpsertMap(statusFieldName)
}

// DecodeKptfile decodes a KptFile from a yaml string.
func DecodeKptfile(kf string) (*kptfileapi.KptFile, error) {
	kptfile := &kptfileapi.KptFile{}
	f := strings.NewReader(kf)
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err := d.Decode(&kptfile); err != nil {
		return &kptfileapi.KptFile{}, fmt.Errorf("invalid 'v1' Kptfile: %w", err)
	}
	return kptfile, nil
}
