---
title: Getting Started on Amazon
layout: default

---
# Running RancherOS on Amazon Web Services
RancherOS is available as an Amazon Web Services AMI, and can be easily run on EC2.  Let’s walk through how to import and create a RancherOS on EC2 machine:

1. First login to your AWS console, and on the EC2 dashboard, click on **“Launch a new instance”**:

    ![RancherOS on AWS 1]({{site.baseurl}}/img/Rancher_aws1.png)

2. Choose Community AMIs and search for RancherOS:

    ![RancherOS on AWS 2]({{site.baseurl}}/img/Rancher_aws2.png)

3. After configuring the network, size of the instance, and the security groups, review the information of the instance and choose a ssh key pair to be used with the EC2instance:

    ![RancherOS on AWS 3]({{site.baseurl}}/img/Rancher_aws3.png)

4. Download your new key pair, and then launch the instance, you should see the instance up and running:

    ![RancherOS on AWS 4]({{site.baseurl}}/img/Rancher_aws4.png)

If you prefer to use the AWS CLI the command below will launch a new instance using the RancherOS AMI: 

```sh
$ aws ec2 run-instances --image-id ami-eeaefc86 --count 1 \
--instance-type t1.micro --key-name MyKey --security-groups new-sg
```
where ami-eeaefc86 is the AMI of RancherOS in us-east-1 region. You can find the codes for AMIs in other regions on the RancherOS GitHub site. Now you can login to the RancherOS system:

```sh
$ ssh -i MyKey.pem rancher@<ip-of-ec2-instance>
[rancher@rancher ~]$
```

# Latest Release 

Region | Type | AMI |
-------|------|------
ap-northeast-1| PV | [ami-71cb3d71](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchAmi=ami-71cb3d71)
ap-southeast-1| PV | [ami-4a9eaf18](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchAmi=ami-4a9eaf18)
ap-southeast-2| PV | [ami-45ef9f7f](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchAmi=ami-45ef9f7f)
eu-west-1| PV | [ami-fd70ee8a](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchAmi=ami-fd70ee8a)
sa-east-1| PV | [ami-85f94298](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchAmi=ami-85f94298)
us-east-1| PV | [ami-5a321d32](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchAmi=ami-5a321d32)
us-west-1| PV | [ami-bfa849fb](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchAmi=ami-bfa849fb)
us-west-2| PV | [ami-a9bc9099](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchAmi=ami-a9bc9099)

<br>
SSH keys are added to the <b>rancher</b> user.
<br>