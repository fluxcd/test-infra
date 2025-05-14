module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.name
  cluster_version = "1.29"

  # Maybe don't need any of these?
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = true

  cluster_addons = {
    coredns = {
      resolve_conflicts = "OVERWRITE"
    }
    kube-proxy = {}
    vpc-cni = {
      resolve_conflicts = "OVERWRITE"
    }
  }

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # Define the default node group configuration.
  eks_managed_node_group_defaults = {
    disk_size            = 50
    instance_types       = ["t2.medium"]
    launch_template_tags = module.tags.tags
  }

  eks_managed_node_groups = {
    # Create node groups using on-demand nodes and spot nodes.
    blue = {}
    green = {
      min_size     = 1
      max_size     = 2
      desired_size = 1

      instance_types = ["t2.medium"]
      capacity_type  = "SPOT"
    }
  }

  enable_cluster_creator_admin_permissions = true

  # Disable log aggregation for such ephemeral clusters.
  cluster_enabled_log_types   = []
  create_cloudwatch_log_group = false

  # Disable encryption unless it's needed for some test.
  cluster_encryption_config = {}
  create_kms_key            = false

  # Enable IAM Roles for Service Accounts (IRSA) for the cluster.
  enable_irsa = var.enable_irsa

  tags = module.tags.tags
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = var.name
  cidr = "10.0.0.0/16"

  azs             = data.aws_availability_zones.available.names
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${var.name}" = "shared"
    "kubernetes.io/role/elb"            = 1
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${var.name}" = "shared"
    "kubernetes.io/role/internal-elb"   = 1
  }

  tags = module.tags.tags
}
