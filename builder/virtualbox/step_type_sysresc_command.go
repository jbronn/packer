package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   config *config
//   driver Driver
//   http_port int
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type stepTypeSysRescCommand struct{}

func (s *stepTypeSysRescCommand) Run(state multistep.StateBag) multistep.StepAction {
	sysresc := state.Get("sysresc_path").(string)
	if sysresc == "" {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	tplData := &bootCommandTemplateData{
		"10.0.2.2",
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the System Rescue CD command...")
	for _, command := range config.SysRescCommand {
		command, err := config.tpl.Process(command, tplData)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		for _, code := range scancodes(command) {
			if code == "wait" {
				time.Sleep(1 * time.Second)
				continue
			}

			if code == "wait5" {
				time.Sleep(5 * time.Second)
				continue
			}

			if code == "wait10" {
				time.Sleep(10 * time.Second)
				continue
			}

			// Since typing is sometimes so slow, we check for an interrupt
			// in between each character.
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}

			if err := driver.VBoxManage("controlvm", vmName, "keyboardputscancode", code); err != nil {
				err := fmt.Errorf("Error sending System Rescue CD command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (*stepTypeSysRescCommand) Cleanup(multistep.StateBag) {}
