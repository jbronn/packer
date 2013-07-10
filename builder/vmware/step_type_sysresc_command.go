package vmware

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"text/template"
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

func (s *stepTypeSysRescCommand) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	httpPort := state["http_port"].(uint)
	ui := state["ui"].(packer.Ui)
	vncPort := state["vnc_port"].(uint)

	// Connect to VNC
	ui.Say("Connecting to VM via VNC")
	nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	c, err := vnc.Client(nc, &vnc.ClientConfig{Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	// Determine the host IP
	ipFinder := &IfconfigIPFinder{"vmnet8"}
	hostIp, err := ipFinder.HostIP()
	if err != nil {
		err := fmt.Errorf("Error detecting host IP: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	tplData := &bootCommandTemplateData{
		hostIp,
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the boot command over VNC...")
	for _, command := range config.SysRescCommand {
		var buf bytes.Buffer
		t := template.Must(template.New("boot").Parse(command))
		t.Execute(&buf, tplData)

		vncSendString(c, buf.String())
	}

	return multistep.ActionContinue
}

func (*stepTypeSysRescCommand) Cleanup(map[string]interface{}) {}
