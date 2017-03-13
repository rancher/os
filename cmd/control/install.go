package control

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/rancher/os/log"

	"github.com/codegangsta/cli"
	"github.com/rancher/catalog-service/utils/version"
	"github.com/rancher/os/cmd/control/install"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs" // TODO: move CopyFile into util or something.
	"github.com/rancher/os/util"
)

var installCommand = cli.Command{
	Name:     "install",
	Usage:    "install RancherOS to disk",
	HideHelp: true,
	Action:   installAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			// TODO: need to validate ? -i rancher/os:v0.3.1 just sat there.
			Name: "image, i",
			Usage: `install from a certain image (e.g., 'rancher/os:v0.7.0')
							use 'ros os list' to see what versions are available.`,
		},
		cli.StringFlag{
			Name: "install-type, t",
			Usage: `generic:    (Default) Creates 1 ext4 partition and installs RancherOS (syslinux)
                        amazon-ebs: Installs RancherOS and sets up PV-GRUB
                        gptsyslinux: partition and format disk (gpt), then install RancherOS and setup Syslinux
                        `,
		},
		cli.StringFlag{
			Name:  "cloud-config, c",
			Usage: "cloud-config yml file - needed for SSH authorized keys",
		},
		cli.StringFlag{
			Name:  "device, d",
			Usage: "storage device",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "[ DANGEROUS! Data loss can happen ] partition/format without prompting",
		},
		cli.BoolFlag{
			Name:  "no-reboot",
			Usage: "do not reboot after install",
		},
		cli.StringFlag{
			Name:  "append, a",
			Usage: "append additional kernel parameters",
		},
		cli.StringFlag{
			Name:   "rollback, r",
			Usage:  "rollback version",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:   "isoinstallerloaded",
			Usage:  "INTERNAL use only: mount the iso to get kernel and initrd",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:  "kexec",
			Usage: "reboot using kexec",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Run installer with debug output",
		},
	},
}

func installAction(c *cli.Context) error {
	if runtime.GOARCH != "amd64" {
		log.Fatalf("ros install / upgrade only supported on 'amd64', not '%s'", runtime.GOARCH)
	}

	if c.Args().Present() {
		log.Fatalf("invalid arguments %v", c.Args())
	}

	if c.Bool("debug") {
		originalLevel := log.GetLevel()
		defer log.SetLevel(originalLevel)
		log.SetLevel(log.DebugLevel)
	}

	kappend := strings.TrimSpace(c.String("append"))
	force := c.Bool("force")
	kexec := c.Bool("kexec")
	reboot := !c.Bool("no-reboot")
	isoinstallerloaded := c.Bool("isoinstallerloaded")

	image := c.String("image")
	cfg := config.LoadConfig()
	if image == "" {
		image = cfg.Rancher.Upgrade.Image + ":" + config.Version + config.Suffix
	}

	installType := c.String("install-type")
	if installType == "" {
		log.Info("No install type specified...defaulting to generic")
		installType = "generic"
	}
	if installType == "rancher-upgrade" ||
		installType == "upgrade" {
		installType = "upgrade" // rancher-upgrade is redundant!
		force = true            // the os.go upgrade code already asks
		reboot = false
		isoinstallerloaded = true // OMG this flag is aweful - kill it with fire
	}
	device := c.String("device")
	if installType != "noformat" &&
		installType != "raid" &&
		installType != "bootstrap" &&
		installType != "upgrade" {
		// These can use RANCHER_BOOT or RANCHER_STATE labels..
		if device == "" {
			log.Fatal("Can not proceed without -d <dev> specified")
		}
	}

	cloudConfig := c.String("cloud-config")
	if cloudConfig == "" {
		if installType != "upgrade" {
			// TODO: I wonder if its plausible to merge a new cloud-config into an existing one on upgrade - so for now, i'm only turning off the warning
			log.Warn("Cloud-config not provided: you might need to provide cloud-config on bootDir with ssh_authorized_keys")
		}
	} else {
		uc := "/opt/user_config.yml"
		if err := util.FileCopy(cloudConfig, uc); err != nil {
			log.WithFields(log.Fields{"cloudConfig": cloudConfig}).Fatal("Failed to copy cloud-config")
		}
		cloudConfig = uc
	}

	if err := runInstall(image, installType, cloudConfig, device, kappend, force, kexec, isoinstallerloaded); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to run install")
	}

	if !kexec && reboot && (force || yes("Continue with reboot")) {
		log.Info("Rebooting")
		power.Reboot()
	}

	return nil
}

