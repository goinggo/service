package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"text/template"
)

//** NEW TYPES

// linuxUpstartService implements the Service interface
type linuxUpstartService struct {
	config *Config
}

//** PUBLIC MEMBER FUNCTIONS

// Install will create the necessary upstart conf file in the correct location
//  config: The configuration for the service
func (service *linuxUpstartService) Install(config *Config) error {
	confPath := service.getServiceFilePath()

	_, err := os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("Init already exists: %s", confPath)
	}

	file, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer file.Close()

	parameters := &struct {
		ExecutableName   string
		WorkingDirectory string
		Name             string
		DisplayName      string
		LongDescription  string
	}{
		service.config.ExecutableName,
		service.config.WorkingDirectory,
		service.config.Name,
		service.config.DisplayName,
		service.config.LongDescription,
	}

	template := template.Must(template.New("upstartScript").Parse(installScript()))
	return template.Execute(file, parameters)
}

// Remove uninstalls the service by removing the upstart conf file
func (service *linuxUpstartService) Remove() error {
	confPath := service.getServiceFilePath()

	return os.Remove(confPath)
}

// Start will execute the proper start command to start the service as a daemon
func (service *linuxUpstartService) Start() error {
	cmd := exec.Command("start", service.config.Name)

	return cmd.Run()
}

// Stop will execute the proper upstart command to stop the running service
func (service *linuxUpstartService) Stop() error {
	cmd := exec.Command("stop", service.config.Name)

	return cmd.Run()
}

// Run will start the service and hook into the OS and block. On a
// termination request by the OS it will call Stop and return
func (service *linuxUpstartService) Run(config *Config) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("******> SERVICE PANIC: %s\n", r)
		}
	}()

	fmt.Print("******> Initing Service\n")

	if config.Init != nil {
		err = config.Init()

		if err != nil {
			return err
		}
	}

	fmt.Print("******> Starting Service\n")

	if config.Start != nil {
		err = config.Start()

		if err != nil {
			return err
		}
	}

	fmt.Print("******> Service Started\n")

	// Create a channel to talk with the OS
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Wait for an event
	<-sigChan

	fmt.Print("******> Service Shutting Down\n")

	if config.Stop != nil {
		err = config.Stop()

		if err != nil {
			return err
		}
	}

	fmt.Print("******> Service Down\n")
	return err
}

//** PRIVATE MEMBER METHODS

// getServiceFilePath return the location for the launchd plist file
func (service *linuxUpstartService) getServiceFilePath() string {
	return fmt.Sprintf("/etc/init/%s.conf", service.config.Name)
}

//** PRIVATE METHODS

// newService create a new instance of the Service object for the Mac OS
func newService(config *Config) (service *linuxUpstartService, err error) {
	service = &linuxUpstartService{
		config: config,
	}

	if err != nil {
		return nil, err
	}

	return service, nil
}

// installScript returns a template for the launchd plist script for the Mac OSX
//  https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/launchd.plist.5.html
func installScript() (script string) {
	return `#	{{.LongDescription}}
description	{{.DisplayName}}

start on filesystem or runlevel [2345]
stop on runlevel [!2345]

#setuid username

kill signal INT

respawn
respawn limit 10 5
umask 022n

console none

pre-start script
	test -x {{.WorkingDirectory}}/{{.ExecutableName}} || { stop; exit 0; }
end script

# Start
	exec {{.WorkingDirectory}}/{{.ExecutableName}}
`
}
