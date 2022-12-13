output "cluster_id" {
  description = "ID of the created EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_ca_data" {
  description = "Certificate authority data of the created EKS cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "cluster_endpoint" {
  description = "Cluster endpoint of the created EKS cluster"
  value       = module.eks.cluster_endpoint
}

output "cluster_arn" {
  description = "ARN of the created EKS cluster"
  value       = module.eks.cluster_arn
}

output "region" {
  description = "Region in which the EKS cluster is created"
  value       = data.aws_region.current.name
}