func runInstall(image, installType, cloudConfig, device, kappend string, force, kexec, isoinstallerloaded bool) error {
	fmt.Printf("Installing from %s\n", image)

	if !force {
		if util.IsRunningInTty() && !yes("Continue") {
			log.Infof("Not continuing with installation due to user not saying 'yes'")
			os.Exit(1)
		}
	}
	diskType := "msdos"

	if installType == "gptsyslinux" {
		diskType = "gpt"
	}

	// Versions before 0.8.0-rc3 use the old calling convention (from the lay-down-os shell script)
	imageVersion := strings.TrimPrefix(image, "rancher/os:")
	if version.GreaterThan("v0.8.0-rc3", imageVersion) {
		log.Infof("user specified to install pre v0.8.0: %s", image)
		imageVersion = strings.Replace(imageVersion, "-", ".", -1)
		vArray := strings.Split(imageVersion, ".")
		v, _ := strconv.ParseFloat(vArray[0]+"."+vArray[1], 32)
		if v < 0.8 || imageVersion == "0.8.0-rc1" {
			log.Infof("starting installer container for %s", image)
			if installType == "generic" ||
				installType == "syslinux" ||
				installType == "gptsyslinux" {
				cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=all-volumes",
					"--entrypoint=/scripts/set-disk-partitions", image, device, diskType)
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
			}
			cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=user-volumes",
				"--volumes-from=command-volumes", image, "-d", device, "-t", installType, "-c", cloudConfig,
				"-a", kappend)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}
	}

	if _, err := os.Stat("/usr/bin/system-docker"); os.IsNotExist(err) {
		if err := os.Symlink("/usr/bin/ros", "/usr/bin/system-docker"); err != nil {
			log.Errorf("ln error %s", err)
		}
	}

	useIso := false
	// --isoinstallerloaded is used if the ros has created the installer container from and image that was on the booted iso
	if !isoinstallerloaded {
		log.Infof("start !isoinstallerloaded")

		if _, err := os.Stat("/dist/initrd-" + config.Version); os.IsNotExist(err) {
			if err = mountBootIso(); err != nil {
				log.Debugf("mountBootIso error %s", err)
			} else {
				log.Infof("trying to load /bootiso/rancheros/installer.tar.gz")
				if _, err := os.Stat("/bootiso/rancheros/"); err == nil {
					cmd := exec.Command("system-docker", "load", "-i", "/bootiso/rancheros/installer.tar.gz")
					cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
					if err := cmd.Run(); err != nil {
						log.Infof("failed to load images from /bootiso/rancheros: %s", err)
					} else {
						log.Infof("Loaded images from /bootiso/rancheros/installer.tar.gz")

						//TODO: add if os-installer:latest exists - we might have loaded a full installer?
						useIso = true
						// now use the installer image
						cfg := config.LoadConfig()

						if image == cfg.Rancher.Upgrade.Image+":"+config.Version+config.Suffix {
							// TODO: fix the fullinstaller Dockerfile to use the ${VERSION}${SUFFIX}
							image = cfg.Rancher.Upgrade.Image + "-installer" + ":latest"
						}
					}
				}
				// TODO: also poke around looking for the /boot/vmlinuz and initrd...
			}

			log.Infof("starting installer container for %s (new)", image)
			installerCmd := []string{
				"run", "--rm", "--net=host", "--privileged",
				// bind mount host fs to access its ros, vmlinuz, initrd and /dev (udev isn't running in container)
				"-v", "/:/host",
				"--volumes-from=all-volumes",
				image,
				//				"install",
				"-t", installType,
				"-d", device,
				"-i", image, // TODO: this isn't used - I'm just using it to over-ride the defaulting
			}
			// Need to call the inner container with force - the outer one does the "are you sure"
			installerCmd = append(installerCmd, "-f")
			// The outer container does the reboot (if needed)
			installerCmd = append(installerCmd, "--no-reboot")
			if cloudConfig != "" {
				installerCmd = append(installerCmd, "-c", cloudConfig)
			}
			if kappend != "" {
				installerCmd = append(installerCmd, "-a", kappend)
			}
			if useIso {
				installerCmd = append(installerCmd, "--isoinstallerloaded=1")
			}
			if kexec {
				installerCmd = append(installerCmd, "--kexec")
			}

			// TODO: mount at /mnt for shared mount?
			if useIso {
				util.Unmount("/bootiso")
			}

			cmd := exec.Command("system-docker", installerCmd...)
			log.Debugf("Run(%v)", cmd)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}
	}

	// TODO: needs to pass the log level on to the container
	log.InitLogger()
	log.SetLevel(log.InfoLevel)

	log.Debugf("running installation")

	if installType == "generic" ||
		installType == "syslinux" ||
		installType == "gptsyslinux" {
		diskType := "msdos"
		if installType == "gptsyslinux" {
			diskType = "gpt"
		}
		log.Debugf("running setDiskpartitions")
		err := setDiskpartitions(device, diskType)
		if err != nil {
			log.Errorf("error setDiskpartitions %s", err)
			return err
		}
		// use the bind mounted host filesystem to get access to the /dev/vda1 device that udev on the host sets up (TODO: can we run a udevd inside the container? `mknod b 253 1 /dev/vda1` doesn't work)
		device = "/host" + device
	}

	if installType == "upgrade" {
		isoinstallerloaded = false
	}

	if isoinstallerloaded {
		log.Debugf("running isoinstallerloaded...")
		// TODO: detect if its not mounted and then optionally mount?
		if err := mountBootIso(); err != nil {
			log.Errorf("error mountBootIso %s", err)
			return err
		}
	}

	err := layDownOS(image, installType, cloudConfig, device, kappend, kexec)
	if err != nil {
		log.Errorf("error layDownOS %s", err)
		return err
	}

	return nil
}

