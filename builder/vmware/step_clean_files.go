package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path/filepath"
)

// These are the extensions of files that are important for the function
// of a VMware virtual machine. Any other file is discarded as part of the
// build.
var KeepFileExtensions = []string{".vmdk", ".vmx"}

// This step removes unnecessary files from the final result.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepCleanFiles struct{}

func (stepCleanFiles) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	// Restoring VMX values.
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

	vmxData := ParseVMX(string(vmxBytes))
	vmxData["ide1:0.deviceType"] = "atapi-cdrom"
	vmxData["ide1:0.startConnected"] = "False"
	delete(vmxData, "ide1:0.fileName")
	delete(vmxData, "bios.bootOrder")

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error creating VMX file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Deleting unnecessary VMware files...")
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// If the file isn't critical to the function of the
			// virtual machine, we get rid of it.
			keep := false
			ext := filepath.Ext(path)
			for _, goodExt := range KeepFileExtensions {
				if goodExt == ext {
					keep = true
					break
				}
			}

			if !keep {
				ui.Message(fmt.Sprintf("Deleting: %s", path))
				return os.Remove(path)
			}
		}

		return nil
	}

	if err := filepath.Walk(config.OutputDir, visit); err != nil {
		state["error"] = err
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCleanFiles) Cleanup(map[string]interface{}) {}
