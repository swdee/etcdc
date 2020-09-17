module github.com/swdee/etcdc

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200824191128-ae9734ed278b

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/prometheus/client_golang v1.7.1
	go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd v0.0.0-20200824191128-ae9734ed278b
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.29.1
	sigs.k8s.io/yaml v1.2.0
)
