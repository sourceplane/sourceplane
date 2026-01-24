# Terraform Provider for Sourceplane

The Terraform provider enables infrastructure as code components in Sourceplane using HashiCorp Terraform modules.

## Overview

This provider allows you to define infrastructure components using Terraform modules, supporting both local and remote module sources. It integrates with major cloud providers (AWS, GCP, Azure) and provides a unified way to manage infrastructure alongside your application components.

## Supported Component Types

### `terraform.database`
Managed database instances including:
- AWS RDS (PostgreSQL, MySQL, MariaDB)
- Google Cloud SQL
- Azure Database for PostgreSQL/MySQL
- DigitalOcean Managed Databases

### `terraform.cluster`
Container orchestration clusters including:
- Amazon EKS (Elastic Kubernetes Service)
- Google GKE (Google Kubernetes Engine)
- Azure AKS (Azure Kubernetes Service)
- Self-managed Kubernetes clusters

### `terraform.network`
Network infrastructure including:
- AWS VPC with subnets, route tables, NAT gateways
- GCP VPC networks
- Azure Virtual Networks
- Security groups and network ACLs

### `terraform.storage`
Object storage and file systems including:
- AWS S3 buckets
- Google Cloud Storage
- Azure Blob Storage
- DigitalOcean Spaces

### `terraform.compute`
Virtual machines and compute instances including:
- AWS EC2 instances
- Google Compute Engine
- Azure Virtual Machines
- DigitalOcean Droplets

## Configuration

### Basic Example

```yaml
components:
  - name: payments-db
    type: terraform.database
    spec:
      module:
        source: "terraform-aws-modules/rds/aws"
        version: "5.1.0"
      variables:
        identifier: "payments-database"
        engine: "postgres"
        engine_version: "14.5"
        instance_class: "db.t3.small"
        allocated_storage: 50
```

### Local Module

```yaml
components:
  - name: eks-cluster
    type: terraform.cluster
    spec:
      module:
        source: "./infra/eks"
      variables:
        cluster_name: "production"
        cluster_version: "1.27"
        node_instance_type: "t3.medium"
        desired_capacity: 3
```

### With Backend Configuration

```yaml
components:
  - name: vpc-network
    type: terraform.network
    spec:
      module:
        source: "terraform-aws-modules/vpc/aws"
        version: "5.0.0"
      variables:
        name: "production-vpc"
        cidr: "10.0.0.0/16"
        azs: ["us-east-1a", "us-east-1b", "us-east-1c"]
      backend:
        type: "s3"
        config:
          bucket: "my-terraform-state"
          key: "network/vpc.tfstate"
          region: "us-east-1"
      workspace: "production"
```

## Spec Reference

### `module` (required)

Terraform module configuration.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Module source (local path or registry) |
| `version` | string | No | Module version (registry modules only) |

**Module Source Examples:**
- Local: `./infra/database` or `../modules/database`
- Registry: `terraform-aws-modules/rds/aws`
- Git: `git::https://github.com/org/repo.git//modules/database`
- Git with ref: `git::https://github.com/org/repo.git//modules/database?ref=v1.0.0`

### `variables` (optional)

Input variables passed to the Terraform module. These correspond to the module's `variable` declarations.

```yaml
variables:
  instance_type: "db.t3.micro"
  engine: "postgres"
  engine_version: "14.5"
  multi_az: true
  backup_retention_period: 7
```

### `backend` (optional)

Terraform backend configuration for remote state storage.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Backend type (s3, gcs, azurerm, remote, local) |
| `config` | object | Yes | Backend-specific configuration |

**Backend Examples:**

**S3 Backend:**
```yaml
backend:
  type: "s3"
  config:
    bucket: "my-terraform-state"
    key: "database/terraform.tfstate"
    region: "us-east-1"
    encrypt: true
```

**GCS Backend:**
```yaml
backend:
  type: "gcs"
  config:
    bucket: "my-terraform-state"
    prefix: "database"
```

**Azure Backend:**
```yaml
backend:
  type: "azurerm"
  config:
    storage_account_name: "mystorageaccount"
    container_name: "terraform-state"
    key: "database.tfstate"
```

### `workspace` (optional)

Terraform workspace name. Defaults to `"default"`.

```yaml
workspace: "production"
```

### `provider_config` (optional)

Cloud provider-specific configuration.

**AWS:**
```yaml
provider_config:
  aws:
    region: "us-east-1"
    profile: "production"
```

**GCP:**
```yaml
provider_config:
  gcp:
    project: "my-project-123"
    region: "us-central1"
```

**Azure:**
```yaml
provider_config:
  azure:
    subscription_id: "12345678-1234-1234-1234-123456789012"
    location: "eastus"
```

### `outputs` (optional)

List of expected Terraform output names.

```yaml
outputs:
  - "endpoint"
  - "port"
  - "connection_string"
  - "security_group_id"
```

## Complete Examples

### Example 1: AWS RDS PostgreSQL Database

