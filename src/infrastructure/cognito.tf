# user pool stores and manages user profile, app client is specific configs inside that user pool that allows connection and authentication
resource "aws_cognito_user_pool" "main" {
  name = "saasysquad-users"

  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_uppercase = true
    require_numbers   = true
    require_symbols   = true
  }

  account_recovery_setting {
    recovery_mechanism {
      name    = "verified_email"
      priorty = 1
    }
  }
}

resource "aws_cognito_user_pool_client" "app" {
  name            = "saasysquad-client"
  user_pool_id    = aws_cognito_user_pool.main.id
  generate_secret = false # client secret disabled, no need since public clients

  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH"
  ]

  prevent_user_existence_errors = "ENABLED"

  access_token_validity  = 60
  id_token_validity      = 60
  refresh_token_validity = 30

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }
}

output "user_pool_id" {
  value = aws_cognito_user_pool.main.id
}

output "client_id" {
  value = aws_cognito_user_pool.app.id
}
