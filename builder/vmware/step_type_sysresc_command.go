package vmware

import (
	"fmt"
	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"runtime"
)

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   config *config
//   http_port int
//   ui     packer.Ui
//   vnc_port uint
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
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vncPort := state.Get("vnc_port").(uint)

	// Connect to VNC
	ui.Say("Connecting to VM via VNC")
	nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	c, err := vnc.Client(nc, &vnc.ClientConfig{Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	// Determine the host IP
	var ipFinder HostIPFinder
	if runtime.GOOS == "windows" {
		ipFinder = new(VMnetNatConfIPFinder)
	} else {
		ipFinder = &IfconfigIPFinder{Device: "vmnet8"}
	}

	hostIp, err := ipFinder.HostIP()
	if err != nil {
		err := fmt.Errorf("Error detecting host IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Host IP for the VMware machine: %s", hostIp)

	tplData := &bootCommandTemplateData{
		hostIp,
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the boot command over VNC...")
	for _, command := range config.SysRescCommand {
		command, err := config.tpl.Process(command, tplData)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Check for interrupts between typing things so we can cancel
		// since this isn't the fastest thing.
		if _, ok := state.GetOk(multistep.StateCancelled); ok {
			return multistep.ActionHalt
		}

		vncSendString(c, command)
	}

	return multistep.ActionContinue
}

func (*stepTypeSysRescCommand) Cleanup(multistep.StateBag) {}
