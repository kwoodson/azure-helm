// +build go1.9

// Copyright 2018 Microsoft Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code was auto-generated by:
// github.com/Azure/azure-sdk-for-go/tools/profileBuilder

package links

import (
	"context"

	original "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-09-01/links"
)

const (
	DefaultBaseURI = original.DefaultBaseURI
)

type Filter = original.Filter

const (
	AtScope Filter = original.AtScope
)

type BaseClient = original.BaseClient
type Operation = original.Operation
type OperationDisplay = original.OperationDisplay
type OperationListResult = original.OperationListResult
type OperationListResultIterator = original.OperationListResultIterator
type OperationListResultPage = original.OperationListResultPage
type OperationsClient = original.OperationsClient
type ResourceLink = original.ResourceLink
type ResourceLinkFilter = original.ResourceLinkFilter
type ResourceLinkProperties = original.ResourceLinkProperties
type ResourceLinkResult = original.ResourceLinkResult
type ResourceLinkResultIterator = original.ResourceLinkResultIterator
type ResourceLinkResultPage = original.ResourceLinkResultPage
type ResourceLinksClient = original.ResourceLinksClient

func New(subscriptionID string) BaseClient {
	return original.New(subscriptionID)
}
func NewOperationListResultIterator(page OperationListResultPage) OperationListResultIterator {
	return original.NewOperationListResultIterator(page)
}
func NewOperationListResultPage(getNextPage func(context.Context, OperationListResult) (OperationListResult, error)) OperationListResultPage {
	return original.NewOperationListResultPage(getNextPage)
}
func NewOperationsClient(subscriptionID string) OperationsClient {
	return original.NewOperationsClient(subscriptionID)
}
func NewOperationsClientWithBaseURI(baseURI string, subscriptionID string) OperationsClient {
	return original.NewOperationsClientWithBaseURI(baseURI, subscriptionID)
}
func NewResourceLinkResultIterator(page ResourceLinkResultPage) ResourceLinkResultIterator {
	return original.NewResourceLinkResultIterator(page)
}
func NewResourceLinkResultPage(getNextPage func(context.Context, ResourceLinkResult) (ResourceLinkResult, error)) ResourceLinkResultPage {
	return original.NewResourceLinkResultPage(getNextPage)
}
func NewResourceLinksClient(subscriptionID string) ResourceLinksClient {
	return original.NewResourceLinksClient(subscriptionID)
}
func NewResourceLinksClientWithBaseURI(baseURI string, subscriptionID string) ResourceLinksClient {
	return original.NewResourceLinksClientWithBaseURI(baseURI, subscriptionID)
}
func NewWithBaseURI(baseURI string, subscriptionID string) BaseClient {
	return original.NewWithBaseURI(baseURI, subscriptionID)
}
func PossibleFilterValues() []Filter {
	return original.PossibleFilterValues()
}
func UserAgent() string {
	return original.UserAgent() + " profiles/2017-03-09"
}
func Version() string {
	return original.Version()
}
