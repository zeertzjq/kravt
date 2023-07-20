package main

import (
	"flag"
	"fmt"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func tryCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	fmt.Fprintf(os.Stderr, "Running: %s\n", cmd)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func handleDefine(args []string) {
	cmd := flag.NewFlagSet(fmt.Sprintf("%s define", os.Args[0]), flag.ExitOnError)
	domainNamePtr := cmd.String("domain", "", "domain name")
	kernelPathPtr := cmd.String("kernel", "", "path to kernel image")
	rootfsPathPtr := cmd.String("rootfs", "", "path to root filesystem")
	mountTagPtr := cmd.String("rootfs-tag", "fs0", "tag name to mount filesystem")
	bridgePtr := cmd.Bool("bridge", false, "bridge network to guest")
	bridgeNamePtr := cmd.String("bridge-name", "virbr0", "name of bridge device")
	bridgeGuestPtr := cmd.String("bridge-guest", "172.44.0.2", "guest IPv4 address")
	bridgeGatewayPtr := cmd.String("bridge-gateway", "172.44.0.1", "gateway IPv4 address")
	bridgeNetmaskPtr := cmd.String("bridge-netmask", "255.255.255.0", "bridge IPv4 subnet mask")
	memoryPtr := cmd.Uint("memory", 32, "assign MiB memory to guest")
	startPtr := cmd.Bool("start", false, "start the domain")
	cmd.Parse(args)
	if *domainNamePtr == "" {
		fmt.Fprintln(os.Stderr, "missing domain name")
		cmd.Usage()
		return
	}
	if *kernelPathPtr == "" {
		fmt.Fprintln(os.Stderr, "missing kernel image path")
		cmd.Usage()
		return
	}
	kernelPath, err := filepath.Abs(*kernelPathPtr)
	if err != nil {
		panic(err)
	}
	rootfsPath := ""
	if *rootfsPathPtr != "" {
		rootfsPath, err = filepath.Abs(*rootfsPathPtr)
		if err != nil {
			panic(err)
		}
	}
	cmdlineArgs := make([]string, 0)
	bridgeGuest := net.ParseIP(*bridgeGuestPtr).To4()
	bridgeGateway := net.ParseIP(*bridgeGatewayPtr).To4()
	bridgeNetmask := net.ParseIP(*bridgeNetmaskPtr).To4()
	if *bridgePtr {
		if bridgeGuest == nil {
			panic("invalid guest IPv4 address")
		}
		if bridgeGateway == nil {
			panic("invalid gateway IPv4 address")
		}
		if bridgeNetmask == nil {
			panic("invalid IPv4 subnet mask")
		}
		for _, arg := range []string{
			fmt.Sprintf("netdev.ipv4_addr=%s", bridgeGuest.String()),
			fmt.Sprintf("netdev.ipv4_gw_addr=%s", bridgeGateway.String()),
			fmt.Sprintf("netdev.ipv4_subnet_mask=%s", bridgeNetmask.String()),
		} {
			cmdlineArgs = append(cmdlineArgs, arg)
		}
	}
	cmdlineArgs = append(cmdlineArgs, "--")
	for _, arg := range cmd.Args() {
		cmdlineArgs = append(cmdlineArgs, arg)
	}
	cmdline := strings.Join(cmdlineArgs, " ")
	domcfg := &libvirtxml.Domain{
		Type: "kvm",
		Name: *domainNamePtr,
		Memory: &libvirtxml.DomainMemory{
			Value: *memoryPtr,
			Unit:  "MiB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     1,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
			Kernel:  kernelPath,
			Cmdline: cmdline,
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-model",
		},
		Clock: &libvirtxml.DomainClock{
			Offset:     "utc",
			Adjustment: "reset",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "preserve",
		Devices: &libvirtxml.DomainDeviceList{
			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						Port:   -1,
						Listen: "127.0.0.1",
					},
				},
			},
			MemBalloon: &libvirtxml.DomainMemBalloon{
				Model: "none",
			},
		},
	}
	if rootfsPath != "" {
		domcfg.Devices.Filesystems = []libvirtxml.DomainFilesystem{
			{
				Driver: &libvirtxml.DomainFilesystemDriver{
					Type: "path",
				},
				Source: &libvirtxml.DomainFilesystemSource{
					Mount: &libvirtxml.DomainFilesystemSourceMount{
						Dir: rootfsPath,
					},
				},
				Target: &libvirtxml.DomainFilesystemTarget{
					Dir: *mountTagPtr,
				},
			},
		}
	}
	if *bridgePtr {
		bridgePrefixSize, _ := net.IPMask(bridgeNetmask).Size()
		bridgeCIDR := fmt.Sprintf("%s/%d", bridgeGateway.String(), bridgePrefixSize)
		tryCommand("sudo", "ip", "link", "add", "dev", *bridgeNamePtr, "type", "bridge")
		tryCommand("sudo", "ip", "address", "add", bridgeCIDR, "dev", *bridgeNamePtr)
		tryCommand("sudo", "ip", "link", "set", "dev", *bridgeNamePtr, "up")
		domcfg.Devices.Interfaces = []libvirtxml.DomainInterface{
			{
				Source: &libvirtxml.DomainInterfaceSource{
					Bridge: &libvirtxml.DomainInterfaceSourceBridge{
						Bridge: *bridgeNamePtr,
					},
				},
				Model: &libvirtxml.DomainInterfaceModel{
					Type: "virtio",
				},
			},
		}
	}
	xmldoc, err := domcfg.Marshal()
	if err != nil {
		panic(err)
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	dom, err := conn.DomainDefineXML(xmldoc)
	if err != nil {
		panic(err)
	}
	if *startPtr {
		err = dom.Create()
		if err != nil {
			panic(err)
		}
	}
}

