package main

import (
	"flag"

	"github.com/ergo-services/ergo"
	"github.com/ergo-services/ergo/gen"
	"github.com/ergo-services/ergo/node"
	log "github.com/sirupsen/logrus"
)

var (
	OptionNodeName   string
	OptionNodeCookie string

	OptionCloudClusterName   string
	OptionCloudClusterCookie string

	OptionTo string
)

func init() {

	flag.StringVar(&OptionNodeName, "name", "producer@localhost", "node name")
	flag.StringVar(&OptionNodeCookie, "cookie", "secret", "cookie for interaction with local cluster")

	flag.StringVar(&OptionCloudClusterName, "cloud-cluster", "", "cluster name")
	flag.StringVar(&OptionCloudClusterCookie, "cloud-cookie", "", "cookie for interaction with cloud cluster")

	flag.StringVar(&OptionTo, "send-message-to", "consumer", "send message to")

}

func main() {
	flag.Parse()

	if OptionCloudClusterName == "" {
		panic("option -cloud-cluster can not be empty")
	}

	if OptionCloudClusterCookie == "" {
		panic("option -cloud-cookie can not be empty")
	}

	nodeOptions := node.Options{}

	// Enable cloud feature.
	nodeOptions.Cloud.Enable = true

	// Set your cluster name and cookie to get access to the cloud
	nodeOptions.Cloud.Cluster = OptionCloudClusterName
	nodeOptions.Cloud.Cookie = OptionCloudClusterCookie

	// We should enable accepting incoming connection requests
	// from the nodes in your cloud cluster.
	nodeOptions.Proxy.Accept = true

	// Start new node
	thisNode, err := ergo.StartNode(OptionNodeName, OptionNodeCookie, nodeOptions)
	if err != nil {
		panic(err)
	}
	log.Infof("node %q succesfully started", OptionNodeName)

	// Spawn new process on this node
	if _, err := thisNode.Spawn("demo", gen.ProcessOptions{}, &producerServer{}); err != nil {
		panic(err)
	}

	thisNode.Wait()
}
