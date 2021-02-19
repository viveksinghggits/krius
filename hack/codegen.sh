#!/bin/bash

execDir=~/go/src/k8s.io/code-generator
"${execDir}"/generate-groups.sh all github.com/viveksinghggits/krius/pkg/client github.com/viveksinghggits/krius/pkg/apis krius.dev:v1alpha1 --go-header-file "${execDir}"/hack/boilerplate.go.txt
