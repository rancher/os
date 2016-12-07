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
	"strings"
	"text/template"

	"github.com/rancher/os/log"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs" // TODO: move CopyFile into util or something.
	"github.com/rancher/os/util"
)

type MenuEntry struct {
	Name, bootDir, Version, KernelArgs, Append string
}
type bootVars struct {
	baseName, bootDir string
	Timeout           uint
	Fallback          int
	Entries           []MenuEntry
}

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
			Usage: `generic:    (Default) Creates 1 ext4 partition and installs RancherOS
                        amazon-ebs: Installs RancherOS and sets up PV-GRUB
                        syslinux: partition and format disk (mbr), then install RancherOS and setup Syslinux
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
		cli.BoolFlag{
			Name:  "mountiso",
			Usage: "mount the iso to get kernel and initrd",
		},
	},
}

func installAction(c *cli.Context) error {
	if c.Args().Present() {
		log.Fatalf("invalid arguments %v", c.Args())
	}
	device := c.String("device")
	if device == "" {
		log.Fatal("Can not proceed without -d <dev> specified")
	}

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

	cloudConfig := c.String("cloud-config")
	if cloudConfig == "" {
		log.Warn("Cloud-config not provided: you might need to provide cloud-config on bootDir with ssh_authorized_keys")
	} else {
		uc := "/opt/user_config.yml"
		if err := util.FileCopy(cloudConfig, uc); err != nil {
			log.WithFields(log.Fields{"cloudConfig": cloudConfig}).Fatal("Failed to copy cloud-config")
		}
		cloudConfig = uc
	}

	kappend := strings.TrimSpace(c.String("append"))
	force := c.Bool("force")
	reboot := !c.Bool("no-reboot")
	mountiso := c.Bool("mountiso")

	if err := runInstall(image, installType, cloudConfig, device, kappend, force, reboot, mountiso); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to run install")
	}

	return nil
}

func runInstall(image, installType, cloudConfig, device, kappend string, force, reboot, mountiso bool) error {
	fmt.Printf("Installing from %s\n", image)

	if !force {
		if !yes("Continue") {
			log.Infof("Not continuing with installation due to user not saying 'yes'")
			os.Exit(1)
		}
	}
	diskType := "msdos"
	if installType == "gptsyslinux" {
		diskType = "gpt"
	}

	if installType == "generic" ||
		installType == "syslinux" ||
		installType == "gptsyslinux" {

	// TODO: generalise to versions before 0.8.0-rc2
	if image == "rancher/os:v0.7.0" {
		log.Infof("starting installer container for %s", image)
		if installType == "generic" {
			cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=all-volumes",
				"--entrypoint=/scripts/set-disk-partitions", image, device, diskType)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=user-volumes",
			"--volumes-from=command-volumes", image, "-d", device, "-t", installType, "-c", cloudConfig, "-a", kappend)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	useIso := false
	if !mountiso {
		if _, err := os.Stat("/dist/initrd"); os.IsNotExist(err) {
			log.Infof("trying to mount /dev/sr0 and then load image")

			if err = mountBootIso(); err == nil {
				log.Infof("Mounted /dev/sr0")
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
						// TODO: fix the fullinstaller Dockerfile to use the ${VERSION}${SUFFIX}
						image = cfg.Rancher.Upgrade.Image + "-installer" + ":latest"
					}
				}
				// TODO: also poke around looking for the /boot/vmlinuz and initrd...
			}

			log.Infof("starting installer container for %s (new)", image)
			installerCmd := []string{
				"run", "--rm", "--net=host", "--privileged",
				// bind mount host fs to access its ros, vmlinuz, initrd and /dev (udev isn't running in container)
				"-v", "/:/host",
				"--volumes-from=user-volumes", "--volumes-from=command-volumes",
				image,
				"install",
				"-t", installType,
				"-d", device,
			}
			if force {
				installerCmd = append(installerCmd, "-f")
			}
			if !reboot {
				installerCmd = append(installerCmd, "--no-reboot")
			}
			if cloudConfig != "" {
				installerCmd = append(installerCmd, "-c", cloudConfig)
			}
			if kappend != "" {
				installerCmd = append(installerCmd, "-a", kappend)
			}
			if useIso {
				installerCmd = append(installerCmd, "--mountiso")
			}

			cmd := exec.Command("system-docker", installerCmd...)
			log.Debugf("Run(%v)", cmd)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				if useIso {
					util.Unmount("/bootiso")
				}
				return err
			}
			if useIso {
				util.Unmount("/bootiso")
			}
			return nil
		}
	}

	// TODO: needs to pass the log level on to the container
	log.InitLogger()
	log.SetLevel(log.DebugLevel)

	log.Infof("running installation")

	if installType == "generic" {
		log.Infof("running setDiskpartitions")
		err := setDiskpartitions(device)
		if err != nil {
			log.Errorf("error setDiskpartitions %s", err)
			return err
		}
		// use the bind mounted host filesystem to get access to the /dev/vda1 device that udev on the host sets up (TODO: can we run a udevd inside the container? `mknod b 253 1 /dev/vda1` doesn't work)
		device = "/host" + device
		log.Infof("done setDiskpartitions")
	}

	if mountiso {
		// TODO: I hope to remove this from here later.
		if err := mountBootIso(); err == nil {
			log.Infof("Mounted /dev/sr0")
		}
	}

	log.Infof("running layDownOS")
	err := layDownOS(image, installType, cloudConfig, device, kappend)
	if err != nil {
		log.Infof("error layDownOS %s", err)
		return err
	}

	if reboot && (force || yes("Continue with reboot")) {
		log.Info("Rebooting")
		power.Reboot()
	}

	return nil
}

func mountBootIso() error {
	// TODO: need to add a label to the iso and mount using that.
	//		ARGH! need to mount this in the host - or share it as a volume..
	os.MkdirAll("/bootiso", 0755)
	cmd := exec.Command("mount", "-t", "iso9660", "/dev/sr0", "/bootiso")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Infof("tried and failed to mount /dev/sr0: %s", err)
	} else {
		log.Infof("Mounted /dev/sr0")
	}
	return err
}

func layDownOS(image, installType, cloudConfig, device, kappend string) error {
	log.Infof("layDownOS")
	// ENV == installType
	//[[ "$ARCH" == "arm" && "$ENV" != "rancher-upgrade" ]] && ENV=arm

	// image == rancher/os:v0.7.0_arm
	// TODO: remove the _arm suffix (but watch out, its not always there..)
	VERSION := image[strings.Index(image, ":")+1:]

	var FILES []string
	DIST := "/dist" //${DIST:-/dist}
	//cloudConfig := SCRIPTS_DIR + "/conf/empty.yml" //${cloudConfig:-"${SCRIPTS_DIR}/conf/empty.yml"}
	CONSOLE := "tty0"
	baseName := "/mnt/new_img"
	bootDir := "boot/" // set by mountdevice
	//# TODO: Change this to a number so that users can specify.
	//# Will need to make it so that our builds and packer APIs remain consistent.
	partition := device + "1"                                                //${partition:=${device}1}
	kernelArgs := "rancher.state.dev=LABEL=RANCHER_STATE rancher.state.wait" // console="+CONSOLE

	// unmount on trap
	defer util.Unmount(baseName)

	switch installType {
	case "generic":
		log.Infof("formatAndMount")
		var err error
		bootDir, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		//log.Infof("installGrub")
		//err = installGrub(baseName, device)
		log.Infof("installSyslinux")
		err = installSyslinux(device, baseName, bootDir)

		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		log.Infof("seedData")
		err = seedData(baseName, cloudConfig, FILES)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		log.Infof("seedData done")
	case "arm":
		var err error
		bootDir, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		seedData(baseName, cloudConfig, FILES)
	case "amazon-ebs-pv":
		fallthrough
	case "amazon-ebs-hvm":
		CONSOLE = "ttyS0"
		var err error
		bootDir, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		if installType == "amazon-ebs-hvm" {
			installGrub(baseName, device)
		}
		//# AWS Networking recommends disabling.
		seedData(baseName, cloudConfig, FILES)
	case "googlecompute":
		CONSOLE = "ttyS0"
		var err error
		bootDir, err = formatAndMount(baseName, bootDir, device, partition)
		if err != nil {
			return err
		}
		installGrub(baseName, device)
		seedData(baseName, cloudConfig, FILES)
	case "noformat":
		var err error
		bootDir, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		createbootDirs(baseName, bootDir)
		installSyslinux(device, baseName, bootDir)
	case "raid":
		var err error
		bootDir, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		createbootDirs(baseName, bootDir)
		installSyslinuxRaid(baseName, bootDir)
	case "bootstrap":
		CONSOLE = "ttyS0"
		var err error
		bootDir, err = mountdevice(baseName, bootDir, partition, true)
		if err != nil {
			return err
		}
		createbootDirs(baseName, bootDir)
		kernelArgs = kernelArgs + " rancher.cloud_init.datasources=[ec2,gce]"
	case "rancher-upgrade":
		var err error
		bootDir, err = mountdevice(baseName, bootDir, partition, false)
		if err != nil {
			return err
		}
		createbootDirs(baseName, bootDir)
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

	//menu := bootVars{
	//	baseName: baseName,
	//	bootDir:  bootDir,
	//	Timeout:  1,
	//	Fallback: 1, // need to be conditional on there being a 'rollback'?
	//	Entries: []MenuEntry{
	//		MenuEntry{"RancherOS-current", bootDir, VERSION, kernelArgs, kappend},
	//		//			MenuEntry{"RancherOS-rollback", bootDir, ROLLBACK_VERSION, kernelArgs, kappend},
	//	},
	//}

	//log.Debugf("grubConfig")
	//grubConfig(menu)
	//log.Debugf("syslinuxConfig")
	//syslinuxConfig(menu)
	//log.Debugf("pvGrubConfig")
	//pvGrubConfig(menu)
	log.Debugf("installRancher")
	err := installRancher(baseName, bootDir, VERSION, DIST, kernelArgs+" "+kappend)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	log.Debugf("installRancher done")

	//unused by us? :)
	//if [ "$KEXEC" = "y" ]; then
	//    kexec -l ${DIST}/vmlinuz --initrd=${DIST}/initrd --append="${kernelArgs} ${APPEND}" -f
	//fi
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

	if strings.HasSuffix(cloudData, "empty.yml") {
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
func setDiskpartitions(device string) error {
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
		log.Debugf("disk %s not found", device)
		return err
	}
	if haspartitions {
		log.Debugf("device %s already partitioned - checking if any are mounted", device)
		file, err := os.Open("/proc/mounts")
		if err != nil {
			log.Debugf("failed to read /proc/mounts %s", err)
			return err
		}
		defer file.Close()
		if partitionMounted(device, file) {
			err = fmt.Errorf("partition %s mounted, cannot repartition", device)
			log.Printf("%s", err)
			return err
		}
		cmd := exec.Command("system-docker", "ps", "-q")
		var outb bytes.Buffer
		cmd.Stdout = &outb
		if err := cmd.Run(); err != nil {
			log.Printf("%s", err)
			return err
		}
		for _, image := range strings.Split(outb.String(), "\n") {
			if image == "" {
				continue
			}
			r, w := io.Pipe()
			go func() {
				// TODO: consider a timeout
				cmd := exec.Command("system-docker", "exec", image, "cat /proc/mount")
				cmd.Stdout = w
				if err := cmd.Run(); err != nil {
					log.Printf("%s", err)
				}
			}()
			if partitionMounted(device, r) {
				err = fmt.Errorf("partition %s mounted in %s, cannot repartition", device, image)
				log.Printf("%s", err)
				return err
			}
		}
	}
	//do it!
	log.Debugf("running dd")
	cmd := exec.Command("dd", "if=/dev/zero", "of="+device, "bs=512", "count=2048")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}
	log.Debugf("running partprobe")
	cmd = exec.Command("partprobe", device)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}

	r, w := io.Pipe()
	go func() {
		w.Write([]byte(`n
