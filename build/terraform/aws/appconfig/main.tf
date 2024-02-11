resource "aws_appconfig_application" "app" {
  name = var.config.name
}

resource "aws_appconfig_environment" "env" {
  for_each = {
    for env in var.config.environments : env.name => env
  }

  name           = each.value.name
  application_id = aws_appconfig_application.app.id
}

# AWS Lambda requires an instant deployment strategy.
resource "aws_appconfig_deployment_strategy" "instant" {
  name                           = "Instant"
  description                    = "This strategy deploys the configuration to all targets immediately with zero bake time."
  deployment_duration_in_minutes = 0
  final_bake_time_in_minutes     = 0
  growth_factor                  = 100
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
}