func mountBootIso() error {
	deviceName := "/dev/sr0"
	deviceType := "iso9660"
	{ // force the defer
		mountsFile, err := os.Open("/proc/mounts")
		if err != nil {
			log.Errorf("failed to read /proc/mounts %s", err)
			return err
		}
		defer mountsFile.Close()

		if partitionMounted(deviceName, mountsFile) {
			return nil
		}
	}

	os.MkdirAll("/bootiso", 0755)

	// find the installation device
	cmd := exec.Command("blkid", "-L", "RancherOS")
	log.Debugf("Run(%v)", cmd)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("Failed to get RancherOS boot device: %s", err)
		return err
	}
	deviceName = strings.TrimSpace(string(out))
	log.Debugf("blkid found -L RancherOS: %s", deviceName)

	cmd = exec.Command("blkid", deviceName)
	log.Debugf("Run(%v)", cmd)
	cmd.Stderr = os.Stderr
	if out, err = cmd.Output(); err != nil {
		log.Errorf("Failed to get RancherOS boot device type: %s", err)
		return err
	}
	deviceType = strings.TrimSpace(string(out))
	s1 := strings.Split(deviceType, "TYPE=\"")
	s2 := strings.Split(s1[1], "\"")
	deviceType = s2[0]
	log.Debugf("blkid type of %s: %s", deviceName, deviceType)

	cmd = exec.Command("mount", "-t", deviceType, deviceName, "/bootiso")
	log.Debugf("Run(%v)", cmd)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf("tried and failed to mount %s: %s", deviceName, err)
	} else {
		log.Debugf("Mounted %s", deviceName)
	}
	return err
}

