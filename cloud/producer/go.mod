module github.com/ergo-services/examples/cloud/producer

go 1.17

require (
	github.com/ergo-services/ergo v1.999.220
	github.com/ergo-services/examples/cloud/consumer v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.0
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/ergo-services/examples/cloud/consumer => ../consumer

replace github.com/ergo-services/ergo => ../../../ergo
