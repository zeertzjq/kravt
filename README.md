# Kravt

A utility to run and manage [Unikraft](https://github.com/unikraft/unikraft) unikernels using [libvirt](https://libvirt.org/).

## Building and Installation

This is a [Go](https://go.dev/) package, but a Makefile is also provided.

```sh
git clone https://github.com/zeertzjq/kravt.git
cd kravt
make && sudo make install
```

Installation prefix defaults to `/usr/local`. You can change it by setting the `PREFIX` variable:

```sh
make && sudo make PREFIX=/opt install
```

## Examples

### Example using [app-nginx](https://github.com/unikraft/app-nginx)

To define a libvirt domain for the unikernel without launching a virtual machine:

```sh
kravt define -kernel build/nginx_qemu-x86_64 -rootfs fs0/ -bridge
```

To define a libvirt domain for the unikernel and launch a virtual machine:

```sh
kravt define -start -kernel build/nginx_qemu-x86_64 -rootfs fs0/ -bridge
```

To launch a virtual machine when none is running:

```sh
kravt start
```

To print information about the virtual machine:

```sh
kravt info
```

To shut down the running virtual machine:

```sh
kravt destroy
```

To undefine the libvirt domain without shutting down the virtual machine:

```sh
kravt undefine
```

To undefine the libvirt domain and shut down the virtual machine:

```sh
kravt undefine -destroy
```

### Example using [app-elfloader](https://github.com/unikraft/app-elfloader) and [nginx](https://github.com/unikraft/dynamic-apps/tree/master/nginx)

To define a libvirt domain for the unikernel without launching a virtual machine:

```sh
kravt define -kernel build/elfloader_qemu-x86_64 -memory 2048 -rootfs ../dynamic-apps/nginx/ -bridge -- /usr/local/nginx/sbin/nginx -c /usr/local/nginx/conf/nginx.conf
```

To define a libvirt domain for the unikernel and launch a virtual machine:

```sh
kravt define -start -kernel build/elfloader_qemu-x86_64 -memory 2048 -rootfs ../dynamic-apps/nginx/ -bridge -- /usr/local/nginx/sbin/nginx -c /usr/local/nginx/conf/nginx.conf
```

To launch a virtual machine when none is running:

```sh
kravt start
```

To print information about the virtual machine:

```sh
kravt info
```

To shut down the running virtual machine:

```sh
kravt destroy
```

To undefine the libvirt domain without shutting down the virtual machine:

```sh
kravt undefine
```

To undefine the libvirt domain and shut down the virtual machine:

```sh
kravt undefine -destroy
```