func layDownOS(image, installType, cloudConfig, device, kappend string, kexec bool) error {
	// ENV == installType
	//[[ "$ARCH" == "arm" && "$ENV" != "upgrade" ]] && ENV=arm

	// image == rancher/os:v0.7.0_arm
	// TODO: remove the _arm suffix (but watch out, its not always there..)
	VERSION := image[strings.Index(image, ":")+1:]

	var FILES []string
	DIST := "/dist" //${DIST:-/dist}
	//cloudConfig := SCRIPTS_DIR + "/conf/empty.yml" //${cloudConfig:-"${SCRIPTS_DIR}/conf/empty.yml"}
	CONSOLE := "tty0"
	baseName := "/mnt/new_img"
	bootDir := "boot/"
	//# TODO: Change this to a number so that users can specify.
	//# Will need to make it so that our builds and packer APIs remain consistent.
	partition := device + "1"                                                //${partition:=${device}1}
	kernelArgs := "rancher.state.dev=LABEL=RANCHER_STATE rancher.state.wait" // console="+CONSOLE

	// unmount on trap
	defer util.Unmount(baseName)

	diskType := "msdos"
	if installType == "gptsyslinux" {
		diskType = "gpt"
	}

	switch installType {
	case "syslinux":
		fallthrough
	case "gptsyslinux":
		fallthrough
	case "generic":
		log.Debugf("formatAndMount")
		var err error
		device, partition, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			log.Errorf("formatAndMount %s", err)
			return err
		}
		err = installSyslinux(device, baseName, bootDir, diskType)
		if err != nil {
			log.Errorf("installSyslinux %s", err)
			return err
		}
		err = seedData(baseName, cloudConfig, FILES)
		if err != nil {
			log.Errorf("seedData %s", err)
			return err
		}
	case "arm":
		var err error
		device, partition, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		seedData(baseName, cloudConfig, FILES)
	case "amazon-ebs-pv":
		fallthrough
	case "amazon-ebs-hvm":
		CONSOLE = "ttyS0"
		var err error
		device, partition, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		if installType == "amazon-ebs-hvm" {
			installSyslinux(device, baseName, bootDir, diskType)
		}
		//# AWS Networking recommends disabling.
		seedData(baseName, cloudConfig, FILES)
	case "googlecompute":
		CONSOLE = "ttyS0"
		var err error
		device, partition, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName, bootDir, diskType)
		seedData(baseName, cloudConfig, FILES)
	case "noformat":
		var err error
		device, partition, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName, bootDir, diskType)
	case "raid":
		var err error
		device, partition, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName, bootDir, diskType)
	case "bootstrap":
		CONSOLE = "ttyS0"
		var err error
		device, partition, err = mountdevice(baseName, bootDir, partition, true)
		if err != nil {
			return err
		}
		kernelArgs = kernelArgs + " rancher.cloud_init.datasources=[ec2,gce]"
	case "rancher-upgrade":
		installType = "upgrade" // rancher-upgrade is redundant
		fallthrough
	case "upgrade":
		var err error
		device, partition, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		log.Debugf("upgrading - %s, %s, %s, %s", device, baseName, bootDir, diskType)
		// TODO: detect pv-grub, and don't kill it with syslinux
		upgradeBootloader(device, baseName, bootDir, diskType)
	default:
		return fmt.Errorf("unexpected install type %s", installType)
	}
	kernelArgs = kernelArgs + " console=" + CONSOLE

	if kappend == "" {
		preservedAppend, _ := ioutil.ReadFile(filepath.Join(baseName, bootDir+"append"))
		kappend = string(preservedAppend)
	} else {
		ioutil.WriteFile(filepath.Join(baseName, bootDir+"append"), []byte(kappend), 0644)
	}

	if installType == "amazon-ebs-pv" {
		menu := install.BootVars{
			BaseName: baseName,
			BootDir:  bootDir,
			Timeout:  0,
			Fallback: 0, // need to be conditional on there being a 'rollback'?
			Entries: []install.MenuEntry{
				install.MenuEntry{"RancherOS-current", bootDir, VERSION, kernelArgs, kappend},
			},
		}
		install.PvGrubConfig(menu)
	}
	log.Debugf("installRancher")
	err := installRancher(baseName, bootDir, VERSION, DIST, kernelArgs+" "+kappend)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	log.Debugf("installRancher done")

	// Used by upgrade
	if kexec {
		//    kexec -l ${DIST}/vmlinuz --initrd=${DIST}/initrd --append="${kernelArgs} ${APPEND}" -f
		cmd := exec.Command("kexec", "-l "+DIST+"/vmlinuz",
			"--initrd="+DIST+"/initrd",
			"--append='"+kernelArgs+" "+kappend+"'",
			"-f")
		log.Debugf("Run(%v)", cmd)
		cmd.Stderr = os.Stderr
		if _, err := cmd.Output(); err != nil {
			log.Errorf("Failed to kexec: %s", err)
			return err
		}
		log.Infof("kexec'd to new install")
	}

	return nil
}

