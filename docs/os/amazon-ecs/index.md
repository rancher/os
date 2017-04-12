---
title: Amazon ECS on RancherOS
layout: os-default

---

## Amazon ECS (EC2 Container Service)
---

[Amazon ECS](https://aws.amazon.com/ecs/) is supported, which allows RancherOS EC2 instances to join your cluster.

### Pre-Requisites

Prior to launching RancherOS EC2 instances, the [ECS Container Instance IAM Role](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/instance_IAM_role.html) will need to have been created. This `ecsInstanceRole` will need to be used when launching EC2 instances. If you have been using ECS, you created this role if you followed the ECS "Get Started" interactive guide.

### Launching an instance with ECS

RancherOS makes it easy to join your ECS cluster. The ECS agent is a [system service]({{site.baseurl}}/os/system-services/adding-system-services/) that is enabled in the ECS enabled AMI. There may be other RancherOS AMIs that don't have the ECS agent enabled by default, but it can easily be added in the user data on any RancherOS AMI.

When launching the RancherOS AMI, you'll need to specify the **IAM Role** and **Advanced Details** -> **User Data** in the **Configure Instance Details** step.

For the **IAM Role**, you'll need to be sure to select the ECS Container Instance IAM role.

For the **User Data**, you'll need to pass in the [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file.

```yaml
#cloud-config
rancher:
  environment:
    ECS_CLUSTER: your-ecs-cluster-name
    # Note: You will need to add this variable, if using awslogs for ECS task.
    ECS_AVAILABLE_LOGGING_DRIVERS: |-
      ["json-file","awslogs"]
# If you have selected a RancherOS AMI that does not have ECS enabled by default,
# you'll need to enable the system service for the ECS agent.
  services_include:
    amazon-ecs-agent: true
```

#### Version

By default, the ECS agent will be using the `latest` tag for the `amazon-ecs-agent` image. In v0.5.0, we introduced the ability to select which version of the `amazon-ecs-agent`.

To select the version, you can update your [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file.

```yaml
#cloud-config
rancher:
  environment:
    ECS_CLUSTER: your-ecs-cluster-name
    # Note: You will need to make sure to include the colon in front of the version.
    ECS_AGENT_VERSION: :v1.9.0
    # If you have selected a RancherOS AMI that does not have ECS enabled by default,
    # you'll need to enable the system service for the ECS agent.
  services_include:
    amazon-ecs-agent: true
```

<br>

> **Note:** The `:` must be in front of the version tag in order for the ECS image to be tagged correctly.

### Amazon ECS enabled AMIs

Latest Release: [v1.0.0](https://github.com/rancher/os/releases/tag/v1.0.0)

Region | Type | AMI
---|--- | ---
ap-south-1 | HVM - ECS enabled | [ami-70ff8c1f](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-70ff8c1f)
eu-west-2 | HVM - ECS enabled | [ami-452f3b21](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-452f3b21)
eu-west-1 | HVM - ECS enabled | [ami-dbcbf1bd](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-dbcbf1bd)
ap-northeast-2 | HVM - ECS enabled | [ami-9dce1cf3](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-9dce1cf3)
ap-northeast-1 | HVM - ECS enabled | [ami-7ffcde18](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-7ffcde18)
sa-east-1 | HVM - ECS enabled | [ami-9c5e3cf0](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-9c5e3cf0)
ca-central-1 | HVM - ECS enabled | [ami-7fe05c1b](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-7fe05c1b)
ap-southeast-1 | HVM - ECS enabled | [ami-72259b11](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-72259b11)
ap-southeast-2 | HVM - ECS enabled | [ami-77a9a714](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-77a9a714)
eu-central-1 | HVM - ECS enabled | [ami-98d301f7](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-98d301f7)
us-east-1 | HVM - ECS enabled | [ami-6c4ec57a](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-6c4ec57a)
us-east-2 | HVM - ECS enabled | [ami-e7b99d82](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-e7b99d82)
us-west-1 | HVM - ECS enabled | [ami-653b1e05](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-653b1e05)
us-west-2 | HVM - ECS enabled | [ami-2eff6c4e](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-2eff6c4e)

