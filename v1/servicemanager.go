package service

import (
	"bufio"
	"fmt"
	"os"
)

//** PUBLIC MEMBER FUNCTIONS

// Fill in configuration, then call Run() to setup basic handling.
// Blocks until program completes. Is intended to handle the standard
// simple cases for running a service.
func (config *Config) Run() {
	service, err := newService(config)

	if err != nil {
		fmt.Printf("%s unable to start: %s", config.DisplayName, err)
		return
	}

	config.service = service

	// Perform a command and then return
	if len(os.Args) > 1 {
		verb := os.Args[1]

		switch verb {
		case "install":
			err = service.Install(config)

			if err != nil {
				fmt.Printf("Failed to install: %s\n", err)
				return
			}

			fmt.Printf("Service \"%s\" installed.\n", config.DisplayName)
			return

		case "remove":
			err = service.Remove()

			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
				return
			}

			fmt.Printf("Service \"%s\" removed.\n", config.DisplayName)
			return

		case "debug":
			err = config.Start()

			if err != nil {
				fmt.Printf("Error Starting Service : %s", err)
				return
			}

			fmt.Println("Starting Up In Debug Mode")

			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')

			fmt.Println("Shutting Down")

			config.Stop()
			return

		case "start":
			err = service.Start()

			if err != nil {
				fmt.Printf("Failed to start: %s\n", err)
				return
			}

			fmt.Printf("Service \"%s\" started.\n", config.DisplayName)
			return

		case "stop":
			err = service.Stop()

			if err != nil {
				fmt.Printf("Failed to stop: %s\n", err)
				return
			}

			fmt.Printf("Service \"%s\" stopped.\n", config.DisplayName)
			return

		default:
			fmt.Printf("Options for \"%s\": (install | remove | debug | start | stop)\n", os.Args[0])
			return
		}
	}

	// Run the service
	err = service.Run(config)
}
