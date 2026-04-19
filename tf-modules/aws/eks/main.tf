module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  # Ensure the leaked-ENI sweeper is in the graph between VPC and cluster so
  # on destroy the order is: cluster -> sweep -> VPC.
  depends_on = [terraform_data.eni_sweep]

  cluster_name    = var.name
  cluster_version = "1.35"

  # Maybe don't need any of these?
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = true

  cluster_addons = {
    coredns = {
      resolve_conflicts = "OVERWRITE"
    }
    kube-proxy             = {}
    eks-pod-identity-agent = {}
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

  tags = module.tags.tags
}

# EKS provisions cross-account ENIs in the VPC for the managed control plane.
# On cluster deletion AWS reaps them asynchronously, and the VPC destroy often
# races ahead and fails with DependencyViolation on the cluster security group
# and subnets. This sweeps any lingering ENIs after the cluster is gone but
# before the VPC is destroyed.
resource "terraform_data" "eni_sweep" {
  depends_on = [module.vpc]

  input = {
    vpc_id = module.vpc.vpc_id
    region = data.aws_region.current.name
  }

  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      set -eu
      VPC_ID='${self.input.vpc_id}'
      REGION='${self.input.region}'
      echo ">>> ENI sweep: VPC=$VPC_ID region=$REGION"
      for i in $(seq 1 30); do
        # Skip NAT gateway ENIs — those are deleted by the VPC module itself.
        ENIS=$(aws ec2 describe-network-interfaces --region "$REGION" \
          --filters "Name=vpc-id,Values=$VPC_ID" \
          --query 'NetworkInterfaces[?InterfaceType!=`nat_gateway`].NetworkInterfaceId' \
          --output text)
        if [ -z "$ENIS" ]; then
          echo ">>> ENI sweep: none remaining (attempt $i)"
          exit 0
        fi
        echo ">>> ENI sweep attempt $i: $ENIS"
        for ENI in $ENIS; do
          ATTACH=$(aws ec2 describe-network-interfaces --region "$REGION" \
            --network-interface-ids "$ENI" \
            --query 'NetworkInterfaces[0].Attachment.AttachmentId' \
            --output text 2>/dev/null || echo None)
          if [ -n "$ATTACH" ] && [ "$ATTACH" != "None" ]; then
            echo ">>> detaching $ENI (attachment $ATTACH)"
            aws ec2 detach-network-interface --region "$REGION" \
              --attachment-id "$ATTACH" --force || true
          fi
          echo ">>> deleting $ENI"
          aws ec2 delete-network-interface --region "$REGION" \
            --network-interface-id "$ENI" || true
        done
        sleep 10
      done
      echo ">>> ENI sweep: timed out after 30 attempts; VPC destroy may still fail."
    EOT
  }
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
