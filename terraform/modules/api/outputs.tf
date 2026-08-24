output "function_name" {
  value = aws_lambda_function.api.function_name
}

output "alias_name" {
  value = aws_lambda_alias.current.name
}

output "api_endpoint" {
  description = "API の公開 URL"
  value       = aws_apigatewayv2_stage.default.invoke_url
}
