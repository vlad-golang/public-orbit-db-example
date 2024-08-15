package main

import (
	orbitdb "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/accesscontroller"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"go.uber.org/zap"
	"log"
	"log/slog"
	"time"
)

func main() {
	ctx := context.Background()

	node, err := core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	})
	slog.Info("My ID:", slog.StringValue(node.PeerHost.ID().String()))

	if err != nil {
		log.Fatalf("Failed to start IPFS node: %v", err)
	}
	defer node.Close()

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		log.Fatalf("Failed to create CoreAPI: %v", err)
	}

	db, err := orbitdb.NewOrbitDB(ctx, api, &orbitdb.NewOrbitDBOptions{
		Logger: zap.NewExample(),
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create or open a public database
	options := &orbitdb.CreateDBOptions{
		Replicate: null.ValueFrom(false).Ptr(),
		AccessController: &accesscontroller.CreateAccessControllerOptions{
			Access: map[string][]string{
				"write": {"*"}, // Allow public write access
				"admin": {"*"},
			},
		},
	}
	store, err := db.KeyValue(ctx, "public-db", options)
	defer store.Close()

	fmt.Println("Database address:", store.Address().String())

	go func() {
		for {
			time.Sleep(5 * time.Second)

			_, err = store.Put(ctx, uuid.NewString(), []byte(uuid.NewString()))
			if err != nil {
				log.Fatalf("Failed to put value in store: %v", err)
			}

			value, err := store.Get(ctx, "name")
			if err != nil {
				log.Fatalf("Failed to get value from store: %v", err)
			}

			fmt.Printf("Value for 'name': %s. All: %v\n", value, len(store.All()))
		}
	}()

	select {}
}
