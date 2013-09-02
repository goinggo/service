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

// _DarwinLaunchdService implements the Service interface
type _DarwinLaunchdService struct {
	_Config *Config
}

//** PUBLIC MEMBER FUNCTIONS

// Install will create the necessary launchd plist file in the correct location
//  config: The configuration for the service
func (service *_DarwinLaunchdService) Install(config *Config) error {

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
		LogLocation      string
	}{
		service._Config.ExecutableName,
		service._Config.WorkingDirectory,
		service._Config.Name,
		service._Config.DisplayName,
		service._Config.LongDescription,
		service._Config.LogLocation,
	}

	template := template.Must(template.New("launchdConfig").Parse(_InstallScript()))
	return template.Execute(file, parameters)
}

// Remove uninstalls the service by removing the launchd plist file
func (service *_DarwinLaunchdService) Remove() error {

	service.Stop()

	confPath := service._GetServiceFilePath()

	return os.Remove(confPath)
}

// Start will execute the proper launchctl command to start the service as a daemon
func (service *_DarwinLaunchdService) Start() error {

	confPath := service._GetServiceFilePath()

	cmd := exec.Command("launchctl", "load", confPath)

	return cmd.Run()
}

// Stop will execute the proper launchctl command to stop the running service
func (service *_DarwinLaunchdService) Stop() error {

	confPath := service._GetServiceFilePath()

	cmd := exec.Command("launchctl", "unload", confPath)

	return cmd.Run()
}

// Run will start the service and hook into the OS and block. On a
// termination request by the OS it will call Stop and return
func (service *_DarwinLaunchdService) Run(config *Config) (err error) {

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
func (service *_DarwinLaunchdService) _GetServiceFilePath() string {

	return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", service._Config.Name)
}

//** PRIVATE METHODS

// _NewService create a new instance of the Service object for the Mac OS
func _NewService(config *Config) (service *_DarwinLaunchdService, err error) {

	service = &_DarwinLaunchdService{
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
