// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"

	"syscall"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/blockchain/protoManager"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/config"
)

// Node represents the blockchain zipper
type Node struct {
	bc  *blockchain.Blockchain
	cfg *config.Option
}

// NewNode returns node daemon instance
func NewNode(cfgFile string) *Node {
	if err := config.ReadInConfig(cfgFile); err != nil {
		log.Panicf("loadConfig error %v", err)
		return nil
	}

	cfg := config.NodeOption()
	log.New(cfg.LogFile)
	log.SetLevel(cfg.LogLevel)
	config.VMConfig(cfg.LogFile, cfg.LogLevel)
	pm := protoManager.NewProtoManager()
	node := &Node{
		bc:  blockchain.NewBlockchain(pm),
		cfg: cfg,
	}

	return node
}

// Start starts the blockchain service
func (node *Node) Start() {
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)

	if node.cfg.ProfPort != "" {
		go func() {
			err := http.ListenAndServe(":"+node.cfg.ProfPort, nil)
			if err != nil {
				log.Errorf("Prof Server start error=%v", err)
			} else {
				log.Debugf("Prof Server start on port=%s", node.cfg.ProfPort)
			}
		}()
	}

	if node.cfg.CPUFile != "" {
		cpuFile := node.cfg.CPUFile
		cpuProfile, _ := os.Create(cpuFile)
		pprof.StartCPUProfile(cpuProfile)

		defer func() {
			memFile := cpuFile + ".mem"
			pprof.StopCPUProfile()
			memProfile, _ := os.Create(memFile)
			pprof.WriteHeapProfile(memProfile)
			memProfile.Close()
			cpuProfile.Close()
		}()
	}

	node.bc.Start()

	<-abort
	node.bc.Stop()
	os.Exit(0)
}
