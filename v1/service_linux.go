package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"
)

//** NEW TYPES

// _LinuxUpstartService implements the Service interface
type _LinuxUpstartService struct {
	_Config *Config
}

//** PUBLIC MEMBER FUNCTIONS

// Install will create the necessary upstart conf file in the correct location
//  config: The configuration for the service
func (service *_LinuxUpstartService) Install(config *Config) error {

	confPath := service._GetServiceFilePath()

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
		service._Config.ExecutableName,
		service._Config.WorkingDirectory,
		service._Config.Name,
		service._Config.DisplayName,
		service._Config.LongDescription,
	}

	template := template.Must(template.New("upstartScript").Parse(_InstallScript()))
	return template.Execute(file, parameters)
}

// Remove uninstalls the service by removing the upstart conf file
func (service *_LinuxUpstartService) Remove() error {

	confPath := service._GetServiceFilePath()

	return os.Remove(confPath)
}

// Start will execute the proper start command to start the service as a daemon
func (service *_LinuxUpstartService) Start() error {

	cmd := exec.Command("start", service._Config.Name)

	return cmd.Run()
}

// Stop will execute the proper upstart command to stop the running service
func (service *_LinuxUpstartService) Stop() error {

	cmd := exec.Command("stop", service._Config.Name)

	return cmd.Run()
}

// Run will start the service and hook into the OS and block. On a
// termination request by the OS it will call Stop and return
func (service *_LinuxUpstartService) Run(config *Config) (err error) {

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

	// Ask the OS to notify us about events
	signal.Notify(sigChan)

	for {

		// Wait for an event
		whatSig := <-sigChan

		// Convert the signal to an integer so we can display the hex number
		sigAsInt, _ := strconv.Atoi(fmt.Sprintf("%d", whatSig))

		fmt.Printf("******> OS Notification: %v : %#x\n", whatSig, sigAsInt)

		// Did we get any of these termination events
		if whatSig == syscall.SIGINT ||
			whatSig == syscall.SIGKILL ||
			whatSig == syscall.SIGQUIT ||
			whatSig == syscall.SIGSTOP ||
			whatSig == syscall.SIGTERM {

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
	}
}

//** PRIVATE MEMBER METHODS

// _GetServiceFilePath return the location for the launchd plist file
func (service *_LinuxUpstartService) _GetServiceFilePath() string {

	return fmt.Sprintf("/etc/init/%s.conf", service._Config.Name)
}

//** PRIVATE METHODS

// _NewService create a new instance of the Service object for the Mac OS
func _NewService(config *Config) (service *_LinuxUpstartService, err error) {

	service = &_LinuxUpstartService{
		_Config: config,
	}

	if err != nil {
		return nil, err
	}

	return service, nil
}

// _InstallScript returns a template for the launchd plist script for the Mac OSX
//  https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/launchd.plist.5.html
func _InstallScript() (script string) {

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
