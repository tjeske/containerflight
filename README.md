[![Build Status](https://travis-ci.org/tjeske/containerflight.svg)](https://travis-ci.org/tjeske/containerflight)
[![Go Report Card](https://goreportcard.com/badge/github.com/tjeske/containerflight)](https://goreportcard.com/report/github.com/tjeske/containerflight)

# What is Container Flight?

Container Flight allows you to run arbitrary Linux applications in a containerized, **isolated** environment in which the current user context is considered. The application is started like any other program on your computer by just typing its name on the console and pressing enter. A single yaml file is enough to **describe** the application and its environment, dependencies etc. Container Flight is based on [Docker](https://www.docker.com/).

*Container Flight is EXPERIMENTAL and API changes are possible!*

# Why Container Flight?

Container technology like Docker is great but is not primarily made for (desktop) applications. Applications need context.

A containerized application runs under its own user id but often needs to process external user data. If the container adds new files and stores them externally, the files are owned by the container user and not by the external user. 

With Container Flight you can easily describe containers which consider their current environment. Processes in the container get the user id of the external user id. Environment variables are evaluated during container creation e.g. to respect proxy settings etc.

## Link tools with your source code

I have often faced the problem that tools like compilers, development IDEs etc. should be part of a source code commit. New team members should start as soon as possible without wasting time while setting-up the (development) environment or running into troubles by using unsupported tools.

Tools like Vagrant and a bunch of bootstrap scripts can solve this problem. The problem is that a whole virtual machine for development is often too heavyweight (especially if you are working on several projects in parallel).

With Container Flight you can describe each tool as a single yaml file and put them in your source code repository. Running and switching between such programs becomes a no-brainer.

## Use and test new software

Using the latest version of a software can become a challenge. In general I prefer a stable Linux base system (LTS version of Ubuntu etc.) over a more unstable system with bleeding-edge software but sometimes I need a newer version of an application. The problems start to begin if the applications depend on a newer version of the system base libraries (e.g. GTK). Upgrading such libraries means upgrading the whole distribution, which is not an option.

Container Flight makes it easy to run such applications with its own library dependencies separately from the base system.


# Getting started

First, you have to download the Container Flight binary from github (or build it from source). It is recommended to store it somewhere where your `$PATH` environment variable points to so that you can directly run Container Flight apps. Don't forget to set the executable flag for the Container Flight executable.

Second, write an app yaml file for your application. At the moment, only Docker as the container runtime is supported.

Example for an app file:
```
#!/usr/local/bin/containerflight run  # <- this should point to the Container Flight executable
compatibility: ">=0.1.0 <1.0.0"

docker:
    dockerfile: |
        FROM python:3.7.0b2-alpine

        ENV http_proxy=${ENV:http_proxy}
        ENV https_proxy=${ENV:https_proxy}

        # set-up the host user/group environment 
        RUN set -x ; \
            addgroup -g ${GROUPID} -S ${GROUPNAME} ; \
            adduser -u ${USERID} -D -S -G ${GROUPNAME} ${USERNAME} && exit 0 ; exit 1

        USER ${USERNAME}

        ENTRYPOINT [ "python" ]

    runargs: [
        "-ti",
        "--rm",
        "-a", "stdin",
        "-a", "stdout",
        "-v", "${PWD}:${PWD}",
        "-w", "${PWD}"
    ] 
```
You find some examples for console and gui apps in the `examples/` folder.

## Parameters

You can use parameters in an app file to consider the current user context.

- `${USERNAME}`: current user name
- `${USERID}`: ID of the current user
- `${GROUPNAME}`: primary group of the current user
- `${GROUPID}`: primary group ID of the current user
- `${HOME}`: current user's home directory
- `${PWD}`: current working directory
- `${ENV:<envname>}`: value of an environment variable (e.g. `${ENV:http_proxy}`)

## Docker

```
docker:
    dockerfile: |
        ...
    runargs: [ ... ] 
```

Container Flight has an integrated Docker client which supports API 1.25 (implemented by Docker 1.13) and can directly talk to the Docker daemon. This makes it easier to run apps on a CI build-slave like Jenkins.

In an app file you can use `docker: dockerfile:` to describe the app container as you would describe a Docker image. See [https://docs.docker.com/engine/reference/builder/](https://docs.docker.com/engine/reference/builder/) for more information.

You can specify additional `docker run ...` arguments as a yaml array via `docker: runargs: [ ... ]`. Type `docker run --help` for more information.

## Versioning

An app file can be linked to a specific Container Flight version.
```
version: "0.1.2"
```
enforces Container Flight runtime version 0.1.2.

Container Flight versions follow the [Semantic Versioning](https://semver.org/) system, where a version string consists of MAJOR.MINOR.PATCH.
- MAJOR version increase makes incompatible changes
- MINOR version increase adds functionality in a backwards-compatible manner
- PATCH version increase makes backwards-compatible bug fixes

Major version zero (0.y.z) is for initial development. Anything may change at any time. Upcoming versions >= 1.0.0 of Container Flight are always compatible until the MAJOR version won't be changed. If you are using Container Flight 1.1.0 you can express compatibility with higher version of the Container Flight runtime by 
```
compatibility: ">=1.1.0 <2.0.0"
```
Please have a look at [https://github.com/blang/semver](https://github.com/blang/semver) to describe more complex version ranges.


# Build

1. Install Go >= 1.9
2. Get dep: `go get -u github.com/golang/dep/cmd/dep`
3. Install dependencies: `dep ensure -v`
3. Build: `go build -ldflags '-s'`


# Contributing

1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/Container Flight && cd Container Flight`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request


# License

Container Flight is released under the Apache 2.0 license. See [LICENSE](https://github.com/tjeske/containerflight/blob/master/LICENSE.txt)