// files is an array of 'sourcefile:destination' - but i've not seen any examples of it being used.
func seedData(baseName, cloudData string, files []string) error {
	log.Debugf("seedData")
	_, err := os.Stat(baseName)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Join(baseName, "/var/lib/rancher/conf/cloud-config.d"), 0755); err != nil {
		return err
	}

	if !strings.HasSuffix(cloudData, "empty.yml") {
		if err = dfs.CopyFile(cloudData, baseName+"/var/lib/rancher/conf/cloud-config.d/", filepath.Base(cloudData)); err != nil {
			return err
		}
	}

	for _, f := range files {
		e := strings.Split(f, ":")
		if err = dfs.CopyFile(e[0], baseName, e[1]); err != nil {
			return err
		}
	}
	return nil
}

// set-disk-partitions is called with device ==  **/dev/sda**
func setDiskpartitions(device, diskType string) error {
	log.Debugf("setDiskpartitions")

	d := strings.Split(device, "/")
	if len(d) != 3 {
		return fmt.Errorf("bad device name (%s)", device)
	}
	deviceName := d[2]

	file, err := os.Open("/proc/partitions")
	if err != nil {
		log.Debugf("failed to read /proc/partitions %s", err)
		return err
	}
	defer file.Close()

	exists := false
	haspartitions := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		last := strings.LastIndex(str, " ")

		if last > -1 {
			dev := str[last+1:]

			if strings.HasPrefix(dev, deviceName) {
				if dev == deviceName {
					exists = true
				} else {
					haspartitions = true
				}
			}
		}
	}
	if !exists {
		return fmt.Errorf("disk %s not found: %s", device, err)
	}
	if haspartitions {
		log.Debugf("device %s already partitioned - checking if any are mounted", device)
		file, err := os.Open("/proc/mounts")
		if err != nil {
			log.Errorf("failed to read /proc/mounts %s", err)
			return err
		}
		defer file.Close()
		if partitionMounted(device, file) {
			err = fmt.Errorf("partition %s mounted, cannot repartition", device)
			log.Errorf("%s", err)
			return err
		}

		cmd := exec.Command("system-docker", "ps", "-q")
		var outb bytes.Buffer
		cmd.Stdout = &outb
		if err := cmd.Run(); err != nil {
			log.Printf("ps error: %s", err)
			return err
		}
		for _, image := range strings.Split(outb.String(), "\n") {
			if image == "" {
				continue
			}
			r, w := io.Pipe()
			go func() {
				// TODO: consider a timeout
				// TODO:some of these containers don't have cat / shell
				cmd := exec.Command("system-docker", "exec", image, "cat /proc/mount")
				cmd.Stdout = w
				if err := cmd.Run(); err != nil {
					log.Debugf("%s cat %s", image, err)
				}
				w.Close()
			}()
			if partitionMounted(device, r) {
				err = fmt.Errorf("partition %s mounted in %s, cannot repartition", device, image)
				log.Errorf("k? %s", err)
				return err
			}
		}
	}
	//do it!
	log.Debugf("running dd")
	cmd := exec.Command("dd", "if=/dev/zero", "of="+device, "bs=512", "count=2048")
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("dd error %s", err)
		return err
	}
	log.Debugf("running partprobe")
	cmd = exec.Command("partprobe", device)
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("partprobe error %s", err)
		return err
	}

	log.Debugf("making single RANCHER_STATE partition")
	cmd = exec.Command("parted", "-s", "-a", "optimal", device,
		"mklabel "+diskType, "--",
		"mkpart primary ext4 1 -1")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("parted: %s", err)
		return err
	}
	if err := setBootable(device, diskType); err != nil {
		return err
	}

	return nil
}

