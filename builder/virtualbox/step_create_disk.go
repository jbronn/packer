package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strconv"
	"strings"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	format := "VDI"
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", vmName, strings.ToLower(format)))

	command := []string{
		"createhd",
		"--filename", path,
		"--size", strconv.FormatUint(uint64(config.DiskSize), 10),
		"--format", format,
		"--variant", "Standard",
	}

	ui.Say("Creating hard drive...")
	err := driver.VBoxManage(command...)
	if err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Add the IDE controller so we can later attach the disk.
	// When the hard disk controller is not IDE, this device is still used
	// by VirtualBox to deliver the guest extensions.
	controllerName := "IDE Controller"
	err = driver.VBoxManage("storagectl", vmName, "--name", controllerName, "--add", "ide")
	if err != nil {
		err := fmt.Errorf("Error creating disk controller: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Add a SATA controller if we were asked to use SATA. We still attach
	// the IDE controller above because some other things (disks) require
	// that.
	if config.HardDriveInterface == "sata" {
		controllerName = "SATA Controller"
		command = []string{
			"storagectl", vmName,
			"--name", controllerName,
			"--add", "sata",
			"--sataportcount", "1",
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error creating disk controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Attach the disk to the controller
	command = []string{
		"storageattach", vmName,
		"--storagectl", controllerName,
		"--port", "0",
		"--device", "0",
		"--type", "hdd",
		"--medium", path,
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error attaching hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
