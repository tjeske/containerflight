[![Build Status](https://travis-ci.org/tjeske/containerflight.svg)](https://travis-ci.org/tjeske/containerflight)
[![Go Report Card](https://goreportcard.com/badge/github.com/tjeske/containerflight)](https://goreportcard.com/report/github.com/tjeske/containerflight)
[![Coverage Status](https://codecov.io/gh/tjeske/containerflight/branch/master/graph/badge.svg)](https://codecov.io/gh/tjeske/containerflight/branch/master)

# What is containerflight?

Containerflight allows you to run arbitrary Linux applications, build scripts etc. in a **defined** and **isolated** environment in which the current user context is considered. A containerflight app is started like any other program on your computer. It is basically a single yaml file which describes the application, its environment and dependencies. Containerflight has zero dependencies. It only needs access to the [Docker](https://www.docker.com/) daemon.

Example for an app file to run [scrapy](https://scrapy.org/) (a high-level web crawling and web scraping framework) without installing or upgrading packages on your host system. If you store the following description on disk and make it executable, you can run scrapy as it has been installed on the host system (just enter `./scrapy` ⏎).

```yaml
#!/usr/local/bin/containerflight run

image:
    base: docker://python:3.6.0
    dockerfile: |
        # install gcc
        RUN apt-get update && \
            export DEBIAN_FRONTEND=noninteractive && \
            apt-get install -y gcc && \
            rm -rf /var/lib/apt/lists/*****

        # install scrapy==1.5.1
        RUN pip install scrapy

        ENTRYPOINT [ "scrapy" ]
```

# Getting started

1. Download the containerflight binary from [github](https://github.com/tjeske/containerflight/releases/latest) (or build it from source). It is recommended to store it in your `$PATH` so that you can directly run containerflight apps. Don't forget to set the executable flag for the containerflight executable. The following snippet installs the latest stable release of containerflight into `/usr/local/bin/`.

    ```bash
    sudo curl https://github.com/tjeske/containerflight/releases/download/0.1/containerflight -o /usr/local/bin/containerflight
    sudo chmod +x /usr/local/bin/containerflight
    ```

2. Write an app yaml file for your application. At the moment, only Docker as the container runtime is supported.

## App file

An application file (or app file) is a [yaml](https://en.wikipedia.org/wiki/YAML) file describing the environment and dependencies needed to run the application. The following app file shows all available parameters. You find examples of some popular applications in the `examples/` folder.

```yaml
#!/usr/local/bin/containerflight run
compatibility: 0.2.x
name: MyApplication
version: 1.0
description: This is just an example.

console: false
gui: true

image:
    base: "docker://ubuntu:18.04"
    dockerfile: |
        RUN sh myscript.sh

runtime:
    docker:
        runargs: [
            "-v", "${HOME}:${HOME}"
        ]
```

If the first line of an app file is `#!<containerflight executable> run` and if the executable flag is set, you can directly run the containerflight application like any other program.

Parameters like `name`, `version` and `description` can be used to describe the application. If Docker is used, these parameters are considered when assigning an image name etc.

The `gui` parameter must be set to `true` to give the application access to the host X server. Set the `console` parameter to `false` (default is `true`) when a TTY should not be allocated and stdin is kept closed.

## Image

```yaml
image:
    base: docker://<docker image>
    dockerfile: |
        ...
```

Containerflight has an integrated Docker client which supports API 1.25 (implemented by Docker 1.13) and can directly talk to the Docker daemon. This makes it easier to run apps on a CI build-slave like Jenkins.

A Docker image serves as a basis (`base`) and can be extended by using the Dockerfile syntax. See [https://docs.docker.com/engine/reference/builder/](https://docs.docker.com/engine/reference/builder/) for more information.

## Runtime

```yaml
runtime:
    docker:
        runargs: [
            ...
        ]
```

At the moment Docker is used as the container runtime. You can specify additional `docker run ...` arguments as a yaml array via `runargs: [ .. ]`. Type `docker run --help` for more information.

## Compatibility

An app file can be linked to a specific containerflight version.

```yaml
compatibility: "0.1.2"
```

enforces containerflight runtime version 0.1.2.

Containerflight versions follow the [Semantic Versioning](https://semver.org/) system, where a version string consists of MAJOR.MINOR.PATCH.

- MAJOR version increase makes incompatible changes
- MINOR version increase adds functionality in a backwards-compatible manner
- PATCH version increase makes backwards-compatible bug fixes

Major version zero (0.y.z) is for initial development. Anything may change at any time. Upcoming versions >= 1.0.0 of containerflight are always compatible until the MAJOR version won't be changed. If you are using containerflight 1.1.0 you can express compatibility with higher version of the containerflight runtime by

```yaml
compatibility: ">=1.1.0 <2.0.0"
```

Please have a look at [https://github.com/blang/semver](https://github.com/blang/semver) to describe more complex version ranges.

## Parameters

You can use parameters in an app file to consider the current user context.

- `${APP_FILE_DIR}`: directory where the app file is located
- `${USERNAME}`: current user name
- `${USERID}`: ID of the current user
- `${GROUPNAME}`: primary group of the current user
- `${GROUPID}`: primary group ID of the current user
- `${HOME}`: current user's home directory
- `${PWD}`: current working directory
- `${ENV:<envname>}`: value of an environment variable (e.g. `${ENV:http_proxy}`)

# Why containerflight?

Container technology like Docker is great but is not primarily made for (desktop) applications. Applications need context.

A containerized application runs under its own user id but often needs to process external user data. If the container adds new files and stores them externally, the files are owned by the container user and not by the external user.

With containerflight you can easily describe containers which consider their current environment. Processes in the container get the user id of the external user id. Environment variables are evaluated during container creation e.g. to respect proxy settings etc.

## Link tools with your source code

I have often faced the problem that tools like compilers, development IDEs etc. should be part of a source code commit. New team members should start as soon as possible without wasting time while setting-up the (development) environment or running into troubles by using unsupported tools.

Tools like Vagrant and a bunch of bootstrap scripts can solve this problem. The problem is that a whole virtual machine for development is often too heavyweight (especially if you are working on several projects in parallel).

With containerflight you can describe each tool as a single yaml file and put them in your source code repository. Running and switching between such programs becomes a no-brainer.

## Use and test new software

Using the latest version of a software can become a challenge. In general I prefer a stable Linux base system (LTS version of Ubuntu etc.) over a more unstable system with bleeding-edge software but sometimes I need a newer version of an application. The problems start to begin if the applications depend on a newer version of the system base libraries (e.g. GTK). Upgrading such libraries means upgrading the whole distribution, which is not an option.

Containerflight makes it easy to run such applications with its own library dependencies separately from the base system.

# Build

1. Install Go >= 1.9
2. Get dep: `go get -u github.com/golang/dep/cmd/dep`
3. Install dependencies: `dep ensure -v`
4. Build: `go build -ldflags '-s'`

# Contributing

1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/containerflight && cd containerflight`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

# License

Containerflight is released under the Apache 2.0 license. See [LICENSE](https://github.com/tjeske/containerflight/blob/master/LICENSE.txt)
