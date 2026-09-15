# terraform-provider-thehive

> ⚠️ Unofficial side project, made for fun.

Terraform provider for TheHive 5. Very early stage, built as a learning/hobby project.

## Resources

- `thehive_organisation`
- `thehive_user`
- `thehive_profile`
- `thehive_custom_field`
- `thehive_observable_type`
- `thehive_case_template`
- `thehive_page_template`

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
