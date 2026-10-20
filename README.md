# terraform-provider-thehive

> Unofficial side project, made for fun.

Terraform provider for TheHive 5. Very early stage, built as a learning/hobby project.

## Requirements

- TheHive 5.x (tested on 5.4+)
- Go 1.24+ (for building)
- Terraform 1.0+

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
- `thehive_case_status`
- `thehive_alert_status`
- `thehive_email_intake`
- `thehive_function`
- `thehive_taxonomy`
- `thehive_alert_feeder`
- `thehive_notification_endpoint`

## Data sources

- `thehive_profile`
- `thehive_organisation`
- `thehive_user`
- `thehive_custom_field`
- `thehive_observable_type`
- `thehive_case_template`

## Usage

```hcl
provider "thehive" {
  url     = "http://localhost:9000"
  api_key = var.api_key
}

data "thehive_profile" "analyst" {
  name = "analyst"
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
  profile      = data.thehive_profile.analyst.name
  organisation = "SOC"
}

resource "thehive_function" "auto_tag" {
  name  = "auto-tag-on-create"
  mode  = "Enabled"
  types = ["notification"]

  definition = <<-JS
    function handle(input, context) {
      if (input.object._type === "Case") {
        context.api.updateCase(input.object._id, {
          addTags: ["auto-tagged"]
        });
      }
    }
  JS
}

resource "thehive_cortex_config" "main" {
  server = [{
    name                   = "Cortex"
    url                    = "http://cortex:9001"
    auth_key               = var.cortex_api_key
    included_organisations = ["*"]
  }]
}

resource "thehive_email_intake" "inbox" {
  name          = "soc-inbox"
  organisations = ["SOC"]

  mailbox = {
    provider_name = "imap"

    credential = {
      email = "soc@example.com"

      basic_auth = {
        password = var.email_password
      }
    }
  }

  alert_properties = {
    type   = "email-intake"
    source = "{name}"
    tags   = ["email-intake"]
  }
}

resource "thehive_taxonomy" "tlp" {
  namespace   = "tlp"
  description = "Traffic Light Protocol"
  version     = 2
  exclusive   = true

  predicates = [{
    value    = "amber"
    expanded = "TLP:AMBER"
  }]

  values = [{
    predicate = "amber"
    entry = [{
      value    = "strict"
      expanded = "TLP:AMBER+STRICT"
    }]
  }]
}

resource "thehive_alert_feeder" "threat_feed" {
  name          = "threat-feed"
  description   = "Fetches indicators from an external feed."
  method        = "GET"
  url           = "https://feed.example.com/indicators"
  function_name = "process-threat-feed"
  enabled       = true

  interval = {
    value = 15
    unit  = "Minutes"
  }

  headers = [{
    key   = "Accept"
    value = "application/json"
  }]

  auth_type     = "basic"
  auth_username = var.feed_username
  auth_password = var.feed_password

  timeout_value = 30
  timeout_unit  = "Seconds"
}

resource "thehive_notification_endpoint" "alerts" {
  name = "incident-webhook"
  type = "webhook"
  url  = "https://notifications.example.com/thehive"

  auth_type     = "bearer"
  auth_password = var.notification_token

  included_organisations = ["SOC"]
  excluded_organisations = []
}
```

## Build

```bash
go mod tidy
go build -o $(go env GOPATH)/bin/terraform-provider-thehive
```

## Run

```bash
terraform init
terraform plan
terraform apply
```