---
title: Getting Started on Amazon
layout: default

---
## Running RancherOS on AWS
----

### Launching RancherOS through the AWS console

RancherOS is available as an Amazon Web Services AMI, and can be easily run on EC2.  Let’s walk through how to import and create a RancherOS on EC2 machine using the AWS console.


1. First login to your AWS console, and go to the EC2 dashboard, click on **Launch Instance**:

    ![RancherOS on AWS 1]({{site.baseurl}}/img/Rancher_aws1.png)

2. Select the **Community AMIs** on the sidebar and search for **RancherOS**. Pick the latest version and click **Select**.

    ![RancherOS on AWS 2]({{site.baseurl}}/img/Rancher_aws2.png)

3. Go through the steps of creating the instance type through the AWS console. Choose your instance type, configure instance, add storage, tag instance, configure security group, and review. After you click on **Launch**, either create a new key pair or choose an existing key pair to be used with the EC2 instance. If you have created a new key pair, download the key pair. If you have chosen an existing key pair, make sure you have the key pair accessible. Click on **Launch Instances**. 

    ![RancherOS on AWS 3]({{site.baseurl}}/img/Rancher_aws3.png)

4. Your instance will be launching and you can click on **View Instances** to see it's status.

    ![RancherOS on AWS 4]({{site.baseurl}}/img/Rancher_aws4.png)
    
    Your instance is now running!
    
    ![RancherOS on AWS 5]({{site.baseurl}}/img/Rancher_aws5.png)

### Launching RancherOS through the AWS Command Line Interface
If you prefer to use the [AWS Command Line Interface](http://aws.amazon.com/cli/), let's walk through that process:

1. If you haven't installed the AWS CLI, follow the instructions on the [AWS CLI page](http://aws.amazon.com/cli/) to install. If you've already installed and configured AWS, just skip step 2. 

2. After you have installed AWS CLI, you'll need to configure your AWS. 

    ```bash
    $ aws configure
    ```
    
    Input your Access Key ID, Secret Access Key and Region name. You do not need to put in a output format name and can just leave it blank. 
    
    Note: Access Key ID and Secret Access Key can be found in the **Security Credentials** section of AWS. If you don't have one, **Create New Access Key**. When created, make sure to save the Secret Access Key. Or you can follow the instructions on AWS on how to create an IAM User account. 

    ```bash
    $ aws configure
    AWS Access Key ID [None]: ABCD 
    AWS Secret Access Key [None]: ABCD 
    Default region name [None]: us-east-1
    Default output format [None]:
    $
    ```
 
3. Once you've configured your AWS, use this command to launch an EC2 instance with the RancherOS AMI. You will need to know your SSH key name and security group name for the _region_ that you are configured for. These can be found from the AWS console.

    Note: See **Latest AMI Releases** for AMI names. 

    ```bash
    $ aws ec2 run-instances --image-id ami-ID# --count 1 --instance-type t1.micro --key-name MySSHKeyName --security-groups sg-name
    ```


## Logging into RancherOS
----

From a command line, log into the EC2 Instance.

```bash
$ ssh -i /Directory/of/MySSHKeyName.pem rancher@<ip-of-ec2-instance>
[rancher@rancher ~]$
```

If you have issues logging into RancherOS, try using this command to help debug the issue.

```bash
$ ssh -v -i /Directory/of/MySSHKeyName.pem rancher@<ip-of-ec2-instance>
```

After you're logged into the system, go back to the [Getting Started Guide]({{site.baseurl}}/docs/getting-started/) to see some examples of what we can do.


## Latest AMI Releases 
----

Please check the [README](https://github.com/rancherio/os/blob/master/README.md) in our RancherOS repository for our latest AMIs.

<br>
<br>
<br>



