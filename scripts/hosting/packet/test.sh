#!/bin/bash

set -ex

# https://www.packet.net/help/kb/how-to-provision-a-host-with-docker-machine/

# needs both docker-machine and the docker-machine packet.net driver
# https://github.com/packethost/docker-machine-driver-packet/releases

if [ "${PACKET_API_KEY}" == "" ]; then
	echo "need to set the PACKET_API_KEY"
	exit
fi
if [ "${PACKET_PROJECT_ID}" == "" ]; then
	echo "need to set the PACKET_PROJECT_ID"
	exit
fi

# facilities
#   New York Metro (EWR1) </span>
#   Silicon Valley (SJC1) </span>
#   Amsterdam, NL (AMS1) </span>
#   Tokyo, JP (NRT1) </span>
FACILITY=sjc1

# plan - the server types
PLAN=baremetal_0

# randomizing the hostname makes debugging things harder atm
#HOSTHASH=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 6 | head -n 1)
HOST=sven-${FACILITY}-${PLAN/_/-}

if ! docker-machine inspect $HOST ; then
	docker-machine create -d packet \
		--packet-api-key=${PACKET_API_KEY} --packet-project-id=${PACKET_PROJECT_ID} \
		--packet-facility-code ${FACILITY} \
		--packet-plan ${PLAN} \
		--packet-os=ubuntu_16_04 \
		${HOST}
fi

SSH="docker-machine ssh $HOST"
SCP="docker-machine scp"

echo "- setup.."

#Spin up an Ubuntu 16.04 Packet instance. There are two different categories: Type-0 and the other types. We'll need to test one from each category.
#SSH into the instance and change the root password.

USER="root"
PASS=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

echo "echo '$USER:$PASS' | chpasswd"
echo "echo '$USER:$PASS' | chpasswd" > /tmp/pass
chmod 755 /tmp/pass
$SCP /tmp/pass $HOST:~/pass
$SSH ./pass

#$SSH echo "root:$HOST" | chpasswd

#Download the initrd and vmlinuz for the RC.
$SCP ./scripts/hosting/packet/packet.sh $HOST:~/

#$SCP ./dist/artifacts/initrd $HOST:~/
#$SCP ./dist/artifacts/vmlinuz-4.9-rancher2 $HOST:~/vmlinuz

$SSH wget -c https://github.com/rancher/os/releases/download/v0.7.1/vmlinuz
$SSH wget -c https://github.com/rancher/os/releases/download/v0.7.1/initrd

#Install the kexec-tools package.

$SSH sudo apt-get update 

#SSH into the SOS shell for the instance. There's a button labelled "Console" on the page for the instance. If you click on that it'll give you an SSH command to paste into your terminal.

FACILITY=$(docker-machine inspect ${HOST} | grep Facility | sed 's/.*Facility": "\(.*\)".*/\1/')
DEVICEID=$(docker-machine inspect ${HOST} | grep DeviceID | sed 's/.*DeviceID": "\(.*\)".*/\1/')
SSHKEYPATH=$(docker-machine inspect ${HOST} | grep SSHKeyPath | sed 's/.*SSHKeyPath": "\(.*\)".*/\1/')

SSHSOS="./scripts/hosting/packet/test.expect $SSHKEYPATH $DEVICEID@sos.$FACILITY.packet.net $USER $PASS"

echo "--------------------------------------------------------------------------"
$SSHSOS uname -a 

#$SSH DEBIAN_FRONTEND=noninteractive sudo apt-get -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" install -yqq kexec-tools
#USING the SOSSSH expect script to get past the "Should kexec-tools handle reboots? [yes/no]"
$SSHSOS sudo DEBIAN_FRONTEND=noninteractive apt-get install -yqq kexec-tools


$SSHSOS reboot

echo "- kexecing"

$SSHSOS sudo kexec -f -l vmlinuz --initrd=initrd --append "rancher.password=${PASS} tinkerbell=http://bdba494d.ngrok.io console=ttyS1,115200n8 rancher.network.interfaces.eth0.dhcp=true rancher.network.interfaces.eth2.dhcp=true"

#The server will restart and then you should be running RancherOS from memory.
$SSHSOS reboot

## need to change the user for the exepct script
USER="rancher"
SSHSOS="./scripts/hosting/packet/test.expect $SSHKEYPATH $DEVICEID@sos.$FACILITY.packet.net $USER $PASS"

echo "--------------------------------------------------------------------------"
$SSHSOS uname -a 

# need to retrieve the packet.sh, vmlinuz and initrd from the disk
# TODO: this makes sense on type-0 - dunno about raid
# TODO: don't use dev, use LABEL - if&when we switch to running this on RancherOS
$SSHSOS sudo mount /dev/sda3 /mnt
$SSHSOS cp /mnt/root/* .
exit

#Clear the disk(s).

$SSHSOS sudo dd if=/dev/zero of=/dev/sda count=4 bs=1024

#If you're not running a type-0, also run the following command: 
if [ "$PLAN" != "baremetal_0" ]; then
	$SSHSOS sudo dd if=/dev/zero of=/dev/sdb count=4 bs=1024
fi

#Both of these will hang after you run them. Just let them run for a second or two and then hit ctrl+c.
#Download and run the Packet install script.

$SSHSOS bash ./packet.sh

#Reboot and then RancherOS should be fully installed.
#$SSHSOS reboot

#$SSH uname -a 
