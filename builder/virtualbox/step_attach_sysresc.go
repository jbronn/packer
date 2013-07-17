package virtualbox

import (
        "fmt"
        "github.com/mitchellh/multistep"
        "github.com/mitchellh/packer/packer"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//
// Produces:
type stepAttachSysResc struct {
        diskPath string
}

func (s *stepAttachSysResc) Run(state map[string]interface{}) multistep.StepAction {
	isoPath := state["sysresc_path"].(string)
	if isoPath == "" {
		return multistep.ActionContinue
	}

        driver := state["driver"].(Driver)
        ui := state["ui"].(packer.Ui)
        vmName := state["vmName"].(string)

        // Attach the disk to the controller
        attach_command := []string{
                "storageattach", vmName,
                "--storagectl", "IDE Controller",
                "--port", "0",
                "--device", "1",
                "--type", "dvddrive",
                "--medium", isoPath,
        }
        if err := driver.VBoxManage(attach_command...); err != nil {
                err := fmt.Errorf("Error attaching ISO: %s", err)
                state["error"] = err
                ui.Error(err.Error())
                return multistep.ActionHalt
        }

        // Make it so System Rescue CD is booted up instead of the hard disk.
        modify_command := []string{
                "modifyvm", vmName,
                "--boot1", "dvd",
        }
        if err := driver.VBoxManage(modify_command...); err != nil {
                err := fmt.Errorf("Error modifying VM to boot from System Rescue CD: %s", err)
                state["error"] = err
                ui.Error(err.Error())
                return multistep.ActionHalt
        }

        // Track the path so that we can unregister it from VirtualBox later
        s.diskPath = isoPath

        return multistep.ActionContinue
}

func (s *stepAttachSysResc) Cleanup(state map[string]interface{}) {}
