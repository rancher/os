---
title: Getting Started on Amazon EC2
layout: os-default

---
## Running RancherOS on AWS
----
RancherOS is available as an Amazon Web Services AMI, and can be easily run on EC2. You can launch RancherOS either using the AWS Command Line Interface (CLI) or using the AWS console. 

### Launching RancherOS through the AWS CLI

If you haven't installed the AWS CLI, follow the instructions on the [AWS CLI page](http://aws.amazon.com/cli/) to install the CLI and configure access key and secret keys.

Once you've installed your AWS CLI, use this command to launch an EC2 instance with the RancherOS AMI. You will need to know your SSH key name and security group name for the _region_ that you are configured for. These can be found from the AWS console.

> **Note:** Check the RancherOS [README](https://github.com/rancher/os/blob/master/README.md) for AMI names for each region. We support PV and HVM types of AMIs. 

```
$ aws ec2 run-instances --image-id ami-ID# --count 1 --instance-type t2.micro --key-name MySSHKeyName --security-groups sg-name
```

Your EC2 instance is now running RancherOS!

### Launching RancherOS through the AWS Console

Let’s walk through how to import and create a RancherOS on EC2 machine using the AWS console.


1. First login to your AWS console, and go to the EC2 dashboard, click on **Launch Instance**:

    ![RancherOS on AWS 1]({{site.baseurl}}/img/os/Rancher_aws1.png)

2. Select the **Community AMIs** on the sidebar and search for **RancherOS**. Pick the latest version and click **Select**.

    ![RancherOS on AWS 2]({{site.baseurl}}/img/os/Rancher_aws2.png)

3. Go through the steps of creating the instance type through the AWS console. If you want to pass in a [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file during boot of RancherOS, you'd pass in the file as **User data** by expanding the **Advanced Details** in **Step 3: Configure Instance Details**. You can pass in the data as text or as a file. 
    
    ![RancherOS on AWS 6]({{site.baseurl}}/img/os/Rancher_aws6.png)

     After going through all the steps, you finally click on **Launch**, and either create a new key pair or choose an existing key pair to be used with the EC2 instance. If you have created a new key pair, download the key pair. If you have chosen an existing key pair, make sure you have the key pair accessible. Click on **Launch Instances**. 

    ![RancherOS on AWS 3]({{site.baseurl}}/img/os/Rancher_aws3.png)

4. Your instance will be launching and you can click on **View Instances** to see it's status.

    ![RancherOS on AWS 4]({{site.baseurl}}/img/os/Rancher_aws4.png)
    
    Your instance is now running!
    
    ![RancherOS on AWS 5]({{site.baseurl}}/img/os/Rancher_aws5.png)

## Logging into RancherOS
----

From a command line, log into the EC2 Instance. If you added ssh keys using a cloud-config,
both those keys, and the one you selected in the AWS UI will be installed.

```
$ ssh -i /Directory/of/MySSHKeyName.pem rancher@<ip-of-ec2-instance>
```

If you have issues logging into RancherOS, try using this command to help debug the issue.

```
$ ssh -v -i /Directory/of/MySSHKeyName.pem rancher@<ip-of-ec2-instance>
```

## Latest AMI Releases 
----

Please check the [README](https://github.com/rancher/os/blob/master/README.md) in our RancherOS repository for our latest AMIs.




