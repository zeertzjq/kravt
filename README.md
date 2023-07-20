# Kravt

An utility to run and manage [Unikraft](https://github.com/unikraft/unikraft) unikernels using [libvirt](https://libvirt.org/).

## Examples

### Example with [app-nginx](https://github.com/unikraft/app-nginx)

To define a libvirt domain for the unikernel without launching a virtual machine:

```sh
kravt define -domain app-nginx -kernel build/nginx_qemu-x86_64 -rootfs fs0/
```

To define a libvirt domain for the unikernel and launch a virtual machine:

```sh
kravt define -domain app-nginx -start -kernel build/nginx_qemu-x86_64 -rootfs fs0/
```

To launch a virtual machine when none is running:

```sh
kravt start -domain app-nginx
```

To print information about the virtual machine:

```sh
kravt info -domain app-nginx
```

To shut down the running virtual machine:

```sh
kravt destroy -domain app-nginx
```

To undefine the libvirt domain without shutting down the virtual machine:

```sh
kravt undefine -domain app-nginx
```

To undefine the libvirt domain and shut down the virtual machine:

```sh
kravt undefine -domain app-nginx -destroy
```

### Example with [app-elfloader](https://github.com/unikraft/app-elfloader) and [nginx](https://github.com/unikraft/dynamic-apps/tree/master/nginx)

To define a libvirt domain for the unikernel without launching a virtual machine:

```sh
kravt define -domain app-elfloader -kernel build/elfloader_qemu-x86_64 -memory 2048 -bridge -rootfs ../dynamic-apps/nginx/ -- /usr/local/nginx/sbin/nginx -c /usr/local/nginx/conf/nginx.conf
```

To define a libvirt domain for the unikernel and launch a virtual machine:

```sh
kravt define -domain app-elfloader -start -kernel build/elfloader_qemu-x86_64 -memory 2048 -bridge -rootfs ../dynamic-apps/nginx/ -- /usr/local/nginx/sbin/nginx -c /usr/local/nginx/conf/nginx.conf
```

To launch a virtual machine when none is running:

```sh
kravt start -domain app-elfloader
```

To print information about the virtual machine:

```sh
kravt info -domain app-elfloader
```

To shut down the running virtual machine:

```sh
kravt destroy -domain app-elfloader
```

To undefine the libvirt domain without shutting down the virtual machine:

```sh
kravt undefine -domain app-elfloader
```

To undefine the libvirt domain and shut down the virtual machine:

```sh
kravt undefine -domain app-elfloader -destroy
```
