package main

import (
	"flag"
	"fmt"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
	"path/filepath"
	"strings"
)

func main() {
	domainNamePtr := flag.String("domain", "", "domain name")
	kernelPathPtr := flag.String("kernel", "", "path to kernel image")
	rootfsPathPtr := flag.String("rootfs", "", "path to root filesystem")
	mountTagPtr := flag.String("rootfs-tag", "fs0", "tag name to mount filesystem")
	bridgePtr := flag.Bool("bridge", false, "bridge network to guest")
	bridgeNamePtr := flag.String("bridge-name", "virbr0", "name of bridge device")
	memoryPtr := flag.Uint("memory", 32, "assign MiB memory to guest")
	flag.Parse()
	if *domainNamePtr == "" {
		panic("missing domain name")
	}
	if *kernelPathPtr == "" {
		panic("missing kernel image path")
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
	if *bridgePtr {
		cmdlineArgs = append(cmdlineArgs, "netdev.ipv4_addr=172.44.0.2")
		cmdlineArgs = append(cmdlineArgs, "netdev.ipv4_gw_addr=172.44.0.1")
		cmdlineArgs = append(cmdlineArgs, "netdev.ipv4_subnet_mask=255.255.255.0")
	}
	cmdlineArgs = append(cmdlineArgs, "--")
	for _, arg := range flag.Args() {
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
	fmt.Scanln()
	dom.Undefine()
}
