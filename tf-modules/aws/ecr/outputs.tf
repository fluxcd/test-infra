output "repository_url" {
  description = "URL of the ECR instance"
  value       = aws_ecr_repository.this.repository_url
}

output "registry_id" {
  description = "ID of the ECR registry"
  value       = aws_ecr_repository.this.registry_id
}
