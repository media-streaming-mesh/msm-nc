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
	"github.com/media-streaming-mesh/msm-nc/internal/config"
	node_mapper "github.com/media-streaming-mesh/msm-nc/pkg/node-mapper"
	stream_mapper "github.com/media-streaming-mesh/msm-nc/pkg/stream-mapper"
)

// App contains minimal list of dependencies to be able to start an application.
type App struct {
	cfg          *config.Cfg
	nodeMapper   *node_mapper.NodeMapper
	streamMapper *stream_mapper.StreamMapper
}

// Start, starts the MSM Network Controller application.
func (a *App) Start() error {
	logger := a.cfg.Logger
	logger.Info("Starting MSM Network Controller")

	//TODO: connect to dp when node mapper find dp

	go func() {
		a.nodeMapper.WatchNode()
	}()

	a.streamMapper.WatchStream()

	return nil
}