p
1


a
1
w
`))
		w.Close()
	}()
	log.Debugf("running fdisk")
	cmd = exec.Command("fdisk", device)
	cmd.Stdin = r
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
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
			log.Printf("%s", err)
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
		log.Debugf("mkfs.ext4: %s", err)
		return err
	}
	return nil
}

func mountdevice(baseName, bootDir, partition string, raw bool) (string, error) {
	log.Debugf("mountdevice %s, raw %v", partition, raw)

	if raw {
		log.Debugf("util.Mount (raw) %s, %s", partition, baseName)
		return bootDir, util.Mount(partition, baseName, "", "")
	}

	rootfs := partition
	// Don't use ResolveDevice - it can fail, whereas `blkid -L LABEL` works more often
	//if dev := util.ResolveDevice("LABEL=RANCHER_BOOT"); dev != "" {
	cmd := exec.Command("blkid", "-L", "RANCHER_BOOT")
	log.Debugf("Run(%v)", cmd)
	cmd.Stderr = os.Stderr
	if out, err := cmd.Output(); err == nil {
		rootfs = string(out)
	} else {
		cmd := exec.Command("blkid", "-L", "RANCHER_STATE")
		log.Debugf("Run(%v)", cmd)
		cmd.Stderr = os.Stderr
		if out, err := cmd.Output(); err == nil {
			rootfs = string(out)
		}
	}

	log.Debugf("util.Mount %s, %s", rootfs, baseName)
	//	return bootDir, util.Mount(rootfs, baseName, "", "")
	os.MkdirAll(baseName, 0755)
	cmd = exec.Command("mount", rootfs, baseName)
	log.Debugf("Run(%v)", cmd)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return bootDir, cmd.Run()

}

func formatAndMount(baseName, bootDir, device, partition string) (string, error) {
	log.Debugf("formatAndMount")

	err := formatdevice(device, partition)
	if err != nil {
		return bootDir, err
	}
	bootDir, err = mountdevice(baseName, bootDir, partition, false)
	if err != nil {
		return bootDir, err
	}
	err = createbootDirs(baseName, bootDir)
	if err != nil {
		return bootDir, err
	}
	return bootDir, nil
}

func createbootDirs(baseName, bootDir string) error {
	log.Debugf("createbootDirs")

	if err := os.MkdirAll(filepath.Join(baseName, bootDir+"grub"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(baseName, bootDir+"syslinux"), 0755); err != nil {
		return err
	}
	return nil
}

func installSyslinux(device, baseName, bootDir string) error {
	log.Debugf("installSyslinux")

	//dd bs=440 count=1 if=/usr/lib/syslinux/mbr/mbr.bin of=${device}
	// ubuntu: /usr/lib/syslinux/mbr/mbr.bin
	// alpine: /usr/share/syslinux/mbr.bin
	cmd := exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/mbr.bin", "of="+device)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	log.Debugf("Run(%v)", cmd)
	if err := cmd.Run(); err != nil {
		log.Printf("dd: %s", err)
		return err
	}
	//cp /usr/lib/syslinux/modules/bios/* ${baseName}/${bootDir}syslinux
	files, _ := ioutil.ReadDir("/usr/share/syslinux/")
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := dfs.CopyFile(filepath.Join("/usr/share/syslinux/", file.Name()), filepath.Join(baseName, bootDir, "syslinux"), file.Name()); err != nil {
			log.Errorf("copy syslinux: %s", err)
			return err
		}
	}

	//extlinux --install ${baseName}/${bootDir}syslinux
	cmd = exec.Command("extlinux", "--install", filepath.Join(baseName, bootDir+"syslinux"))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	log.Debugf("Run(%v)", cmd)
	if err := cmd.Run(); err != nil {
		log.Printf("extlinuux: %s", err)
		return err
	}
	return nil
}

func installSyslinuxRaid(baseName, bootDir string) error {
	log.Debugf("installSyslinuxRaid")

	//dd bs=440 count=1 if=/usr/lib/syslinux/mbr/mbr.bin of=/dev/sda
	//dd bs=440 count=1 if=/usr/lib/syslinux/mbr/mbr.bin of=/dev/sdb
	//cp /usr/lib/syslinux/modules/bios/* ${baseName}/${bootDir}syslinux
	//extlinux --install --raid ${baseName}/${bootDir}syslinux
	cmd := exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/mbr.bin", "of=/dev/sda")
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}
	cmd = exec.Command("dd", "bs=440", "count=1", "if=/usr/share/syslinux/mbr.bin", "of=/dev/sdb")
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}
	//cp /usr/lib/syslinux/modules/bios/* ${baseName}/${bootDir}syslinux
	files, _ := ioutil.ReadDir("/usr/share/syslinux/")
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := dfs.CopyFile(filepath.Join("/usr/share/syslinux/", file.Name()), filepath.Join(baseName, bootDir, "syslinux"), file.Name()); err != nil {
			log.Errorf("copy syslinux: %s", err)
			return err
		}
	}
	cmd = exec.Command("extlinux", "--install", filepath.Join(baseName, bootDir+"syslinux"))
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}
	return nil
}

func installGrub(baseName, device string) error {
	log.Debugf("installGrub")

	//grub-install --boot-directory=${baseName}/boot ${device}
	cmd := exec.Command("grub-install", "--boot-directory="+baseName+"/boot", device)
	if err := cmd.Run(); err != nil {
		log.Printf("%s", err)
		return err
	}
	return nil
}

func grubConfig(menu bootVars) error {
	log.Debugf("grubConfig")

	filetmpl, err := template.New("grub2config").Parse(`{{define "grub2menu"}}menuentry "{{.Name}}" {
  set root=(hd0,msdos1)
  linux /{{.bootDir}}vmlinuz-{{.Version}}-rancheros {{.KernelArgs}} {{.Append}}
  initrd /{{.bootDir}}initrd-{{.Version}}-rancheros
}

