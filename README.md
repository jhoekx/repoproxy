# Repoproxy

Repoproxy creates a temporary lazy mirror of a CentOS mirror.

This is useful to provision multiple CentOS systems without having to download all rpms and other resources again and again. It is suited for home labs.

This is not a real mirroring utility.
Use [mrepo](https://github.com/dagwieers/mrepo) or `reposync` instead.

Metadata will always be fetched from the remote repository.
RPMs are considered immutable and won't be downloaded again once they have been requested.
PXE files or live images are cached for one run.
Restart when newer versions are available.

## Building

For running from checkout, use:

```
$ go run repoproxy.go
```

To create a docker container, use:

```
$ docker build -t repoproxy .
```

## Running

`repoproxy` can be configured using flags or environment variables.
Environment variables are preferred over flags.

`CENTOS_MIRROR` or `--mirror` configure the CentOS mirror to proxy.

`RPM_DIR` or `--rpmdir` configure the directory to store mirrored RPMS.

`repoproxy` listens on port 8080.

### Docker

Start the docker container with:

```
$ docker run --rm -v rpm_mirror:/var/lib/repoproxy/rpms -p 8080:8080 -it repoproxy:latest
```

This creates a docker volume to store the RPMs.
