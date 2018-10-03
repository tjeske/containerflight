#! /bin/bash

go build
go install

../../../../bin/containerflight run ./examples/ansible ansible-playbook --version
../../../../bin/containerflight run ./examples/python3.7 --version
../../../../bin/containerflight run ./examples/youtube-dl --version
../../../../bin/containerflight run ./examples/scrapy --version
../../../../bin/containerflight run ./examples/gitg
../../../../bin/containerflight run ./examples/hugin
../../../../bin/containerflight run ./examples/eclipse
