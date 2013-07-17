package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step starts the virtual machine.
//
// Uses:
//
// Produces:
type stepSysRescRun struct {
	vmName string
}

func (s *stepSysRescRun) Run(state map[string]interface{}) multistep.StepAction {
	sysresc := state["sysresc_path"].(string)
	if sysresc == "" {
		return multistep.ActionContinue
	}

	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	ui.Say("Starting the virtual machine...")
	guiArgument := "gui"
	if config.Headless == true {
		ui.Message("WARNING: The VM will be started in headless mode, as configured.\n" +
			"In headless mode, errors during the boot sequence or OS setup\n" +
			"won't be easily visible. Use at your own discretion.")
		guiArgument = "headless"
	}
	command := []string{"startvm", vmName, "--type", guiArgument}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait))
		time.Sleep(config.bootWait)
	}

	return multistep.ActionContinue
}


func (s *stepSysRescRun) Cleanup(state map[string]interface{}) {
	if s.vmName == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.VBoxManage("controlvm", s.vmName, "poweroff"); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
