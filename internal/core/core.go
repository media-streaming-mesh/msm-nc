/*
 * Copyright (c) 2022 Cisco and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"github.com/media-streaming-mesh/msm-k8s/pkg/model"
	"github.com/media-streaming-mesh/msm-k8s/pkg/node_mapper"
	"github.com/media-streaming-mesh/msm-nc/internal/config"
	"github.com/media-streaming-mesh/msm-nc/internal/stream_mapper"
)

// App contains minimal list of dependencies to be able to start an application.
type App struct {
	cfg          *config.Cfg
	nodeMapper   *node_mapper.NodeMapper
	streamMapper *stream_mapper.StreamMapper
	nodeChan     chan model.Node
}

// Start, starts the MSM Network Controller application.
func (a *App) Start() error {
	logger := a.cfg.Logger
	logger.Info("Starting MSM Network Controller")

	//Watch for node and connect to node via GRPC
	a.waitForData()
	go func() {
		a.nodeMapper.WatchNode(a.nodeChan)
	}()

	//Watch for streams and send streamgraph to dp
	a.streamMapper.WatchStream()

	return nil
}

func (a *App) waitForData() {
	go func() {
		node := <-a.nodeChan
		if node.State == model.AddNode {
			a.streamMapper.ConnectClient(node.IP)
		} else if node.State == model.DeleteNode {
			a.streamMapper.DisconnectClient(node.IP)
		}
		a.waitForData()
	}()
}
