package control

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/burmilla/os/cmd/control/install"
	"github.com/burmilla/os/cmd/power"
	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/dfs" // TODO: move CopyFile into util or something.
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"

	"github.com/codegangsta/cli"
	"github.com/pkg/errors"
)

var installCommand = cli.Command{
	Name:     "install",
	Usage:    "install BurmillaOS to disk",
	HideHelp: true,
	Action:   installAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			// TODO: need to validate ? -i burmilla/os:v0.3.1 just sat there.
			Name: "image, i",
			Usage: `install from a certain image (e.g., 'rancher/os:v0.7.0')
							use 'ros os list' to see what versions are available.`,
		},
		cli.StringFlag{
			Name: "install-type, t",
			Usage: `generic:    (Default) Formats the disk and installs BurmillaOS (Syslinux).
                        The boot mode is selected automatically: when booted via UEFI the disk
                        is partitioned as GPT with an EFI system partition and Syslinux EFI is
                        installed, otherwise a single MBR/ext4 partition is used.
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
		cli.StringFlag{
			Name:  "partition, p",
			Usage: "partition to install to",
		},
		cli.StringFlag{
			Name:  "statedir",
			Usage: "install to rancher.state.directory",
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
			Name:  "kexec, k",
			Usage: "reboot using kexec",
		},
		cli.BoolFlag{
			Name:  "save, s",
			Usage: "save services and images for next booting",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Run installer with debug output",
		},
	},
}

func installAction(c *cli.Context) error {
	log.InitLogger()
	debug := c.Bool("debug")
	if debug {
		log.Info("Log level is debug")
		originalLevel := log.GetLevel()
		defer log.SetLevel(originalLevel)
		log.SetLevel(log.DebugLevel)
	}

	if c.Args().Present() {
		log.Fatalf("invalid arguments %v", c.Args())
	}

	kappend := strings.TrimSpace(c.String("append"))
	force := c.Bool("force")
	kexec := c.Bool("kexec")
	reboot := !c.Bool("no-reboot")
	isoinstallerloaded := c.Bool("isoinstallerloaded")

	image := c.String("image")
	cfg := config.LoadConfig()
	if image == "" {
		image = fmt.Sprintf("%s:%s%s",
			cfg.Rancher.Upgrade.Image,
			config.Version,
			config.Suffix)
		image = formatImage(image, cfg)
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
	partition := c.String("partition")
	statedir := c.String("statedir")
	if statedir != "" && installType != "noformat" {
		log.Fatalf("--statedir %s requires --type noformat", statedir)
	}
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
			log.Warn("Cloud-config not provided: you might need to provide cloud-config on boot with ssh_authorized_keys")
		}
	} else {
		os.MkdirAll("/opt", 0755)
		uc := "/opt/user_config.yml"
		if strings.HasPrefix(cloudConfig, "http://") || strings.HasPrefix(cloudConfig, "https://") {
			if err := util.HTTPDownloadToFile(cloudConfig, uc); err != nil {
				log.WithFields(log.Fields{"cloudConfig": cloudConfig, "error": err}).Fatal("Failed to http get cloud-config")
			}
		} else {
			if err := util.FileCopy(cloudConfig, uc); err != nil {
				log.WithFields(log.Fields{"cloudConfig": cloudConfig, "error": err}).Fatal("Failed to copy cloud-config")
			}
		}
		cloudConfig = uc
	}

	savedImages := []string{}
	if c.Bool("save") && cloudConfig != "" && installType != "upgrade" {
		savedImages = install.GetCacheImageList(cloudConfig, cfg)
		log.Debugf("Will cache these images: %s", savedImages)
	}

	if err := runInstall(image, installType, cloudConfig, device, partition, statedir, kappend, force, kexec, isoinstallerloaded, debug, savedImages); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to run install")
		return err
	}

	if !kexec && reboot && (force || yes("Continue with reboot")) {
		log.Info("Rebooting")
		power.Reboot()
	}

	return nil
}

func runInstall(image, installType, cloudConfig, device, partition, statedir, kappend string, force, kexec, isoinstallerloaded, debug bool, savedImages []string) error {
	fmt.Printf("Installing from %s\n", image)

	if !force {
		if util.IsRunningInTty() && !yes("Continue") {
			log.Infof("Not continuing with installation due to user not saying 'yes'")
			os.Exit(1)
		}
	}

	if install.IsEFIFirmware() && partition != "" &&
		(installType == "generic" || installType == "syslinux") {
		return fmt.Errorf("--partition can not be used when booted in UEFI mode - a whole disk (-d) is needed for the EFI system partition")
	}

	useIso := false
	// --isoinstallerloaded is used if the ros has created the installer container from and image that was on the booted iso
	if !isoinstallerloaded {
		log.Infof("start !isoinstallerloaded")

		if _, err := os.Stat("/dist/initrd-" + config.Version); os.IsNotExist(err) {
			deviceName, deviceType, err := getBootIso()
			if err != nil {
				log.Errorf("Failed to get boot iso: %v", err)
				fmt.Println("There is no boot iso drive, terminate the task")
				return err
			}
			if err = mountBootIso(deviceName, deviceType); err != nil {
				log.Debugf("Failed to mountBootIso: %v", err)
			} else {
				log.Infof("trying to load /bootiso/rancheros/installer.tar.gz")
				if _, err := os.Stat("/bootiso/rancheros/"); err == nil {
					cmd := exec.Command("system-docker", "load", "-i", "/bootiso/rancheros/installer.tar.gz")
					cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
					if err := cmd.Run(); err != nil {
						log.Infof("failed to load images from /bootiso/rancheros: %v", err)
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
			if debug {
				installerCmd = append(installerCmd, "--debug")
			}
			if partition != "" {
				installerCmd = append(installerCmd, "--partition", partition)
			}
			if statedir != "" {
				installerCmd = append(installerCmd, "--statedir", statedir)
			}
			if len(savedImages) > 0 {
				installerCmd = append(installerCmd, "--save")
			}

			// TODO: mount at /mnt for shared mount?
			if useIso {
				util.Unmount("/bootiso")
			}
			cmd := exec.Command("system-docker", installerCmd...)
			log.Debugf("Run(%v)", cmd)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

			return cmd.Run()
		}
	}

	log.Debugf("running installation")

	if partition == "" {
		if installType == "generic" ||
			installType == "syslinux" {
			// select the boot mode based on how this system was booted
			efi := install.IsEFIFirmware()
			log.Debugf("running setDiskpartitions (efi: %v)", efi)
			err := setDiskpartitions(device, efi)
			if err != nil {
				log.Errorf("error setDiskpartitions %s", err)
				return err
			}
			// use the bind mounted host filesystem to get access to the /dev/vda1 device that udev on the host sets up (TODO: can we run a udevd inside the container? `mknod b 253 1 /dev/vda1` doesn't work)
			device = "/host" + device
			//# TODO: Change this to a number so that users can specify.
			//# Will need to make it so that our builds and packer APIs remain consistent.
			if efi {
				// partition 1 is the EFI system partition, the state lives on partition 2
				partition = install.GetPartition(device, 2)
			} else {
				partition = install.GetDefaultPartition(device)
			}
		}
	}

	if installType == "upgrade" {
		isoinstallerloaded = false
	}

	if isoinstallerloaded {
		log.Debugf("running isoinstallerloaded...")
		// TODO: detect if its not mounted and then optionally mount?
		deviceName, deviceType, err := getBootIso()
		if err != nil {
			log.Errorf("Failed to get boot iso: %v", err)
			fmt.Println("There is no boot iso drive, terminate the task")
			return err
		}
		if err := mountBootIso(deviceName, deviceType); err != nil {
			log.Errorf("error mountBootIso %s", err)
			//return err
		}
	}

	err := layDownOS(image, installType, cloudConfig, device, partition, statedir, kappend, kexec)
	if err != nil {
		log.Errorf("error layDownOS %s", err)
		return err
	}

	if len(savedImages) > 0 {
		return install.RunCacheScript(partition, savedImages)
	}

	return nil
}

func getDeviceByLabel(label string) (string, string) {
	d, t, err := util.Blkid(label)
	if err != nil {
		log.Warnf("Failed to run blkid for %s", label)
		return "", ""
	}
	return d, t
}

func getBootIso() (string, string, error) {
	deviceName := "/dev/sr0"
	deviceType := "iso9660"

	// Our ISO LABEL is BurmillaOS (older releases used RancherOS)
	// But some tools(like rufus) will change LABEL to upper case
	for _, label := range []string{"RancherOS", "RANCHEROS", "BurmillaOS", "BURMILLAOS"} {
		d, t := getDeviceByLabel(label)
		if d != "" {
			deviceName = d
			deviceType = t
			continue
		}
	}

	// Check the sr deive if exist
	if _, err := os.Stat(deviceName); os.IsNotExist(err) {
		return "", "", err
	}

	return deviceName, deviceType, nil
}

func mountBootIso(deviceName, deviceType string) error {
	mountsFile, err := os.Open("/proc/mounts")
	if err != nil {
		return errors.Wrap(err, "Failed to read /proc/mounts")
	}
	defer mountsFile.Close()

	if partitionMounted(deviceName, mountsFile) {
		return nil
	}

	os.MkdirAll("/bootiso", 0755)
	cmd := exec.Command("mount", "-t", deviceType, deviceName, "/bootiso")
	log.Debugf("mount (%#v)", cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Tried and failed to mount %s: stderr output: %s", deviceName, errBuf.String())
	}
	log.Debugf("Mounted %s, output: %s", deviceName, outBuf.String())
	return nil
}

func layDownOS(image, installType, cloudConfig, device, partition, statedir, kappend string, kexec bool) error {
	// ENV == installType
	//[[ "$ARCH" == "arm" && "$ENV" != "upgrade" ]] && ENV=arm

	// image == burmilla/os:v0.7.0_arm
	// TODO: remove the _arm suffix (but watch out, its not always there..)
	VERSION := image[strings.Index(image, ":")+1:]

	var FILES []string
	DIST := "/dist" //${DIST:-/dist}
	//cloudConfig := SCRIPTS_DIR + "/conf/empty.yml" //${cloudConfig:-"${SCRIPTS_DIR}/conf/empty.yml"}
	CONSOLE := "tty0"
	baseName := "/mnt/new_img"
	kernelArgs := "printk.devkmsg=on rancher.state.dev=LABEL=RANCHER_STATE rancher.state.wait transparent_hugepage=madvise iommu=pt intel_iommu=off scsi_mod.use_blk_mq=1 apparmor=1 security=apparmor panic=10" // console="+CONSOLE
	if statedir != "" {
		kernelArgs = kernelArgs + " rancher.state.directory=" + statedir
	}

	// unmount on trap - a separate boot partition (if any) is mounted below the state partition
	defer util.Unmount(baseName)
	defer util.Unmount(filepath.Join(baseName, config.BootDir))

	// efi tells whether the Syslinux EFI layout (EFI system partition) is used
	// instead of the traditional BIOS/MBR one
	efi := false

	switch installType {
	case "syslinux":
		fallthrough
	case "generic":
		// select the boot mode based on how this system was booted
		efi = install.IsEFIFirmware()
		log.Debugf("formatAndMount (efi: %v)", efi)
		var err error
		device, _, err = formatAndMount(baseName, device, partition)
		if err != nil {
			log.Errorf("formatAndMount %s", err)
			return err
		}
		if efi {
			err = formatAndMountESP(baseName, device)
			if err != nil {
				log.Errorf("formatAndMountESP %s", err)
				return err
			}
			err = installSyslinuxEFI(baseName)
		} else {
			err = installSyslinux(device, baseName)
		}
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
		_, _, err = formatAndMount(baseName, device, partition)
		if err != nil {
			return err
		}
		seedData(baseName, cloudConfig, FILES)
	case "amazon-ebs-hvm":
		CONSOLE = "ttyS0"
		var err error
		device, _, err = formatAndMount(baseName, device, partition)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName)
		//# AWS Networking recommends disabling.
		seedData(baseName, cloudConfig, FILES)
	case "googlecompute":
		CONSOLE = "ttyS0"
		var err error
		device, _, err = formatAndMount(baseName, device, partition)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName)
		seedData(baseName, cloudConfig, FILES)
	case "noformat":
		var err error
		device, _, err = install.MountDevice(baseName, device, partition, false)
		if err != nil {
			return err
		}
		efi = isEFILayout(baseName)
		if efi {
			installSyslinuxEFI(baseName)
		} else {
			installSyslinux(device, baseName)
		}
		if err := os.MkdirAll(filepath.Join(baseName, statedir), 0755); err != nil {
			return err
		}
		err = seedData(baseName, cloudConfig, FILES)
		if err != nil {
			log.Errorf("seedData %s", err)
			return err
		}
	case "raid":
		var err error
		device, _, err = install.MountDevice(baseName, device, partition, false)
		if err != nil {
			return err
		}
		installSyslinux(device, baseName)
	case "bootstrap":
		CONSOLE = "ttyS0"
		var err error
		_, _, err = install.MountDevice(baseName, device, partition, true)
		if err != nil {
			return err
		}
		kernelArgs = kernelArgs + " rancher.cloud_init.datasources=[ec2,gce]"
	case "rancher-upgrade":
		installType = "upgrade" // rancher-upgrade is redundant
		fallthrough
	case "upgrade":
		var err error
		device, _, err = install.MountDevice(baseName, device, partition, false)
		if err != nil {
			return err
		}
		// keep whatever boot layout the installed system already uses
		efi = isEFILayout(baseName)
		log.Debugf("upgrading - %s, %s, efi: %v", device, baseName, efi)
		if efi {
			// The EFI system partition has to be the filesystem which was
			// just mounted on <baseName>/boot. If it is not, installRancher
			// would write the new kernel, initrd and cfgs into a plain
			// directory on the state partition, the ESP the firmware boots
			// from would keep the previous version, and the upgrade would
			// look like it succeeded.
			if err := install.VerifyBootMounted(baseName); err != nil {
				log.Errorf("%s", err)
				return err
			}
			// installRancher only refreshes the kernel, initrd and the cfg
			// files, so update Syslinux EFI itself here - otherwise the ESP
			// keeps the bootloader and the com32 modules of the version the
			// system was originally installed with.
			if err := upgradeSyslinuxEFI(baseName); err != nil {
				log.Errorf("upgradeSyslinuxEFI %s", err)
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected install type %s", installType)
	}
	if efi {
		// mount the EFI system partition to /boot on the installed system
		kernelArgs = kernelArgs + " rancher.state.boot_dev=LABEL=RANCHER_EFI rancher.state.boot_fstype=vfat"
	}
	kernelArgs = kernelArgs + " console=" + CONSOLE

	if kappend == "" {
		preservedAppend, _ := ioutil.ReadFile(filepath.Join(baseName, config.BootDir, "append"))
		kappend = string(preservedAppend)
	} else {
		ioutil.WriteFile(filepath.Join(baseName, config.BootDir, "append"), []byte(kappend), 0644)
	}

	log.Debugf("installRancher")
	_, err := installRancher(baseName, VERSION, DIST, kernelArgs+" "+kappend, efi)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	log.Debugf("installRancher done")

	if kexec {
		power.Kexec(false, filepath.Join(baseName, config.BootDir), kernelArgs+" "+kappend)
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

	stateSeedDir := "state_seed"
	cloudConfigBase := "/var/lib/rancher/conf/cloud-config.d"
	cloudConfigDir := ""

	// If there is a separate boot partition (RANCHER_BOOT or the RANCHER_EFI
	// EFI system partition), cloud-config should be written to RANCHER_STATE partition.
	bootPartition, _ := install.GetBootPartition()
	if bootPartition != "" {
		stateSeedFullPath := filepath.Join(baseName, stateSeedDir)
		if err = os.MkdirAll(stateSeedFullPath, 0700); err != nil {
			return err
		}

		defer util.Unmount(stateSeedFullPath)

		statePartition := install.GetStatePartition()
		cmd := exec.Command("mount", statePartition, stateSeedFullPath)
		//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		log.Debugf("seedData: mount %s to %s", statePartition, stateSeedFullPath)
		if err = cmd.Run(); err != nil {
			return err
		}

		cloudConfigDir = filepath.Join(baseName, stateSeedDir, cloudConfigBase)
	} else {
		cloudConfigDir = filepath.Join(baseName, cloudConfigBase)
	}

	if err = os.MkdirAll(cloudConfigDir, 0700); err != nil {
		return err
	}

	if !strings.HasSuffix(cloudData, "empty.yml") {
		if err = dfs.CopyFile(cloudData, cloudConfigDir, filepath.Base(cloudData)); err != nil {
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
func setDiskpartitions(device string, efi bool) error {
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
	log.Debugf("running dd device: %s", device)
	cmd := exec.Command("dd", "if=/dev/zero", "of="+device, "bs=512", "count=2048")
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("dd error %s", err)
		return err
	}
	log.Debugf("running partprobe: %s", device)
	cmd = exec.Command("partprobe", device)
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to partprobe device %s: %v", device, err)
		return err
	}

	if efi {
		// UEFI needs a GPT disk with a FAT32 EFI system partition for Syslinux EFI,
		// the kernel and the initrd. The rest of the disk is used for RANCHER_STATE.
		log.Debugf("making GPT disk with ESP and RANCHER_STATE partitions, device: %s", device)
		cmd = exec.Command("parted", "-s", "-a", "optimal", device,
			"mklabel gpt", "--",
			"mkpart ESP fat32 1MiB 513MiB",
			"set 1 esp on",
			"mkpart primary ext4 513MiB 100%")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Failed to parted device %s: %v", device, err)
			return err
		}
		return nil
	}

	log.Debugf("making single RANCHER_STATE partition, device: %s", device)
	cmd = exec.Command("parted", "-s", "-a", "optimal", device,
		"mklabel msdos", "--",
		"mkpart primary ext4 1 -1")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to parted device %s: %v", device, err)
		return err
	}
	return setBootable(device)
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
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mkfs.ext4: %s", err)
		return err
	}
	return nil
}

func formatAndMount(baseName, device, partition string) (string, string, error) {
	log.Debugf("formatAndMount")

	err := formatdevice(device, partition)
	if err != nil {
		log.Errorf("formatdevice %s", err)
		return device, partition, err
	}
	device, partition, err = install.MountDevice(baseName, device, partition, false)
	if err != nil {
		log.Errorf("mountdevice %s", err)
		return device, partition, err
	}
	return device, partition, nil
}

// formatAndMountESP formats partition 1 as the FAT32 EFI system partition and
// mounts it on the boot dir of the already mounted state partition.
// FAT labels are limited to 11 characters, so RANCHER_EFI is used instead of RANCHER_BOOT.
func formatAndMountESP(baseName, device string) error {
	esp := install.GetPartition(device, 1)
	log.Debugf("formatting EFI system partition %s", esp)

	cmd := exec.Command("mkfs.vfat", "-F", "32", "-n", "RANCHER_EFI", esp)
	log.Debugf("Run(%v)", cmd)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mkfs.vfat: %s", err)
		return err
	}

	bootDir := filepath.Join(baseName, config.BootDir)
	if err := os.MkdirAll(bootDir, 0755); err != nil {
		log.Errorf("MkdirAll(%s): %s", bootDir, err)
		return err
	}
	cmd = exec.Command("mount", esp, bootDir)
	log.Debugf("Run(%v)", cmd)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount %s: %s", esp, err)
		return err
	}
	return nil
}

// alpine: /usr/share/syslinux/efi64 (syslinux.efi, ldlinux.e64 and *.c32 modules)
const syslinuxEFIDir = "/usr/share/syslinux/efi64"

// isEFILayout detects whether the (already mounted) target uses the Syslinux EFI
// boot layout - either an EFI directory on the boot partition, or an EFI
// system partition that has not been populated yet
func isEFILayout(baseName string) bool {
	if _, err := os.Stat(filepath.Join(baseName, config.BootDir, "EFI")); err == nil {
		return true
	}
	d, _ := install.ResolveLabel(install.EFILabel)
	return d != ""
}

// upgradeSyslinuxEFI refreshes the Syslinux EFI bootloader on the EFI system
// partition of an installed system. A missing source directory is not fatal:
// the bootloader which is already on the ESP keeps working, and the cfg files
// installRancher writes are what selects the new version.
func upgradeSyslinuxEFI(baseName string) error {
	if _, err := os.Stat(syslinuxEFIDir); os.IsNotExist(err) {
		log.Errorf("%s does not exist, not updating the EFI bootloader", syslinuxEFIDir)
		return nil
	}
	return installSyslinuxEFI(baseName)
}

// installSyslinuxEFI copies the Syslinux EFI binary and its modules to the
// EFI/BOOT fallback path of the EFI system partition, which firmwares use
// when no boot entry exists in NVRAM
func installSyslinuxEFI(baseName string) error {
	log.Debugf("installSyslinuxEFI(%s)", baseName)

	efiBootDir := filepath.Join(baseName, config.BootDir, "EFI", "BOOT")
	if err := os.MkdirAll(efiBootDir, 0755); err != nil {
		log.Errorf("MkdirAll(%s): %s", efiBootDir, err)
		return err
	}

	srcDir := syslinuxEFIDir
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Errorf("ReadDir(%s): %s", srcDir, err)
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if name == "syslinux.efi" {
			name = "bootx64.efi"
		}
		// overwrite: on an upgrade the ESP still holds the bootloader and the
		// com32 modules of the previously installed version, and they have to
		// stay in sync with the syslinux.cfg written by installRancher
		if err := dfs.CopyFileOverwrite(filepath.Join(srcDir, file.Name()), efiBootDir, name, true); err != nil {
			log.Errorf("copy syslinux efi: %s", err)
			return err
		}
	}
	return nil
}

// efiCfg rewrites a Syslinux cfg for use on the EFI system partition: the
// BIOS cfgs reference files relative to the syslinux/ (or boot/isolinux/)
// directory with "../", while syslinux.efi reads its cfg from EFI/BOOT and
// the kernel, initrd and cfg files live in the root of the partition.
func efiCfg(src string) ([]byte, error) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}
	return bytes.Replace(data, []byte("../"), []byte("/"), -1), nil
}

func copyEFICfg(src, folder, name string, overwrite bool) error {
	target := filepath.Join(folder, name)
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			return nil
		}
	}
	data, err := efiCfg(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(folder, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(target, data, 0644)
}

func setBootable(device string) error {
	// TODO make conditional - if there is a bootable device already, don't break it
	// TODO: make RANCHER_BOOT bootable - it might not be device 1

	log.Debugf("making device 1 on %s bootable", device)
	cmd := exec.Command("parted", "-s", "-a", "optimal", device, "set 1 boot on")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("parted: %s", err)
		return err
	}
	return nil
}

func installSyslinux(device, baseName string) error {
	log.Debugf("installSyslinux(%s)", device)

	mbrFile := "mbr.bin"

	//dd bs=440 count=1 if=/usr/lib/syslinux/mbr/mbr.bin of=${device}
	// ubuntu: /usr/lib/syslinux/mbr/mbr.bin
	// alpine: /usr/share/syslinux/mbr.bin
	if device == "/dev/" {
		log.Debugf("installSyslinuxRaid(%s)", device)
		//RAID - assume sda&sdb
		//TODO: fix this - not sure how to detect what disks should have mbr - perhaps we need a param
		//      perhaps just assume and use the devices that make up the raid - mdadm
		device = "/dev/sda"
		if err := setBootable(device); err != nil {
			log.Errorf("setBootable(%s): %s", device, err)
			//return err
		}
		cmd := exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/"+mbrFile, "of="+device)
		if err := cmd.Run(); err != nil {
			log.Errorf("%s", err)
			return err
		}
		device = "/dev/sdb"
		if err := setBootable(device); err != nil {
			log.Errorf("setBootable(%s): %s", device, err)
			//return err
		}
		cmd = exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/"+mbrFile, "of="+device)
		if err := cmd.Run(); err != nil {
			log.Errorf("%s", err)
			return err
		}
	} else {
		if err := setBootable(device); err != nil {
			log.Errorf("setBootable(%s): %s", device, err)
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

	sysLinuxDir := filepath.Join(baseName, config.BootDir, "syslinux")
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

func different(existing, new string, efi bool) bool {
	// assume existing file exists
	if _, err := os.Stat(new); os.IsNotExist(err) {
		return true
	}
	data, err := ioutil.ReadFile(existing)
	if err != nil {
		return true
	}
	var newData []byte
	if efi {
		// the existing cfg was rewritten for the ESP layout on install
		newData, err = efiCfg(new)
	} else {
		newData, err = ioutil.ReadFile(new)
	}
	if err != nil {
		return true
	}
	md5sum := md5.Sum(data)
	newmd5sum := md5.Sum(newData)
	if md5sum != newmd5sum {
		return true
	}
	return false
}

func installRancher(baseName, VERSION, DIST, kappend string, efi bool) (string, error) {
	log.Debugf("installRancher (efi: %v)", efi)

	// detect if there already is a linux-current.cfg, if so, move it to linux-previous.cfg,
	currentCfg := filepath.Join(baseName, config.BootDir, "linux-current.cfg")
	if _, err := os.Stat(currentCfg); !os.IsNotExist(err) {
		existingCfg := filepath.Join(DIST, "linux-current.cfg")
		// only remove previous if there is a change to the current
		if different(currentCfg, existingCfg, efi) {
			previousCfg := filepath.Join(baseName, config.BootDir, "linux-previous.cfg")
			if _, err := os.Stat(previousCfg); !os.IsNotExist(err) {
				if err := os.Remove(previousCfg); err != nil {
					return currentCfg, err
				}
			}
			if err := os.Rename(currentCfg, previousCfg); err != nil {
				return currentCfg, err
			}
			// TODO: now that we're parsing syslinux.cfg files, maybe we can delete old kernels and initrds
		}
	}

	// The image/ISO have all the files in it - the syslinux cfg's and the kernel&initrd, so we can copy them all from there
	// A failure here has to be fatal: leaving the kernel, the initrd or a cfg
	// file behind gives a boot menu which still points at the previous version.
	files, _ := ioutil.ReadDir(DIST)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// TODO: should overwrite anything other than the global.cfg
		overwrite := true
		if file.Name() == "global.cfg" {
			overwrite = false
		}
		if efi && strings.HasSuffix(file.Name(), ".cfg") {
			if err := copyEFICfg(filepath.Join(DIST, file.Name()), filepath.Join(baseName, config.BootDir), file.Name(), overwrite); err != nil {
				log.Errorf("copy %s: %s", file.Name(), err)
				return currentCfg, err
			}
			continue
		}
		if err := dfs.CopyFileOverwrite(filepath.Join(DIST, file.Name()), filepath.Join(baseName, config.BootDir), file.Name(), overwrite); err != nil {
			log.Errorf("copy %s: %s", file.Name(), err)
			return currentCfg, err
		}
	}

	// the general INCLUDE syslinuxcfg
	isolinuxFile := filepath.Join(DIST, "isolinux", "isolinux.cfg")
	if efi {
		efiBootDir := filepath.Join(baseName, config.BootDir, "EFI", "BOOT")
		if err := copyEFICfg(isolinuxFile, efiBootDir, "syslinux.cfg", true); err != nil {
			log.Errorf("copy global syslinux.cfg %s: %s", "syslinux.cfg", err)
			return currentCfg, err
		}
		log.Debugf("installRancher copy global EFI syslinux.cfg OK")
	} else {
		syslinuxDir := filepath.Join(baseName, config.BootDir, "syslinux")
		if err := dfs.CopyFileOverwrite(isolinuxFile, syslinuxDir, "syslinux.cfg", true); err != nil {
			log.Errorf("copy global syslinux.cfgS%s: %s", "syslinux.cfg", err)
			return currentCfg, err
		}
		log.Debugf("installRancher copy global syslinux.cfgS OK")
	}

	// The global.cfg INCLUDE - useful for over-riding the APPEND line
	globalFile := filepath.Join(baseName, config.BootDir, "global.cfg")
	if _, err := os.Stat(globalFile); !os.IsNotExist(err) {
		err := ioutil.WriteFile(globalFile, []byte("APPEND "+kappend), 0644)
		if err != nil {
			log.Errorf("write (%s) %s", "global.cfg", err)
			return currentCfg, err
		}
	}
	return currentCfg, nil
}
