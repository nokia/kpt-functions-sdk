// Copyright 2025 The kpt and Nephio Authors
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
	"slices"

	kptfileapi "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

type SubObjectMatcher func(obj *fn.SubObject) bool

// IsType returns with a predicate that is true if the "type" field
// of the given object is equal to the given expectedType
func IsType(expectedType string) SubObjectMatcher {
	return func(obj *fn.SubObject) bool {
		typeStr, _, _ := obj.NestedString("type")
		return typeStr == expectedType
	}
}

// IsTypeAndStatus returns with a predicate that is true if the "type" and "status" fields
// ob the status condition SubObject matches with the given expected values
func IsTypeAndStatus(expectedType string, expectedStatus kptfileapi.ConditionStatus) SubObjectMatcher {
	return func(obj *fn.SubObject) bool {
		typeStr, _, _ := obj.NestedString("type")
		statusStr, _, _ := obj.NestedString("status")
		return typeStr == expectedType && statusStr == string(expectedStatus)
	}
}

// IsConditionType returns with a predicate that if the "conditionType" field
// of the given object is equal to the given expectedType
func IsConditionType(expectedType string) SubObjectMatcher {
	return func(obj *fn.SubObject) bool {
		return obj.GetString("conditionType") == expectedType
	}
}

func (kf *Kptfile) Conditions() fn.SliceSubObjects {
	ret, _, _ := kf.NestedSlice("status", "conditions")
	return ret
}

func (kf *Kptfile) SetConditions(conditions fn.SliceSubObjects) error {
	return kf.Status().SetSlice(conditions, "conditions")
}

func (kf *Kptfile) GetCondition(conditionType string) *fn.SubObject {
	conditions := kf.Conditions()
	i := slices.IndexFunc(conditions, IsType(conditionType))
	if i < 0 {
		return nil
	}
	return conditions[i]
}

// IsStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
// Inspired by https://pkg.go.dev/k8s.io/apimachinery/pkg/api/meta#IsStatusConditionPresentAndEqual
func (kf *Kptfile) IsStatusConditionPresentAndEqual(conditionType string, status kptfileapi.ConditionStatus) bool {
	conditions := kf.Conditions()
	for _, cond := range conditions {
		typeStr, _, _ := cond.NestedString("type")
		if typeStr == conditionType {
			statusStr, _, _ := cond.NestedString("status")
			return statusStr == string(status)
		}
	}
	return false
}

// IsStatusConditionTrue returns true when the conditionType is present and set to kptfileapi.ConditionTrue
// Inspired by https://pkg.go.dev/k8s.io/apimachinery/pkg/api/meta#IsStatusConditionTrue
func (kf *Kptfile) IsStatusConditionTrue(conditionType string) bool {
	return kf.IsStatusConditionPresentAndEqual(conditionType, kptfileapi.ConditionTrue)
}

// IsStatusConditionFalse returns true when the conditionType is present and set to kptfileapi.ConditionFalse
// Inspired by https://pkg.go.dev/k8s.io/apimachinery/pkg/api/meta#IsStatusConditionFalse
func (kf *Kptfile) IsStatusConditionFalse(conditionType string) bool {
	return kf.IsStatusConditionPresentAndEqual(conditionType, kptfileapi.ConditionFalse)
}

// GetTypedCondition returns with the condition whose type is `conditionType` as its first return value,
// or nil if the condition is missing.
func (kf *Kptfile) GetTypedCondition(conditionType string) (*kptfileapi.Condition, error) {
	cObj := kf.GetCondition(conditionType)
	if cObj == nil {
		return nil, nil
	}

	var cond kptfileapi.Condition
	err := cObj.As(&cond)
	if err != nil {
		return nil, err
	}
	return &cond, nil
}

// SetTypedCondition creates or updates the given condition using the Type field as the primary key
func (kf *Kptfile) SetTypedCondition(condition kptfileapi.Condition) error {
	conditions := kf.Conditions()
	i := slices.IndexFunc(conditions, IsType(condition.Type))
	if i >= 0 {
		// if the condition already exists, update it
		// NOTE: use the SetNestedString methods as opposed to SetNestedStringMap
		// in order to keep the order of new fields deterministic
		conditions[i].SetNestedString(string(condition.Status), "status")
		// the "if" prevents a corner case where changing the Reason/Message
		// from empty string to empty string and accidentally adding a Reason/Message
		// field to the condition that wasn't there originally
		if condition.Reason != conditions[i].GetString("reason") {
			conditions[i].SetNestedString(condition.Reason, "reason")
		}
		if condition.Message != conditions[i].GetString("message") {
			conditions[i].SetNestedString(condition.Message, "message")
		}
	} else {
		// otherwise, add the condition
		ko, err := fn.NewFromTypedObject(condition)
		if err != nil {
			return fmt.Errorf("failed to set condition %q: %w", condition.Type, err)
		}
		conditions = append(conditions, &ko.SubObject)
	}
	return kf.SetConditions(conditions)
}

// ApplyDefaultCondition adds the given condition to the Kptfile if a condition
// with the same type doesn't exist yet.
func (kf *Kptfile) ApplyDefaultCondition(condition kptfileapi.Condition) error {
	conditions := kf.Conditions()
	// if condition exists, do nothing
	if slices.ContainsFunc(conditions, IsType(condition.Type)) {
		return nil
	}

	// otherwise, add the condition
	ko, err := fn.NewFromTypedObject(condition)
	if err != nil {
		return fmt.Errorf("failed to apply default condition %q: %w", condition.Type, err)
	}
	conditions = append(conditions, &ko.SubObject)
	return kf.SetConditions(conditions)
}

// DeleteByTpe deletes all conditions with the given type
func (kf *Kptfile) DeleteConditionByType(conditionType string) error {
	conditions := kf.Conditions()
	if conditions == nil {
		return nil
	}
	conditions = slices.DeleteFunc(conditions, IsType(conditionType))
	return kf.SetConditions(conditions)
}

func (kf *Kptfile) ReadinessGates() fn.SliceSubObjects {
	ret, _, _ := kf.NestedSlice("info", "readinessGates")
	return ret
}

func (kf *Kptfile) SetReadinessGates(gates fn.SliceSubObjects) error {
	return kf.UpsertMap("info").SetSlice(gates, "readinessGates")
}

// EnsureReadinessGates ensures that the given readiness gates are present in the Kptfile.
func (kf *Kptfile) EnsureReadinessGates(gates []kptfileapi.ReadinessGate) error {
	if len(gates) == 0 {
		return nil
	}
	gateObjs := kf.ReadinessGates()
	for _, gate := range gates {
		if !slices.ContainsFunc(gateObjs, IsConditionType(gate.ConditionType)) {
			// if readiness gate is not in list, add it
			ko, err := fn.NewFromTypedObject(gate)
			if err != nil {
				return fmt.Errorf("failed to add readiness gate %s: %w", gate.ConditionType, err)
			}
			gateObjs = append(gateObjs, &ko.SubObject)
		}
	}
	return kf.SetReadinessGates(gateObjs)
}
