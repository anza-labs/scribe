# Scribe

[![GitHub License](https://img.shields.io/github/license/anza-labs/scribe)][license]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](code_of_conduct.md)
[![GitHub issues](https://img.shields.io/github/issues/anza-labs/scribe)](https://github.com/anza-labs/scribe/issues)
[![GitHub release](https://img.shields.io/github/release/anza-labs/scribe)](https://GitHub.com/anza-labs/scribe/releases/)
[![Go Reference](https://pkg.go.dev/badge/github.com/anza-labs/scribe)](https://pkg.go.dev/github.com/anza-labs/scribe)
[![Go Report Card](https://goreportcard.com/badge/github.com/anza-labs/scribe)](https://goreportcard.com/report/github.com/anza-labs/scribe)

Scribe is a tool that automates the propagation of annotations across Kubernetes resources based on the annotations in a Namespace. This simplifies the process of managing annotations for multiple resources, ensuring consistency and reducing manual intervention.

- [Scribe](#scribe)
  - [Annotations](#annotations)
  - [Prometheus PodMonitors and ServiceMonitors](#prometheus-podmonitors-and-servicemonitors)
    - [ServiceMonitors](#servicemonitors)
      - [1. Simplest Example](#1-simplest-example)
      - [2. Selecting a Specific Port](#2-selecting-a-specific-port)
      - [3. Custom Path for Metrics](#3-custom-path-for-metrics)
    - [PodMonitors](#podmonitors)
      - [1. Simplest Example](#1-simplest-example-1)
      - [2. Selecting a Specific Port](#2-selecting-a-specific-port-1)
      - [3. Custom Path for Metrics](#3-custom-path-for-metrics-1)
  - [Installation](#installation)
  - [Contributing](#contributing)
    - [End-to-End Tests](#end-to-end-tests)
  - [License](#license)

## Annotations

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

## Prometheus PodMonitors and ServiceMonitors

> [!NOTE]
> While using annotations like `prometheus.io/scrape`, `prometheus.io/port`, and `prometheus.io/path` is a convenient way to enable basic metrics scraping, they do not provide the full flexibility and functionality offered by directly managing `ServiceMonitor` or `PodMonitor` resources.
> 
> Annotations are limited to straightforward configurations, such as enabling scraping, selecting ports, and specifying a metrics path. However, more advanced use cases - like defining custom scrape intervals, relabeling metrics, or adding TLS configurations - require direct manipulation of `ServiceMonitor` or `PodMonitor` objects.

One extra feature of Scribe is managing simple monitors from Prometheus. By default, Scribe enables the creation of [`ServiceMonitors`](https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.ServiceMonitor) and [`PodMonitors`](https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.PodMonitor) to facilitate metrics scraping for Kubernetes resources. This feature can be disabled by adding the annotation `scribe.anza-labs.dev/monitors=disabled` to the resource.

Scribe leverages common Prometheus annotations (`prometheus.io/*`) to determine the configuration for `ServiceMonitors` and `PodMonitors`. These annotations allow users to define:

- Whether a resource should be scraped (`prometheus.io/scrape`).
- Specific ports to target (`prometheus.io/port`).
- Custom paths for metrics (`prometheus.io/path`).

This enables seamless integration with Prometheus while keeping resource configurations simple and intuitive.

> [!WARNING]
> To successfully create a `ServiceMonitor` or `PodMonitor`, the resource must include at least one label. This is necessary because the `selector.matchLabels` field in the generated monitor uses these labels to identify the corresponding `Service` or `Pod` that Prometheus should scrape.

### ServiceMonitors

`ServiceMonitors` are used to scrape metrics from Kubernetes `Service` resources. Scribe automatically generates a `ServiceMonitor` based on the metadata annotations and resource specifications.

#### 1. Simplest Example

The following `Service` configuration uses the `prometheus.io/scrape` annotation to enable scraping for all exposed ports.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: service-monitor-example
  labels:
    app: example
  annotations:
    prometheus.io/scrape: 'true'
spec:
  selector:
    app: example
  ports:
  - name: web
    port: 8080
    targetPort: 80
  - name: web-secure
    port: 8443
    targetPort: 443
```

Generated ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: service-monitor-example
  labels:
    app: example
spec:
  selector:
    matchLabels:
      app: example
  endpoints:
  - port: web
    targetPort: 80
    path: '/metrics'
  - port: web-secure
    targetPort: 443
    path: '/metrics'
```

#### 2. Selecting a Specific Port

If a specific port should be scraped, use the `prometheus.io/port` annotation. It can reference a port number or name.

Using a port number:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/port: '8080'
```

Using a port name:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/port: 'web-secure'
```

#### 3. Custom Path for Metrics

Define a custom metrics path using the `prometheus.io/path` annotation:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/path: '/custom/metrics'
```

### PodMonitors

`PodMonitors` are used to scrape metrics directly from `Pod` resources. Similar to `ServiceMonitors`, they rely on annotations and resource specifications for configuration.

#### 1. Simplest Example

A `Pod` exposing multiple ports can use the `prometheus.io/scrape` annotation to enable scraping for all defined ports:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-monitor-example
  labels:
    app: example
  annotations:
    prometheus.io/scrape: 'true'
spec:
  containers:
  - name: nginx
    image: nginx:latest
    ports:
    - name: web
      containerPort: 80
    - name: web-secure
      containerPort: 443
```

Generated PodMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: pod-monitor-example
  labels:
    app: example
spec:
  selector:
    matchLabels:
      app: example
  podMetricsEndpoints:
  - port: web
    path: '/metrics'
  - port: web-secure
    path: '/metrics'
```

#### 2. Selecting a Specific Port

To scrape metrics from a specific port, use the `prometheus.io/port` annotation. The port can be specified as a number or name.

Using a port number:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/port: '8080'
```

Using a port name:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/port: 'web-secure'
```

#### 3. Custom Path for Metrics

Define a custom metrics path using the `prometheus.io/path` annotation:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: 'true'
    prometheus.io/path: '/custom/metrics'
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

Here's a draft for the **End to End Tests** section of your README:

### End-to-End Tests

The following steps outline how to run the end-to-end tests for this project. These tests validate the functionality of the system in a fully deployed environment.

1. **Create the Test Cluster**
    Start a local Kubernetes cluster for testing purposes:

    ```sh
    make cluster
    ```

2. **Build and Push the Docker Image**
    Build the Docker image and push it to the local registry:

    ```sh
    make docker-build docker-push IMG=localhost:5005/manager:e2e
    ```

3. **Deploy the Application**
    Deploy the application using the test image:

    ```sh
    make deploy IMG=localhost:5005/manager:e2e
    ```

After completing these steps, the application will be deployed and ready for end-to-end testing in the local Kubernetes environment.

The E2E tests can be now run using the following command:

```sh
make test-e2e
```

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