func partitionMounted(device string, file io.Reader) bool {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		// /dev/sdb1 /data ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
		ele := strings.Split(str, " ")
		if len(ele) > 5 {
			if strings.HasPrefix(ele[0], device) {
				return true
			}
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("scanner %s", err)
			return false
		}
	}
	return false
}

func formatdevice(device, partition string) error {
	log.Debugf("formatdevice %s", partition)

	//mkfs.ext4 -F -i 4096 -L RANCHER_STATE ${partition}
	// -O ^64bit: for syslinux: http://www.syslinux.org/wiki/index.php?title=Filesystem#ext
	cmd := exec.Command("mkfs.ext4", "-F", "-i", "4096", "-O", "^64bit", "-L", "RANCHER_STATE", partition)
	log.Debugf("Run(%v)", cmd)
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mkfs.ext4: %s", err)
		return err
	}
	return nil
}

func mountdevice(baseName, bootDir, partition string, raw bool) (string, string, error) {
	log.Debugf("mountdevice %s, raw %v", partition, raw)

	if raw {
		log.Debugf("util.Mount (raw) %s, %s", partition, baseName)

		cmd := exec.Command("lsblk", "-no", "pkname", partition)
		log.Debugf("Run(%v)", cmd)
		cmd.Stderr = os.Stderr
		device := ""
		if out, err := cmd.Output(); err == nil {
			device = "/dev/" + strings.TrimSpace(string(out))
		}

		return device, partition, util.Mount(partition, baseName, "", "")
	}

	//rootfs := partition
	// Don't use ResolveDevice - it can fail, whereas `blkid -L LABEL` works more often

	cfg := config.LoadConfig()
	if dev := util.ResolveDevice(cfg.Rancher.State.Dev); dev != "" {
		// try the rancher.state.dev setting
		partition = dev
	} else {
		cmd := exec.Command("blkid", "-L", "RANCHER_BOOT")
		log.Debugf("Run(%v)", cmd)
		cmd.Stderr = os.Stderr
		if out, err := cmd.Output(); err == nil {
			partition = strings.TrimSpace(string(out))
			baseName = filepath.Join(baseName, "boot")
		} else {
			cmd := exec.Command("blkid", "-L", "RANCHER_STATE")
			log.Debugf("Run(%v)", cmd)
			cmd.Stderr = os.Stderr
			if out, err := cmd.Output(); err == nil {
				partition = strings.TrimSpace(string(out))
			}
		}
	}
	device := ""
	cmd := exec.Command("lsblk", "-no", "pkname", partition)
	log.Debugf("Run(%v)", cmd)
	cmd.Stderr = os.Stderr
	if out, err := cmd.Output(); err == nil {
		device = "/dev/" + strings.TrimSpace(string(out))
	}

	log.Debugf("util.Mount %s, %s", partition, baseName)
	os.MkdirAll(baseName, 0755)
	cmd = exec.Command("mount", partition, baseName)
	log.Debugf("Run(%v)", cmd)
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return device, partition, cmd.Run()
}

func formatAndMount(baseName, bootDir, device, partition string) (string, string, error) {
	log.Debugf("formatAndMount")

	err := formatdevice(device, partition)
	if err != nil {
		log.Errorf("formatdevice %s", err)
		return device, partition, err
	}
	device, partition, err = mountdevice(baseName, bootDir, partition, false)
	if err != nil {
		log.Errorf("mountdevice %s", err)
		return device, partition, err
	}
	//err = createbootDirs(baseName, bootDir)
	//if err != nil {
	//	log.Errorf("createbootDirs %s", err)
	//	return bootDir, err
	//}
	return device, partition, nil
}

func NOPEcreatebootDir(baseName, bootDir string) error {
	log.Debugf("createbootDirs")

	if err := os.MkdirAll(filepath.Join(baseName, bootDir+"grub"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(baseName, bootDir+"syslinux"), 0755); err != nil {
		return err
	}
	return nil
}

func setBootable(device, diskType string) error {
	// TODO make conditional - if there is a bootable device already, don't break it
	// TODO: make RANCHER_BOOT bootable - it might not be device 1

	bootflag := "boot"
	if diskType == "gpt" {
		bootflag = "legacy_boot"
	}
	log.Debugf("making device 1 on %s bootable as %s", device, diskType)
	cmd := exec.Command("parted", "-s", "-a", "optimal", device, "set 1 "+bootflag+" on")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("parted: %s", err)
		return err
	}
	return nil
}

