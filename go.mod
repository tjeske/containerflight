module github.com/tjeske/containerflight

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Microsoft/go-winio v0.4.5
	github.com/Nvveen/Gotty v0.0.0-20170406111628-a8b993ba6abd
	github.com/agl/ed25519 v0.0.0-20140907235247-d2b94fd789ea
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/continuity v0.0.0-20170913164642-35d55c5e8dd2
	github.com/davecgh/go-spew v1.1.0
	github.com/docker/cli v17.12.1-ce+incompatible
	github.com/docker/distribution v2.6.0-rc.1.0.20170726174610-edc3ab29cdff+incompatible
	github.com/docker/docker v1.4.2-0.20171207004338-a1be987ea9e0
	github.com/docker/docker-credential-helpers v0.5.3-0.20170816090621-3c90bd29a46b
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.3.2-0.20170127094116-9e638d38cf69
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gogo/protobuf v0.0.0-20170307180453-100ba4e88506
	github.com/golang/protobuf v0.0.0-20170523065751-7a211bcf3bce
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f
	github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/mattn/go-shellwords v1.0.3
	github.com/miekg/pkcs11 v0.0.0-20160222192025-df8ae6ca7304
	github.com/mitchellh/mapstructure v0.0.0-20161020161836-f3009df150da
	github.com/moby/buildkit v0.0.0-20170922161955-aaff9d591ef1
	github.com/opencontainers/go-digest v0.0.0-20170111181659-21dfd564fd89
	github.com/opencontainers/image-spec v1.0.0
	github.com/opencontainers/runc v1.0.0-rc4.0.20171108154827-b2567b37d7b7
	github.com/pkg/errors v0.8.1-0.20161002052512-839d9e913e06
	github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus v1.0.3
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.2-0.20171119075854-34ceca591bcf
	github.com/spf13/pflag v1.0.1-0.20171020110617-97afa5e7ca8a
	github.com/stretchr/testify v1.1.5-0.20170130113145-4d4bfba8f1d1
	github.com/theupdateframework/notary v0.5.2-0.20171026220044-05985dc5d1c7
	github.com/tonistiigi/fsutil v0.0.0-20170929161712-dea3a0da73ae
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c
	github.com/xeipuuv/gojsonschema v0.0.0-20160323030313-93e72a773fad
	golang.org/x/crypto v0.0.0-20170728183002-558b6879de74
	golang.org/x/net v0.0.0-20170525011637-7dcfb8076726
	golang.org/x/sync v0.0.0-20161206014632-450f422ab23c
	golang.org/x/sys v0.0.0-20171031081856-95c657629925
	golang.org/x/text v0.3.0
	golang.org/x/time v0.0.0-20160202183820-a4bde1265759
	google.golang.org/genproto v0.0.0-20170523043604-d80a6e20e776
	google.golang.org/grpc v1.3.0
	gopkg.in/yaml.v2 v2.0.0-20170125143719-4c78c975fe7c
)

replace github.com/Nvveen/Gotty => github.com/ijc25/Gotty v0.0.0-20170406111628-a8b993ba6abd

replace github.com/docker/cli => github.com/tjeske/containerflight-docker-cli v17.12.1-ce+incompatible
