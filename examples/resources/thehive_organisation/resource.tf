# Create an organisation for incident response.
resource "thehive_organisation" "security_ops" {
  name        = "SecurityOps"
  description = "Primary SOC organisation for incident response."
}

# Import an existing organisation:
#   terraform import thehive_organisation.security_ops <org_id>
