/*
	Package service provides a simple way to create a system service.
	Currently supports Windows, Linux/Upstart, and OSX/Launchd.

	Many services that run on different platforms cannot rely
	on flags to be passed for configuration. Some platforms
	require explicit install commands. This package handles the common
	boilerplate code. The following command may be passed to the
	executable as the first argument:

		install | remove | debug | start | stop

	These commands will do the following actions:

		install - Install the running executable as a service on the system.
		remove - Remove the running executable as a service on the system.
		debug - Run the service as a command line application.
		start - Starts the service via system controls.
		stop - Stops the service via system controls.

	Example Use

	The following shows a sample on how to use service. Pass debug on the command line to start.
	Then use the install and start command to run as a service.

		package main

		import (
			"fmt"
			"Kardianos/service"
			"path/filepath"
		)

		// main starts the application
		func main() {

			// Capture the working directory
			workingDirectory, _ := filepath.Abs("")

			// Create a config object to start the service
			config := &service.Config{
				ExecutableName:   "MyService",
				WorkingDirectory: workingDirectory,
				Name:             "MyService",
				DisplayName:      "My Service",
				LongDescription:  "My Service provides support for...",
				LogLocation:      _Straps.Strap("baseFilePath"),

				Init:  InitService,
				Start: StartService,
				Stop:  StopService,
			}

			// Run any command line options or start the service
			config.Run()
		}

		// InitService is called by the ServiceManager when the server before the service is ready to start
		func InitService() {

			fmt.Printf("Service Inited")
		}

		// StartService is called by the ServiceManager when the server is ready to start
		func StartService() {

			fmt.Printf("Service Started\n")
		}

		// StopService is called by the ServiceManager when the server is ready to be shutdown
		func StopService() {

			fmt.Printf("Service Stopped\n")
		}
*/
package service

import ()

//** NEW TYPES

// Service must be implemented by all OS based service implementations
type Service interface {
	Installer
	Controller
	Runner
}

// Config provides detailed information for command and control of the service
type Config struct {
	ExecutableName   string // The name of the application
	WorkingDirectory string // The working directory fo the application
	Name             string // The internal name of the service. It should not contain spaces
	DisplayName      string // The pretty print display name for the application
	LongDescription  string // The description of the service
	LogLocation      string // If install files support this, the file location to write all stdout messages

	Init  func() error // Called to init the service before it starts
	Start func() error // Called when the service starts. IT MUST NOT BLOCK
	Stop  func() error // Called when the service stops. IT MUST NOT BLOCK

	service Service // A reference to the service object
}

//** INTERFACES

// Installer implements the way to install and remove the service from the target OS
type Installer interface {
	// Installs this service on the system.  May return an
	// error if this service is already installed.
	Install(config *Config) error

	// Removes this service from the system.  May return an
	// error if this service is not already installed.
	Remove() error
}

// Controller implements the way to self start and stop the service
type Controller interface {
	// Starts this service on the system
	Start() error

	// Stops this service on the system
	Stop() error
}

// Runner allows the program to run as a service
type Runner interface {
	// Run the program as a service
	Run(config *Config) error
}
