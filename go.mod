module p9t.io/skafos

go 1.17

require (
	github.com/golang/glog v1.0.0
	github.com/spf13/cobra v1.4.0
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	p9t.io/kuberboat v0.0.1
)

require (
	github.com/creasty/defaults v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
)

replace p9t.io/kuberboat => ./kuberboat