func upgradeBootloader(device, baseName, bootDir, diskType string) error {
	log.Debugf("start upgradeBootloader")

	grubDir := filepath.Join(baseName, bootDir+"grub")
	if _, err := os.Stat(grubDir); os.IsNotExist(err) {
		log.Debugf("%s does not exist - no need to upgrade bootloader", grubDir)
		// we've already upgraded
		// TODO: in v0.9.0, need to detect what version syslinux we have
		return nil
	}
	// deal with systems which were previously upgraded, then rolled back, and are now being re-upgraded
	grubBackup := filepath.Join(baseName, bootDir+"grub_backup")
	if err := os.RemoveAll(grubBackup); err != nil {
		log.Errorf("RemoveAll (%s): %s", grubBackup, err)
		return err
	}
	backupSyslinuxDir := filepath.Join(baseName, bootDir+"syslinux_backup")
	if err := os.RemoveAll(backupSyslinuxDir); err != nil {
		log.Errorf("RemoveAll (%s): %s", backupSyslinuxDir, err)
		return err
	}

	if err := os.Rename(grubDir, grubBackup); err != nil {
		log.Errorf("Rename(%s): %s", grubDir, err)
		return err
	}

	syslinuxDir := filepath.Join(baseName, bootDir+"syslinux")
	// it seems that v0.5.0 didn't have a syslinux dir, while 0.7 does
	if _, err := os.Stat(syslinuxDir); !os.IsNotExist(err) {
		if err := os.Rename(syslinuxDir, backupSyslinuxDir); err != nil {
			log.Infof("error Rename(%s, %s): %s", syslinuxDir, backupSyslinuxDir, err)
		} else {
			//mv the old syslinux into linux-previous.cfg
			oldSyslinux, err := ioutil.ReadFile(filepath.Join(backupSyslinuxDir, "syslinux.cfg"))
			if err != nil {
				log.Infof("error read(%s / syslinux.cfg): %s", backupSyslinuxDir, err)
			} else {
				cfg := string(oldSyslinux)
				//DEFAULT RancherOS-current
				//
				//LABEL RancherOS-current
				//    LINUX ../vmlinuz-v0.7.1-rancheros
				//    APPEND rancher.state.dev=LABEL=RANCHER_STATE rancher.state.wait console=tty0 rancher.password=rancher
				//    INITRD ../initrd-v0.7.1-rancheros

				cfg = strings.Replace(cfg, "current", "previous", -1)
				// TODO consider removing the APPEND line - as the global.cfg should have the same result
				ioutil.WriteFile(filepath.Join(baseName, bootDir, "linux-current.cfg"), []byte(cfg), 0644)

				lines := strings.Split(cfg, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "APPEND") {
						// TODO: need to append any extra's the user specified
						ioutil.WriteFile(filepath.Join(baseName, bootDir, "global.cfg"), []byte(cfg), 0644)
						break
					}
				}
			}
		}
	}

	return installSyslinux(device, baseName, bootDir, diskType)
}

