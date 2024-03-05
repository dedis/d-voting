module github.com/c4dt/d-voting

go 1.20

require (
	github.com/gorilla/mux v1.8.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.17.0
	github.com/rs/zerolog v1.28.0
	github.com/stretchr/testify v1.8.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	go.dedis.ch/dela v0.0.0-20231004135936-647c76e51d8a
	go.dedis.ch/dela-apps v0.0.0-20230929051236-6d89286321f7
	go.dedis.ch/kyber/v3 v3.1.0
	golang.org/x/net v0.10.0
	golang.org/x/tools v0.6.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

replace (
	go.dedis.ch/dela => github.com/c4dt/dela v0.0.0-20240305154115-1db49b37b0a3
	go.dedis.ch/dela-apps => github.com/c4dt/dela-apps v0.0.0-20231121155105-f3a8a6f4b3b8
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20200813121455-4a6760c71486 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230919144854-baaa0384eac5 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/urfave/cli/v2 v2.2.0 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.dedis.ch/protobuf v1.0.11 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
