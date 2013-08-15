package vmware

import (
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

// This step downloads the System Rescue CD specified.
//
// Uses:
//   cache packer.Cache
//   config *config
//   ui     packer.Ui
//
// Produces:
//   sysresc_path string
type stepDownloadSysResc struct{}

func (s stepDownloadSysResc) Run(state map[string]interface{}) multistep.StepAction {
	cache := state["cache"].(packer.Cache)
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	if config.SysRescURL == "" {
		state["sysresc_path"] = ""
		return multistep.ActionContinue
	}

	checksum, err := hex.DecodeString(config.SysRescChecksum)
	if err != nil {
		err := fmt.Errorf("Error parsing checksum: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Acquiring lock to download the System Rescue CD.")
	cachePath := cache.Lock(config.SysRescURL)
	defer cache.Unlock(config.SysRescURL)

	downloadConfig := &common.DownloadConfig{
		Url:        config.SysRescURL,
		TargetPath: cachePath,
		CopyFile:   false,
		Hash:       common.HashForType(config.SysRescChecksumType),
		Checksum:   checksum,
	}

	download := common.NewDownloadClient(downloadConfig)

	downloadCompleteCh := make(chan error, 1)
	go func() {
		ui.Say("Copying or downloading System Rescue CD. Progress will be reported periodically.")
		cachePath, err = download.Get()
		downloadCompleteCh <- err
	}()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

DownloadWaitLoop:
	for {
		select {
		case err := <-downloadCompleteCh:
			if err != nil {
				err := fmt.Errorf("Error downloading System Rescue CD: %s", err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			break DownloadWaitLoop
		case <-progressTicker.C:
			progress := download.PercentProgress()
			if progress >= 0 {
				ui.Message(fmt.Sprintf("Download progress: %d%%", progress))
			}
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				ui.Say("Interrupt received. Cancelling download...")
				return multistep.ActionHalt
			}
		}
	}

	log.Printf("Path to System Rescue CD on disk: %s", cachePath)
	state["sysresc_path"] = cachePath

	return multistep.ActionContinue
}

func (stepDownloadSysResc) Cleanup(map[string]interface{}) {}
