package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
)

// This step creates the VMX file for the VM.
//
// Uses:
//   config *config
//   sysresc_path string
//   ui     packer.Ui
//
// Produces:
type stepAttachSysResc struct{}

func (stepAttachSysResc) Run(state map[string]interface{}) multistep.StepAction {
	isoPath := state["sysresc_path"].(string)
	if isoPath == "" {
		return multistep.ActionContinue
	}

	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	ui.Say("Attaching System Rescue CD")

	// Getting existing VMX data.
	f, err := os.Open(vmxPath)
	if err != nil {
		err := fmt.Errorf("Error opening VMX for reading: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer f.Close()

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		err := fmt.Errorf("Error reading contents of VMX: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Setting the CD filename to the System Rescue CD, and make
	// VMware Fusion boot from it.
	vmxData := ParseVMX(string(vmxBytes))
	vmxData["ide1:0.fileName"] = isoPath
	vmxData["ide1:0.startConnected"] = "TRUE"
	vmxData["ide1:0.present"] = "TRUE"
	vmxData["bios.bootOrder"] = "CDROM"

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error creating VMX file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepAttachSysResc) Cleanup(map[string]interface{}) {}