func handleStart(args []string) {
	cmd := flag.NewFlagSet(fmt.Sprintf("%s start", os.Args[0]), flag.ExitOnError)
	domainNamePtr := cmd.String("domain", "", "domain name")
	cmd.Parse(args)
	if len(cmd.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		cmd.Usage()
		return
	}
	if *domainNamePtr == "" {
		fmt.Fprintln(os.Stderr, "missing domain name")
		cmd.Usage()
		return
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	dom, err := conn.LookupDomainByName(*domainNamePtr)
	if err != nil {
		panic(err)
	}
	err = dom.Create()
	if err != nil {
		panic(err)
	}
}

func handleInfo(args []string) {
	cmd := flag.NewFlagSet(fmt.Sprintf("%s info", os.Args[0]), flag.ExitOnError)
	domainNamePtr := cmd.String("domain", "", "domain name")
	cmd.Parse(args)
	if len(cmd.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		cmd.Usage()
		return
	}
	if *domainNamePtr == "" {
		fmt.Fprintln(os.Stderr, "missing domain name")
		cmd.Usage()
		return
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	dom, err := conn.LookupDomainByName(*domainNamePtr)
	if err != nil {
		panic(err)
	}
	info, err := dom.GetInfo()
	if err != nil {
		panic(err)
	}
	switch info.State {
	case libvirt.DOMAIN_NOSTATE:
		fmt.Println("State: no state")
	case libvirt.DOMAIN_RUNNING:
		fmt.Println("State: running")
	case libvirt.DOMAIN_BLOCKED:
		fmt.Println("State: blocked on resource")
	case libvirt.DOMAIN_PAUSED:
		fmt.Println("State: paused")
	case libvirt.DOMAIN_SHUTDOWN:
		fmt.Println("State: being shut down")
	case libvirt.DOMAIN_SHUTOFF:
		fmt.Println("State: shut off")
	case libvirt.DOMAIN_CRASHED:
		fmt.Println("State: crashed")
	case libvirt.DOMAIN_PMSUSPENDED:
		fmt.Println("State: suspended by guest power management")
	default:
		fmt.Println("State: unknown")
	}
}

func handleDestroy(args []string) {
	cmd := flag.NewFlagSet(fmt.Sprintf("%s destroy", os.Args[0]), flag.ExitOnError)
	domainNamePtr := cmd.String("domain", "", "domain name")
	cmd.Parse(args)
	if len(cmd.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		cmd.Usage()
		return
	}
	if *domainNamePtr == "" {
		fmt.Fprintln(os.Stderr, "missing domain name")
		cmd.Usage()
		return
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	dom, err := conn.LookupDomainByName(*domainNamePtr)
	if err != nil {
		panic(err)
	}
	err = dom.Destroy()
	if err != nil {
		panic(err)
	}
}

func handleUndefine(args []string) {
	cmd := flag.NewFlagSet(fmt.Sprintf("%s undefine", os.Args[0]), flag.ExitOnError)
	domainNamePtr := cmd.String("domain", "", "domain name")
	destroyPtr := cmd.Bool("destroy", false, "destroy the domain")
	cmd.Parse(args)
	if len(cmd.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		cmd.Usage()
		return
	}
	if *domainNamePtr == "" {
		fmt.Fprintln(os.Stderr, "missing domain name")
		cmd.Usage()
		return
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	dom, err := conn.LookupDomainByName(*domainNamePtr)
	if err != nil {
		panic(err)
	}
	interfaces := make([]libvirtxml.DomainInterface, 0)
	xmldoc, err := dom.GetXMLDesc(0)
	if err == nil {
		domcfg := &libvirtxml.Domain{}
		err = domcfg.Unmarshal(xmldoc)
		if err == nil {
			interfaces = domcfg.Devices.Interfaces
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	if *destroyPtr {
		err = dom.Destroy()
		if err != nil {
			panic(err)
		}
	}
	err = dom.Undefine()
	if err != nil {
		panic(err)
	}
	for _, iface := range interfaces {
		if iface.Source == nil || iface.Source.Bridge == nil {
			continue
		}
		bridgeName := iface.Source.Bridge.Bridge
		tryCommand("sudo", "ip", "link", "set", "dev", bridgeName, "down")
		tryCommand("sudo", "ip", "link", "del", "dev", bridgeName)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  define")
	fmt.Fprintln(os.Stderr, "  start")
	fmt.Fprintln(os.Stderr, "  info")
	fmt.Fprintln(os.Stderr, "  destroy")
	fmt.Fprintln(os.Stderr, "  undefine")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	switch os.Args[1] {
	case "define":
		handleDefine(os.Args[2:])
	case "start":
		handleStart(os.Args[2:])
	case "info":
		handleInfo(os.Args[2:])
	case "destroy":
		handleDestroy(os.Args[2:])
	case "undefine":
		handleUndefine(os.Args[2:])
	default:
		printUsage()
		return
	}
}
