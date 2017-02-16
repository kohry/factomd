package engine

// All plugins we can intiate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-plugin"
)

// How often to check the buffer
var CHECK_BUFFER time.Duration = 2 * time.Second

var _ log.Logger
var _ = ioutil.Discard

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	// Plugin to manage dbstates
	"manager": &IManagerPlugin{},
}

func LaunchConsulPlugin() (*consulapi.Client, string) {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		panic(err)
	}
	session := consul.Session()
	sessionID, _, err := session.Create(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Consul Session ID:", sessionID)
	kv := consul.KV()
	kvPairList, _, err := kv.List("", nil)
	if err == nil && kvPairList != nil {
		fmt.Println("Full Consul List:")
		for _, kvPair := range kvPairList {
			fmt.Println(kvPair.Key, ":", string(kvPair.Value))
		}
	}

	return consul, sessionID
}

// LaunchDBStateManagePlugin launches the plugin and returns an interface that
// can be interacted with like a usual interface. The client returned must be
// killed before we exit
func LaunchTorrentDBStateManagePlugin(path string, inQueue chan interfaces.IMsg, sigKey *primitives.PrivateKey) (interfaces.IManagerController, error) {
	//log.SetOutput(ioutil.Discard)

	var managerHandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "Torrent_Manager",
		MagicCookieValue: "factom_torrent",
	}

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: managerHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path+"factomd-torrent", "plugin"),
	})

	stop := make(chan int, 10)

	// Make sure we close our client on close
	AddInterruptHandler(func() {
		fmt.Println("Manager pluging is now closing...")
		client.Kill()
		stop <- 0
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("manager")
	if err != nil {
		return nil, err
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	manager := raw.(interfaces.IManagerController)

	if sigKey != nil {
		manager.SetSigningKey(sigKey.Key[:32])
	}

	go manageDrain(inQueue, manager, stop)

	return manager, nil
}

func manageDrain(inQueue chan interfaces.IMsg, man interfaces.IManagerController, quit chan int) {
	for {
		select {
		case <-quit:
			return
		default:
			fmt.Println(man.IsBufferEmpty())
			if !man.IsBufferEmpty() {
				var data []byte
				// Exit conditions: If empty, quit. If length == 1 and first/only byte it 0x00
				for !(man.IsBufferEmpty() || (len(data) == 1 && data[0] == 0x00)) {
					// Msgs are waiting!
					data = man.FetchFromBuffer()
					dbMsg := new(messages.DBStateMsg)
					err := dbMsg.UnmarshalBinary(data)
					if err != nil {
						log.Printf("%x %t\n", data, (len(data) == 1 && data[0] == 0x00))
						log.Println("Error unmarshaling dbstate from plugin: ", err)
						continue
					}

					inQueue <- dbMsg
				}
			}
			time.Sleep(CHECK_BUFFER)
		}
	}
}