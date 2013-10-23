package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"text/template"
)

//** NEW TYPES

// darwinLaunchdService implements the Service interface
type darwinLaunchdService struct {
	config *Config
}

//** PUBLIC MEMBER FUNCTIONS

// Install will create the necessary launchd plist file in the correct location
//  config: The configuration for the service
func (service *darwinLaunchdService) Install(config *Config) error {
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
		LogLocation      string
	}{
		service.config.ExecutableName,
		service.config.WorkingDirectory,
		service.config.Name,
		service.config.DisplayName,
		service.config.LongDescription,
		service.config.LogLocation,
	}

	template := template.Must(template.New("launchdConfig").Parse(installScript()))
	return template.Execute(file, parameters)
}

// Remove uninstalls the service by removing the launchd plist file
func (service *darwinLaunchdService) Remove() error {
	service.Stop()

	confPath := service.getServiceFilePath()

	return os.Remove(confPath)
}

// Start will execute the proper launchctl command to start the service as a daemon
func (service *darwinLaunchdService) Start() error {
	confPath := service.getServiceFilePath()

	cmd := exec.Command("launchctl", "load", confPath)

	return cmd.Run()
}

// Stop will execute the proper launchctl command to stop the running service
func (service *darwinLaunchdService) Stop() error {
	confPath := service.getServiceFilePath()

	cmd := exec.Command("launchctl", "unload", confPath)

	return cmd.Run()
}

// Run will start the service and hook into the OS and block. On a
// termination request by the OS it will call Stop and return
func (service *darwinLaunchdService) Run(config *Config) (err error) {
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
func (service *darwinLaunchdService) getServiceFilePath() string {
	return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", service.config.Name)
}

//** PRIVATE METHODS

// newService create a new instance of the Service object for the Mac OS
func newService(config *Config) (service *darwinLaunchdService, err error) {
	service = &darwinLaunchdService{
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
	return `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
<plist version='1.0'>
<dict>
	<key>Label</key><string>{{.DisplayName}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.WorkingDirectory}}/{{.ExecutableName}}</string>
	</array>
	<key>WorkingDirectory</key><string>{{.WorkingDirectory}}</string>
	<key>StandardOutPath</key><string>{{.LogLocation}}/{{.Name}}.log</string>
	<key>KeepAlive</key><true/>
	<key>Disabled</key><false/>
</dict>
</plist>`
}
