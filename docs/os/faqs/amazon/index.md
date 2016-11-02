---
title: Amazon FAQs
layout: os-default

---

## Amazon Frequently Asked Questions
---

### How can I extend my disk size?

Assuming your EC2 instance with RancherOS with more disk space than what's being read, run the following command to extend the disk size. This allows RancherOS to see the disk size.

```
$ docker run --privileged --rm --it debian:jessie resize2fs /dev/xvda1
```

`xvda1` should be the right disk for your own setup. In the future, we will be trying to create a system service that would automatically do this on boot in AWS.
