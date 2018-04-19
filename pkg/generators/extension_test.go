/*
Copyright 2018 The Kubernetes Authors.

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
	"reflect"
	"strings"
	"testing"
)

func TestSingleTagExtension(t *testing.T) {

	// Comments only contain one tag extension and one value.
	var tests = []struct {
		comments        []string
		extensionTag    string
		extensionName   string
		extensionValues []string
	}{
		{
			comments:        []string{"+patchMergeKey=name"},
			extensionTag:    "patchMergeKey",
			extensionName:   "x-kubernetes-patch-merge-key",
			extensionValues: []string{"name"},
		},
		{
			comments:        []string{"+patchStrategy=merge"},
			extensionTag:    "patchStrategy",
			extensionName:   "x-kubernetes-patch-strategy",
			extensionValues: []string{"merge"},
		},
		{
			comments:        []string{"+listType=atomic"},
			extensionTag:    "listType",
			extensionName:   "x-kubernetes-list-type",
			extensionValues: []string{"atomic"},
		},
		{
			comments:        []string{"+listMapKey=port"},
			extensionTag:    "listMapKey",
			extensionName:   "x-kubernetes-list-map-keys",
			extensionValues: []string{"port"},
		},
		{
			comments:        []string{"+k8s:openapi-gen=x-kubernetes-member-tag:member_test"},
			extensionTag:    "k8s:openapi-gen",
			extensionName:   "x-kubernetes-member-tag",
			extensionValues: []string{"member_test"},
		},
		{
			comments:        []string{"+k8s:openapi-gen=x-kubernetes-member-tag:member_test:member_test2"},
			extensionTag:    "k8s:openapi-gen",
			extensionName:   "x-kubernetes-member-tag",
			extensionValues: []string{"member_test:member_test2"},
		},
		{
			// Test that poorly formatted extensions aren't added.
			comments: []string{
				"+k8s:openapi-gen=x-kubernetes-no-value",
				"+k8s:openapi-gen=x-kubernetes-member-success:success",
				"+k8s:openapi-gen=x-kubernetes-wrong-separator;error",
			},
			extensionTag:    "k8s:openapi-gen",
			extensionName:   "x-kubernetes-member-success",
			extensionValues: []string{"success"},
		},
	}
	for _, test := range tests {
		extensions, _ := parseExtensions(test.comments)
		actual := extensions[0]
		if actual.tag != test.extensionTag {
			t.Errorf("Extension Tag: expected (%s), actual (%s)\n", test.extensionTag, actual.tag)
		}
		if actual.name != test.extensionName {
			t.Errorf("Extension Name: expected (%s), actual (%s)\n", test.extensionName, actual.name)
		}
		if !reflect.DeepEqual(actual.values, test.extensionValues) {
			t.Errorf("Extension Values: expected (%s), actual (%s)\n", test.extensionValues, actual.values)
		}
		if actual.hasMultipleValues() {
			t.Errorf("%s: hasMultipleValues() should be false\n", actual.name)
		}
	}

}

func TestMultipleTagExtensions(t *testing.T) {

	var tests = []struct {
		comments        []string
		extensionTag    string
		extensionName   string
		extensionValues []string
	}{
		{
			comments: []string{
				"+listMapKey=port",
				"+listMapKey=protocol",
			},
			extensionTag:    "listMapKey",
			extensionName:   "x-kubernetes-list-map-keys",
			extensionValues: []string{"port", "protocol"},
		},
	}
	for _, test := range tests {
		extensions, errors := parseExtensions(test.comments)
		if len(errors) > 0 {
			t.Errorf("Unexpected errors: %v\n", errors)
		}
		actual := extensions[0]
		if actual.tag != test.extensionTag {
			t.Errorf("Extension Tag: expected (%s), actual (%s)\n", test.extensionTag, actual.tag)
		}
		if actual.name != test.extensionName {
			t.Errorf("Extension Name: expected (%s), actual (%s)\n", test.extensionName, actual.name)
		}
		if !reflect.DeepEqual(actual.values, test.extensionValues) {
			t.Errorf("Extension Values: expected (%s), actual (%s)\n", test.extensionValues, actual.values)
		}
		if !actual.hasMultipleValues() {
			t.Errorf("%s: hasMultipleValues() should be true\n", actual.name)
		}
	}

}

func TestExtensionErrors(t *testing.T) {

	var tests = []struct {
		comments     []string
		errorMessage string
	}{
		{
			// Missing extension value should be an error.
			comments: []string{
				"+k8s:openapi-gen=x-kubernetes-no-value",
			},
			errorMessage: "x-kubernetes-no-value",
		},
		{
			// Wrong separator should be an error.
			comments: []string{
				"+k8s:openapi-gen=x-kubernetes-wrong-separator;error",
			},
			errorMessage: "x-kubernetes-wrong-separator;error",
		},
		{
			// disallowed is not one of the allowed values for listType.
			comments: []string{
				"+listType=disallowed",
			},
			errorMessage: "listType",
		},
		{
			// Missing list type value should be an error.
			comments: []string{
				"+listType",
			},
			errorMessage: "listType",
		},
		{
			// badStrategy is not one of the allowed values for patchStrategy.
			comments: []string{
				"+patchStrategy=badStrategy",
			},
			errorMessage: "patchStrategy",
		},
	}

	for _, test := range tests {
		_, errors := parseExtensions(test.comments)
		if len(errors) == 0 {
			t.Errorf("Expected errors while parsing: %v\n", test.comments)
		}
		error := errors[0]
		if !strings.Contains(error.Error(), test.errorMessage) {
			t.Errorf("Error (%v) should contain substring (%s)\n", error, test.errorMessage)
		}
	}
}

func TestExtensionAllowedValues(t *testing.T) {

	var successTests = []struct {
		e extension
	}{
		{
			e: extension{
				tag:    "patchStrategy",
				name:   "x-kubernetes-patch-strategy",
				values: []string{"merge"},
			},
		},
		{
			// Validate multiple values.
			e: extension{
				tag:    "patchStrategy",
				name:   "x-kubernetes-patch-strategy",
				values: []string{"merge", "retainKeys"},
			},
		},
		{
			e: extension{
				tag:    "patchMergeKey",
				name:   "x-kubernetes-patch-merge-key",
				values: []string{"key1"},
			},
		},
		{
			e: extension{
				tag:    "listType",
				name:   "x-kubernetes-list-type",
				values: []string{"atomic"},
			},
		},
	}
	for _, test := range successTests {
		actualErr := test.e.validateAllowedValues()
		if actualErr != nil {
			t.Errorf("Expected no error for (%v), but received: %v\n", test.e, actualErr)
		}
	}

	var failureTests = []struct {
		e extension
	}{
		{
			// Every value must be allowed.
			e: extension{
				tag:    "patchStrategy",
				name:   "x-kubernetes-patch-strategy",
				values: []string{"disallowed", "merge"},
			},
		},
		{
			e: extension{
				tag:    "patchStrategy",
				name:   "x-kubernetes-patch-strategy",
				values: []string{"foo"},
			},
		},
		{
			e: extension{
				tag:    "listType",
				name:   "x-kubernetes-list-type",
				values: []string{"not-allowed"},
			},
		},
	}
	for _, test := range failureTests {
		actualErr := test.e.validateAllowedValues()
		if actualErr == nil {
			t.Errorf("Expected error, but received none: %v\n", test.e)
		}
	}

}
