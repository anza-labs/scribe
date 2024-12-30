# Scribe

[![GitHub License](https://img.shields.io/github/license/anza-labs/scribe)][license]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](code_of_conduct.md)
[![GitHub issues](https://img.shields.io/github/issues/anza-labs/scribe)](https://github.com/anza-labs/scribe/issues)
[![GitHub release](https://img.shields.io/github/release/anza-labs/scribe)](https://GitHub.com/anza-labs/scribe/releases/)
[![Go Reference](https://pkg.go.dev/badge/github.com/anza-labs/scribe)](https://pkg.go.dev/github.com/anza-labs/scribe)
[![Go Report Card](https://goreportcard.com/badge/github.com/anza-labs/scribe)](https://goreportcard.com/report/github.com/anza-labs/scribe)

Scribe is a tool that automates the propagation of annotations across Kubernetes resources based on the annotations in a Namespace. This simplifies the process of managing annotations for multiple resources, ensuring consistency and reducing manual intervention.

## Description

Scribe is designed to streamline the management of Kubernetes annotations by observing namespaces and propagating specific annotations to all resources under its scope, such as Deployments, Pods, or other related objects. For example, you can use Scribe to automatically add annotations to enable features like auto-reload on updates, using tools like Reloader. 

The following example demonstrates how Scribe propagates the `reloader.stakater.com/auto=true` annotation from the namespace to all observed resources within that namespace:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: reloader-example
  annotations:
    # annotations below will be propagated to all resources
    # observed by the scribe controller - e.g. Deployments
    scribe.anza-labs.dev/annotations: |
      reloader.stakater.com/auto=true
```

This ensures all resources in the `reloader-example` namespace will inherit the specified annotation, allowing for seamless automation and consistency across deployments.

In addition to propagating static annotations, Scribe also supports annotation templating. This allows for more dynamic and flexible annotations, where values are injected based on the metadata of the observed resources.

For example, you could set a template annotation that injects the resource's name into the annotation, like so:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: reloader-example
  annotations:
    scribe.anza-labs.dev/annotations: |
      object.name={{ .metadata.name }}
```

## Installation

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/anza-labs)](https://artifacthub.io/packages/search?repo=anza-labs)

Installation of scribe is done via Helm Chart. To submit the issues or read more about installation process, please visit [the Charts repository](https://github.com/anza-labs/charts).

## Contributing

We welcome contributions to Scribe! Whether you're fixing a bug, improving the documentation, or adding new features, we'd love your help. 

To contribute:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature-branch-name`).
3. Commit your changes following the **Conventional Commits** specification (e.g., `feat: add new feature` or `fix: resolve issue with X`).
   - See [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for more details.
4. Push to the branch (`git push origin feature-branch-name`).
5. Open a Pull Request for review.

Make sure to follow the coding standards, and ensure that your code passes all tests and validations. We use `make` to help with common tasks. Run `make help` for more information on all available `make` targets.

## License

Copyright 2024 anza-labs contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[license]: https://github.com/registry-operator/registry-operator/blob/main/LICENSE
