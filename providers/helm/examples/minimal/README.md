# Minimal Helm Service Example

This is the simplest possible Helm service deployment.

## Structure

```
.
├── intent.yaml          # Sourceplane intent file
└── charts/
    └── hello-app/      # Simple Helm chart
        ├── Chart.yaml
        ├── values.yaml
        └── templates/
            ├── deployment.yaml
            └── service.yaml
```

## Intent Definition

See [intent.yaml](./intent.yaml) for the complete configuration.

## Usage

```bash
# Preview what will be deployed
sp plan

# Deploy the service
sp apply

# Check status
kubectl get pods -n default

# Verify
curl http://localhost:8080  # if using port-forward
```

## What This Example Shows

- Minimal component definition
- Local chart path reference
- Provider defaults in action
- Convention-based release naming