{{end}}
set default="0"
set timeout="{{.Timeout}}"
{{if .Fallback}}set fallback={{.Fallback}}{{end}}

{{- range .Entries}}
{{template "grub2menu" .}}
{{- end}}

`)
	if err != nil {
		log.Errorf("grub2config %s", err)
		return err
	}

	cfgFile := filepath.Join(menu.baseName, menu.bootDir+"grub/grub.cfg")
	log.Debugf("grubConfig written to %s", cfgFile)

	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}
	err = filetmpl.Execute(f, menu)
	if err != nil {
		return err
	}
	return nil
}

func syslinuxConfig(menu bootVars) error {
	log.Debugf("syslinuxConfig")

	filetmpl, err := template.New("syslinuxconfig").Parse(`{{define "syslinuxmenu"}}
LABEL {{.Name}}
    LINUX ../vmlinuz-{{.Version}}-rancheros
    APPEND {{.KernelArgs}} {{.Append}}
    INITRD ../initrd-{{.Version}}-rancheros
{{end}}
TIMEOUT 20   #2 seconds
DEFAULT RancherOS-current

{{- range .Entries}}
{{template "syslinuxmenu" .}}
{{- end}}

`)
	if err != nil {
		log.Errorf("syslinuxconfig %s", err)
		return err
	}

	cfgFile := filepath.Join(menu.baseName, menu.bootDir+"syslinux/syslinux.cfg")
	log.Debugf("syslinuxConfig written to %s", cfgFile)
	f, err := os.Create(cfgFile)
	if err != nil {
		log.Errorf("Create(%s) %s", cfgFile, err)
		return err
	}
	err = filetmpl.Execute(f, menu)
	if err != nil {
		return err
	}
	return nil
}

func installRancher(baseName, bootDir, VERSION, DIST, kappend string) error {
	log.Debugf("installRancher")

	// TODO detect if there already is a linux-current.cfg, if so, move it to linux-previous.cfg, and replace only current with the one in the image/iso

	// The image/ISO have all the files in it - the syslinux cfg's and the kernel&initrd, so we can copy them all from there
	files, _ := ioutil.ReadDir(DIST)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := dfs.CopyFile(filepath.Join(DIST, file.Name()), filepath.Join(baseName, bootDir), file.Name()); err != nil {
			log.Errorf("copy %s: %s", file.Name(), err)
			return err
		}
	}
	// the main syslinuxcfg
	if err := dfs.CopyFile(filepath.Join(DIST, "isolinux", "isolinux.cfg"), filepath.Join(baseName, bootDir, "syslinux"), "syslinux.cfg"); err != nil {
		log.Errorf("copy %s: %s", "syslinux.cfg", err)
		return err
	}
	// The global.cfg INCLUDE - useful for over-riding the APPEND line
	err := ioutil.WriteFile(filepath.Join(filepath.Join(baseName, bootDir), "global.cfg"), []byte("APPEND "+kappend), 0644)
	if err != nil {
		log.Errorf("write (%s) %s", "global.cfg", err)
		return err
	}
	return nil
}

func pvGrubConfig(menu bootVars) error {
	log.Debugf("pvGrubConfig")

	filetmpl, err := template.New("grublst").Parse(`{{define "grubmenu"}}
title RancherOS {{.Version}}-({{.Name}})
root (hd0)
kernel /${bootDir}vmlinuz-{{.Version}}-rancheros {{.KernelArgs}} {{.Append}}
initrd /${bootDir}initrd-{{.Version}}-rancheros

{{end}}
default 0
timeout {{.Timeout}}
{{if .Fallback}}fallback {{.Fallback}}{{end}}
hiddenmenu

{{- range .Entries}}
{{template "grubmenu" .}}
{{- end}}

`)
	if err != nil {
		log.Errorf("pv grublst: %s", err)

		return err
	}

	cfgFile := filepath.Join(menu.baseName, menu.bootDir+"grub/menu.lst")
	log.Debugf("grubMenu written to %s", cfgFile)
	f, err := os.Create(cfgFile)
	if err != nil {
		log.Errorf("Create(%s) %s", cfgFile, err)

		return err
	}
	err = filetmpl.Execute(f, menu)
	if err != nil {
		log.Errorf("execute %s", err)
		return err
	}
	return nil
}
