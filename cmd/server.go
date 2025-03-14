package cmd

import (
	"fmt"
	"log"

	"github.com/ovatu/redalert/config"
	"github.com/ovatu/redalert/core"
	"github.com/ovatu/redalert/notifiers"
	"github.com/ovatu/redalert/rpc"
	"github.com/ovatu/redalert/storage"
	"github.com/ovatu/redalert/web"
	"github.com/spf13/cobra"
)

type serverConfig struct {
	webPort      int
	disableBrand bool
	readOnly	 bool
	rpcPort      int
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run checks and server stats",
	Long:  "Run checks and server stats",
	Run: func(cmd *cobra.Command, args []string) {
		var configStore config.Store
		var err error
		if cmd.Flag("config-db").Changed {
			log.Println("Config via db")
			configDb := cmd.Flag("config-db").Value.String()
			configStore, err = config.NewDBStore(configDb)
			if err != nil {
				log.Fatal("DB config error via :", configDb, " Error: ", err)
			}
		} else if cmd.Flag("config-url").Changed {
			log.Println("Config via URL")
			configURL := cmd.Flag("config-url").Value.String()
			configStore, err = config.NewURLStore(configURL)
			if err != nil {
				log.Fatal("URL config error via :", configURL, " Error: ", err)
			}
		} else if cmd.Flag("config-s3").Changed {
			log.Println("Config via S3")
			configS3 := cmd.Flag("config-s3").Value.String()
			configStore, err = config.NewS3Store(configS3)
			if err != nil {
				log.Fatal("S3 config error via :", configS3, " Error: ", err)
			}
		} else if cmd.Flag("config-env").Changed {
			log.Println("Config via environment variable")
			configEnv := cmd.Flag("config-env").Value.String()
			configStore, err = config.NewEnvStore(configEnv)
			if err != nil {
				log.Fatal("Env config error via :", configEnv, " Error: ", err)
			}
		} else {
			log.Println("Config via file")
			configFile := cmd.Flag("config-file").Value.String()
			configStore, err = config.NewFileStore(configFile)
			if err != nil {
				log.Fatal("File config error via :", configFile, " Error: ", err)
			}
		}
		runServer(configStore, serverConfig{
			webPort:      webPort,
			disableBrand: disableBrand,
			readOnly: 	  readOnly,
			rpcPort:      rpcPort,
		})
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func runServer(configStore config.Store, cfg serverConfig) {
	// Event Storage
	const MaxEventsStored = 100

	service := core.NewService()

	// Setup StdErr Notifications

	stdErrNotifier, err := notifiers.New(notifiers.Config{
		Name: "stderr",
		Type: "stderr",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = service.RegisterNotifier( "stderr", stdErrNotifier)
	if err != nil {
		log.Fatal(err)
	}

	// Load Notifications

	savedNotifications, err := configStore.Notifications()
	if err != nil {
		log.Fatal(err)
	}

	for _, notificationConfig := range savedNotifications {

		notifier, err := notifiers.New(notificationConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = service.RegisterNotifier(notificationConfig.Name, notifier)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Load Checks
	var checkIdx int

	savedChecks, err := configStore.Checks()
	if err != nil {
		log.Fatal(err)
	}

	preferences, err := configStore.Preferences()
	if err != nil {
		log.Fatal(err)
	}

	for _, checkConfig := range savedChecks {

		check, err := core.NewCheck(checkConfig, storage.NewMemoryList(MaxEventsStored), preferences)
		if err != nil {
			log.Fatal(err)
		}

		err = service.RegisterCheck(check, checkConfig.SendAlerts, checkIdx)
		if err != nil {
			log.Fatal(err)
		}
		checkIdx++
	}

	service.Start()

	go web.Run(service, cfg.webPort, cfg.disableBrand, cfg.readOnly)
	go rpc.Run(service, cfg.rpcPort, cfg.readOnly)
	fmt.Println(`
____ ____ ___  ____ _    ____ ____ ___
|--< |=== |__> |--| |___ |=== |--<  |

`)
	fmt.Println("Web Running on port ", cfg.webPort)
	fmt.Println("RPC Running on port ", cfg.rpcPort)
	fmt.Println("Read only ", cfg.readOnly)

	service.KeepRunning()
}
