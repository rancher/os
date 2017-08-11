---
title: RancherOS security
layout: os-default

---

## RancherOS security
---


<table width="100%">
<tr style="vertical-align: top;">
<td width="30%" style="border: none;">
<h4>Security policy</h4>
<p style="padding: 8px">Rancher Labs supports responsible disclosure, and endeavours to resolve all issues in a reasonable time frame. RancherOS is a minimal Linux distribution, built with entirely using open source components.</p>
</td>
<td width="30%" style="border: none;">
<h4>Reporting process</h4>
<p style="padding: 8px">Please submit possible security issues by emailing <a href="security@rancher.com">security@rancher.com</a></p>
</td>
<td width="30%" style="border: none;">
<h4>Announcments</h4>
<p style="padding: 8px">Subscribe to the <a href="https://forums.rancher.com/c/announcements">Rancher announcements forum</a> for release updates.</p>
</td>
</tr>
</table>

### RancherOS Vulnerabilities

| ID | Description | Date | Resolution |
|----|-------------|------|------------|
| [CVE-2017-6074](http://seclists.org/oss-sec/2017/q1/471) | Local privilege-escalation using a user after free issue in [Datagram Congestion Control Protocol (DCCP)](https://wiki.linuxfoundation.org/networking/dccp). DCCP is built into the RancherOS kernel as a dynamically loaded module, and isn't loaded by default. | 17 Feb 2017 | [RancherOS v0.8.1](https://github.com/rancher/os/releases/tag/v0.8.1) using a [patched 4.9.12 Linux kernel](https://github.com/rancher/os-kernel/releases/tag/v4.9.12-rancher) |
| [CVE-2017-7184](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-7184) | Allows local users to obtain root privileges or cause a denial of service (heap-based out-of-bounds access) by leveraging the CAP_NET_ADMIN capability. | 3 April 2017 | [RancherOS v0.9.2-rc1](https://github.com/rancher/os/releases/tag/v0.9.2-rc1) using Linux 4.9.20 |
| [CVE-2017-1000364](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-1000364) | Linux Kernel is prone to a local memory-corruption vulnerability. Attackers may be able to exploit this issue to execute arbitrary code with elevated privileges | 19 June 2017 | [RancherOS v1.0.3](https://github.com/rancher/os/releases/tag/v1.0.3) |
| [CVE-2017-1000366](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-1000366) | glibc contains a vulnerability that allows manipulation of the heap/stack. Attackers may be able to exploit this issue to execute arbitrary code with elevated privileges | 19 June 2017 | [RancherOS v1.0.3](https://github.com/rancher/os/releases/tag/v1.0.3) |

