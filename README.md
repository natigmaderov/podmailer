# PodMailer Operator

## Overview
PodMailer is a Kubernetes operator that monitors pods in specified namespaces and sends email notifications when pods are down. It provides an easy way to set up email alerts for your Kubernetes cluster's pod health.

## Features
- Monitor pods in multiple namespaces
- Send email notifications when pods are down
- Configurable check intervals
- Support for SMTP email servers
- Easy to deploy and configure

## Quick Start

### Prerequisites
- Kubernetes cluster v1.11.3+
- kubectl configured to access your cluster
- SMTP server details for sending emails

### Installation

1. Install the operator:
```sh
kubectl apply -f https://raw.githubusercontent.com/natigmaderov/podmailer/main/dist/install.yaml
```

2. Create a PodMailer custom resource (replace the values with your configuration):
```yaml
apiVersion: podmailer.podmailer.io/v1alpha1
kind: PodMailer
metadata:
  name: podmailer-sample
spec:
  checkInterval: 60  # Check interval in seconds
  namespaces:
    - default
    - kube-system
  recipients:
    - user@example.com
  smtp:
    server: smtp.example.com
    port: 465
    username: your-username
    password: your-password
    fromEmail: alerts@example.com
```

Save this as `podmailer.yaml` and apply:
```sh
kubectl apply -f podmailer.yaml
```

## Configuration

### PodMailer Spec
| Field | Description | Type | Required |
|-------|-------------|------|----------|
| checkInterval | Interval in seconds between pod checks | int | Yes |
| namespaces | List of namespaces to monitor | []string | Yes |
| recipients | List of email addresses to receive notifications | []string | Yes |
| smtp.server | SMTP server address | string | Yes |
| smtp.port | SMTP server port (usually 465 or 587) | int | Yes |
| smtp.username | SMTP authentication username | string | Yes |
| smtp.password | SMTP authentication password | string | Yes |
| smtp.fromEmail | Email address to send notifications from | string | Yes |

## Uninstallation

To remove the operator and its resources:

```sh
kubectl delete -f https://raw.githubusercontent.com/natigmaderov/podmailer/main/dist/install.yaml
```

## Building from Source

If you want to build the operator from source:

1. Clone the repository:
```sh
git clone https://github.com/natigmaderov/podmailer.git
cd podmailer
```

2. Build and push the operator image:
```sh
make docker-build docker-push IMG=<your-registry>/podmailer:tag
```

3. Deploy to your cluster:
```sh
make deploy IMG=<your-registry>/podmailer:tag
```

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
Licensed under the Apache License, Version 2.0. See LICENSE file for details.

