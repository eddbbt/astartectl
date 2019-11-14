// Copyright © 2019 Ispirata Srl
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

package cluster

import (
	"context"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v28/github"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func listAstartes() (map[string]*unstructured.UnstructuredList, error) {
	ret := make(map[string]*unstructured.UnstructuredList)
	for k, v := range astarteResourceClients {
		list, err := v.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		if len(list.Items) > 0 {
			ret[k] = list
		}
	}

	return ret, nil
}

func getAstarte(astarteCRD dynamic.NamespaceableResourceInterface, name string, namespace string) (*unstructured.Unstructured, error) {
	return astarteCRD.Namespace(namespace).Get(name, metav1.GetOptions{})
}

func getAstarteOperator() (*appsv1.Deployment, error) {
	return kubernetesClient.AppsV1().Deployments("kube-system").Get("astarte-operator", metav1.GetOptions{})
}

func getLastOperatorRelease() (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	tags, _, err := client.Repositories.ListTags(ctx, "astarte-platform", "astarte-kubernetes-operator", &github.ListOptions{})
	if err != nil {
		return "", err
	}

	collection := semver.Collection{}

	for _, tag := range tags {
		ver, err := semver.NewVersion(strings.Replace(tag.GetName(), "v", "", -1))
		if err != nil {
			continue
		}

		collection = append(collection, ver)
	}

	sort.Sort(collection)

	return collection[len(collection)-1].Original(), nil
}

func getOperatorContent(path string, tag string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	content, _, _, err := client.Repositories.GetContents(ctx, "astarte-platform", "astarte-kubernetes-operator",
		path, &github.RepositoryContentGetOptions{Ref: "v" + tag})

	if err != nil {
		return "", nil
	}

	return content.GetContent()
}