func installSyslinux(device, baseName, bootDir, diskType string) error {

	mbrFile := "mbr.bin"
	if diskType == "gpt" {
		mbrFile = "gptmbr.bin"
	}

	//dd bs=440 count=1 if=/usr/lib/syslinux/mbr/mbr.bin of=${device}
	// ubuntu: /usr/lib/syslinux/mbr/mbr.bin
	// alpine: /usr/share/syslinux/mbr.bin
	if device == "/dev/" {
		log.Debugf("installSyslinuxRaid(%s)", device)
		//RAID - assume sda&sdb
		//TODO: fix this - not sure how to detect what disks should have mbr - perhaps we need a param
		//      perhaps just assume and use the devices that make up the raid - mdadm
		device = "/dev/sda"
		if err := setBootable(device, diskType); err != nil {
			log.Errorf("setBootable(%s, %s): %s", device, diskType, err)
			//return err
		}
		cmd := exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/"+mbrFile, "of="+device)
		if err := cmd.Run(); err != nil {
			log.Errorf("%s", err)
			return err
		}
		device = "/dev/sdb"
		if err := setBootable(device, diskType); err != nil {
			log.Errorf("setBootable(%s, %s): %s", device, diskType, err)
			//return err
		}
		cmd = exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/"+mbrFile, "of="+device)
		if err := cmd.Run(); err != nil {
			log.Errorf("%s", err)
			return err
		}
	} else {
		if err := setBootable(device, diskType); err != nil {
			log.Errorf("setBootable(%s, %s): %s", device, diskType, err)
			//return err
		}
		log.Debugf("installSyslinux(%s)", device)
		cmd := exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/"+mbrFile, "of="+device)
		log.Debugf("Run(%v)", cmd)
		if err := cmd.Run(); err != nil {
			log.Errorf("dd: %s", err)
			return err
		}
	}

	sysLinuxDir := filepath.Join(baseName, bootDir, "syslinux")
	if err := os.MkdirAll(sysLinuxDir, 0755); err != nil {
		log.Errorf("MkdirAll(%s)): %s", sysLinuxDir, err)
		//return err
	}

	//cp /usr/lib/syslinux/modules/bios/* ${baseName}/${bootDir}syslinux
	files, _ := ioutil.ReadDir("/usr/share/syslinux/")
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := dfs.CopyFile(filepath.Join("/usr/share/syslinux/", file.Name()), sysLinuxDir, file.Name()); err != nil {
			log.Errorf("copy syslinux: %s", err)
			return err
		}
	}

	//extlinux --install ${baseName}/${bootDir}syslinux
	cmd := exec.Command("extlinux", "--install", sysLinuxDir)
	if device == "/dev/" {
		//extlinux --install --raid ${baseName}/${bootDir}syslinux
		cmd = exec.Command("extlinux", "--install", "--raid", sysLinuxDir)
	}
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	log.Debugf("Run(%v)", cmd)
	if err := cmd.Run(); err != nil {
		log.Errorf("extlinux: %s", err)
		return err
	}
	return nil
}

func installRancher(baseName, bootDir, VERSION, DIST, kappend string) error {
	log.Debugf("installRancher")

	// detect if there already is a linux-current.cfg, if so, move it to linux-previous.cfg,
	currentCfg := filepath.Join(baseName, bootDir, "linux-current.cfg")
	if _, err := os.Stat(currentCfg); !os.IsNotExist(err) {
		previousCfg := filepath.Join(baseName, bootDir, "linux-previous.cfg")
		if _, err := os.Stat(previousCfg); !os.IsNotExist(err) {
			if err := os.Remove(previousCfg); err != nil {
				return err
			}
		}
		os.Rename(currentCfg, previousCfg)
	}

	// The image/ISO have all the files in it - the syslinux cfg's and the kernel&initrd, so we can copy them all from there
	files, _ := ioutil.ReadDir(DIST)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := dfs.CopyFile(filepath.Join(DIST, file.Name()), filepath.Join(baseName, bootDir), file.Name()); err != nil {
			log.Errorf("copy %s: %s", file.Name(), err)
			//return err
		}
		log.Debugf("copied %s to %s as %s", filepath.Join(DIST, file.Name()), filepath.Join(baseName, bootDir), file.Name())
	}
	// the general INCLUDE syslinuxcfg
	if err := dfs.CopyFile(filepath.Join(DIST, "isolinux", "isolinux.cfg"), filepath.Join(baseName, bootDir, "syslinux"), "syslinux.cfg"); err != nil {
		log.Errorf("copy global syslinux.cfgS%s: %s", "syslinux.cfg", err)
		//return err
	}

	// The global.cfg INCLUDE - useful for over-riding the APPEND line
	globalFile := filepath.Join(filepath.Join(baseName, bootDir), "global.cfg")
	if _, err := os.Stat(globalFile); !os.IsNotExist(err) {
		err := ioutil.WriteFile(globalFile, []byte("APPEND "+kappend), 0644)
		if err != nil {
			log.Errorf("write (%s) %s", "global.cfg", err)
			return err
		}
	}
	return nil
}
