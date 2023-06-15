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

// Package config
package config

import (
	"flag"
	"github.com/media-streaming-mesh/msm-nc/log"
	"os"

	"github.com/sirupsen/logrus"
)

// Cfg Config holds the configuration data for the MSM control plane
// application
type Cfg struct {
	DataPlane string
	Protocol  string
	Remote    string
	Logger    *logrus.Logger
	Grpc      *grpcOpts
}

type grpcOpts struct {
	Port string
}

// New initializes the configuration plugin and is shared across
// the internal plugins via the API interface
func New() *Cfg {
	cf := new(Cfg)
	grpcOpt := new(grpcOpts)

	flag.StringVar(&cf.DataPlane, "dataplane", "msm", "dataplane to connect to (msm, vpp)")
	flag.StringVar(&grpcOpt.Port, "grpcPort", "9000", "port to listen for GRPC on")
	flag.StringVar(&cf.Protocol, "protocol", "rtsp", "control plane protocol mode (rtsp, rist)")

	flag.Parse()

	cf.Logger = logrus.New()
	cf.Logger.SetOutput(os.Stdout)
	log.SetLogLvl(cf.Logger)
	log.SetLogType(cf.Logger)

	return &Cfg{
		DataPlane: cf.DataPlane,
		Protocol:  cf.Protocol,
		Logger:    cf.Logger,
		Remote:    cf.Remote,
		Grpc: &grpcOpts{
			Port: grpcOpt.Port,
		},
	}
}