```yaml
components:
  - name: payments-db
    type: terraform.database
    spec:
      module:
        source: "terraform-aws-modules/rds/aws"
        version: "5.1.0"
      variables:
        identifier: "payments-database"
        engine: "postgres"
        engine_version: "14.5"
        instance_class: "db.t3.small"
        allocated_storage: 50
        storage_encrypted: true
        multi_az: true
        backup_retention_period: 7
        backup_window: "03:00-04:00"
        maintenance_window: "mon:04:00-mon:05:00"
        family: "postgres14"
        major_engine_version: "14"
      backend:
        type: "s3"
        config:
          bucket: "company-terraform-state"
          key: "payments/database.tfstate"
          region: "us-east-1"
          encrypt: true
      workspace: "production"
      outputs:
        - "db_instance_endpoint"
        - "db_instance_port"
        - "db_instance_name"
    relationships:
      - target: vpc-network
        type: depends_on
```

### Example 2: Amazon EKS Cluster

```yaml
components:
  - name: eks-cluster
    type: terraform.cluster
    spec:
      module:
        source: "terraform-aws-modules/eks/aws"
        version: "19.0.0"
      variables:
        cluster_name: "production-cluster"
        cluster_version: "1.27"
        vpc_id: "vpc-xxx"
        subnet_ids:
          - "subnet-xxx"
          - "subnet-yyy"
          - "subnet-zzz"
        cluster_endpoint_public_access: true
        eks_managed_node_groups:
          general:
            desired_size: 3
            min_size: 2
            max_size: 6
            instance_types: ["t3.medium"]
      backend:
        type: "s3"
        config:
          bucket: "company-terraform-state"
          key: "eks/cluster.tfstate"
          region: "us-east-1"
      workspace: "production"
    relationships:
      - target: vpc-network
        type: depends_on
```

### Example 3: Local Terraform Module

```yaml
components:
  - name: custom-infra
    type: terraform.compute
    spec:
      module:
        source: "./infra/custom-setup"
      variables:
        environment: "production"
        instance_count: 3
        enable_monitoring: true
      backend:
        type: "local"
        config:
          path: "terraform.tfstate"
```

### Example 4: Multi-Cloud Setup

```yaml
components:
  - name: aws-storage
    type: terraform.storage
    spec:
      module:
        source: "terraform-aws-modules/s3-bucket/aws"
        version: "3.10.0"
      variables:
        bucket: "my-app-assets"
        versioning_enabled: true
      provider_config:
        aws:
          region: "us-east-1"
          profile: "production"
  
  - name: gcp-storage
    type: terraform.storage
    spec:
      module:
        source: "terraform-google-modules/cloud-storage/google"
        version: "4.0.0"
      variables:
        name: "my-app-backups"
        location: "US"
      provider_config:
        gcp:
          project: "my-project-123"
          region: "us-central1"
```

## Best Practices

### 1. Use Registry Modules

Prefer well-maintained registry modules over custom modules:
```yaml
module:
  source: "terraform-aws-modules/rds/aws"
  version: "5.1.0"  # Pin to specific version
```

### 2. Pin Module Versions

Always specify module versions for reproducible infrastructure:
```yaml
module:
  source: "terraform-aws-modules/eks/aws"
  version: "19.0.0"  # Don't use version constraints like "~> 19.0"
```

### 3. Use Remote State

Configure remote backends for team collaboration:
```yaml
backend:
  type: "s3"
  config:
    bucket: "terraform-state"
    key: "component.tfstate"
    region: "us-east-1"
    encrypt: true
    dynamodb_table: "terraform-locks"
```

### 4. Organize with Workspaces

Use workspaces for environment separation:
```yaml
workspace: "production"  # or "staging", "dev"
```

### 5. Document Expected Outputs

List outputs that other components depend on:
```yaml
outputs:
  - "vpc_id"
  - "subnet_ids"
  - "security_group_id"
```

## Integration with Sourceplane

### Relationships

Terraform components can have relationships with other components:

```yaml
components:
  - name: vpc
    type: terraform.network
    # ... spec ...
  
  - name: database
    type: terraform.database
    relationships:
      - target: vpc
        type: depends_on
        description: "Database requires VPC to be created first"
  
  - name: api-service
    type: helm.service
    relationships:
      - target: database
        type: connects_to
        description: "API connects to database"
```

### Provider Configuration

Declare Terraform provider in your intent.yaml:

```yaml
providers:
  terraform:
    version: "0.1.0"
```

## Troubleshooting

### Module Not Found

**Error:** `Module not found at source path`

**Solution:** Verify the module source path is correct. For local modules, ensure the path is relative to the intent.yaml file.

### Invalid Variables

**Error:** `Variable X is not defined in the module`

**Solution:** Check the module's documentation for available input variables.

### Backend Configuration Issues

**Error:** `Failed to configure backend`

**Solution:** Ensure backend credentials are configured (AWS credentials, GCP service account, etc.).

### Version Compatibility

**Error:** `Module version X is not compatible`

**Solution:** Check the module's changelog for breaking changes and adjust your configuration.

## Additional Resources

- [Terraform Module Registry](https://registry.terraform.io/)
- [AWS Modules](https://registry.terraform.io/namespaces/terraform-aws-modules)
- [GCP Modules](https://registry.terraform.io/namespaces/terraform-google-modules)
- [Azure Modules](https://registry.terraform.io/namespaces/Azure)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)

## Provider Metadata

- **Version:** 0.1.0
- **API Version:** sourceplane.io/v1
- **Supported Platforms:** AWS, GCP, Azure, DigitalOcean, Kubernetes
- **Component Types:** 5 (database, cluster, network, storage, compute)
