# terraform-provider-thehive

> Unofficial side project, made for fun.

Terraform provider for TheHive 5. Very early stage, built as a learning/hobby project.

## Resources

- `thehive_organisation`
- `thehive_user`
- `thehive_profile`
- `thehive_custom_field`
- `thehive_observable_type`
- `thehive_case_template`
- `thehive_page_template`
- `thehive_cortex_config`
- `thehive_misp_config`

## Usage

```hcl
provider "thehive" {
  url     = "http://localhost:9000"
  api_key = var.api_key
}

resource "thehive_custom_field" "severity" {
  name         = "business_severity"
  display_name = "Business Severity"
  type         = "string"
  group        = "incident"
  options      = ["low", "medium", "high", "critical"]
}

resource "thehive_user" "alice" {
  login        = "alice@example.com"
  name         = "Alice"
  profile      = "analyst"
  organisation = "SOC"
}

resource "thehive_cortex_config" "main" {
  server = [{
    name                   = "Cortex"
    url                    = "http://cortex:9001"
    auth_key               = var.cortex_api_key
    included_organisations = ["*"]
  }]
}
```

## Build

```bash
go mod tidy
go build -o $(go env GOPATH)/bin/terraform-provider-thehive
```

## Usage

```bash
terraform init
terraform plan
terraform apply
```