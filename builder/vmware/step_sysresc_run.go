package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step runs the created virtual machine.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepSysRescRun struct {
	bootTime time.Time
	vmxPath  string
}

func (s *stepSysRescRun) Run(state multistep.StateBag) multistep.StepAction {
	sysresc := state.Get("sysresc_path").(string)
	if sysresc == "" {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)
	vncPort := state.Get("vnc_port").(uint)

	// Set the VMX path so that we know we started the machine
	s.bootTime = time.Now()
	s.vmxPath = vmxPath

	ui.Say("Starting virtual machine...")
	if config.Headless {
		ui.Message(fmt.Sprintf(
			"The VM will be run headless, without a GUI. If you want to\n"+
				"view the screen of the VM, connect via VNC without a password to\n"+
				"127.0.0.1:%d", vncPort))
	}

	if err := driver.Start(vmxPath, config.Headless); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait the wait amount
	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait.String()))
		time.Sleep(config.bootWait)
	}

	return multistep.ActionContinue
}

func (s *stepSysRescRun) Cleanup(state multistep.StateBag) {}
