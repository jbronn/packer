package virtualbox

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"text/template"
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

func (s *stepTypeSysRescCommand) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	httpPort := state["http_port"].(uint)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	tplData := &bootCommandTemplateData{
		"10.0.2.2",
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the System Rescue CD command...")
	for _, command := range config.SysRescCommand {
		var buf bytes.Buffer
		t := template.Must(template.New("boot").Parse(command))
		t.Execute(&buf, tplData)

		for _, code := range scancodes(buf.String()) {
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
			if _, ok := state[multistep.StateCancelled]; ok {
				return multistep.ActionHalt
			}

			if err := driver.VBoxManage("controlvm", vmName, "keyboardputscancode", code); err != nil {
				err := fmt.Errorf("Error sending System Rescue CD command: %s", err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (*stepTypeSysRescCommand) Cleanup(map[string]interface{}) {}
