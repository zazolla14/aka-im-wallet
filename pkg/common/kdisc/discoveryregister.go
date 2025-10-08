// Copyright © 2023 AkaIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kdisc

import (
	"time"

	"google.golang.org/grpc"

	"github.com/1nterdigital/aka-im-tools/discovery"
	"github.com/1nterdigital/aka-im-tools/discovery/etcd"
	"github.com/1nterdigital/aka-im-tools/discovery/kubernetes"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
)

const (
	ETCDCONST          = "etcd"
	KUBERNETESCONST    = "kubernetes"
	DefaultMessageSize = 20
	KB                 = 1024
	MB                 = KB * KB
	DefaultTimeout     = 10 * time.Second
)

// NewDiscoveryRegister creates a new service discovery and registry client based on the provided environment type.
func NewDiscoveryRegister(
	disc *config.Discovery, _ string, watchNames []string,
) (registry discovery.SvcDiscoveryRegistry, err error) {
	switch disc.Enable {
	case KUBERNETESCONST:
		return kubernetes.NewKubernetesConnManager(disc.Kubernetes.Namespace,
			disc.Etcd.Address, // TODO: find what is this for
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(MB*DefaultMessageSize),
			),
		)
	case ETCDCONST:
		return etcd.NewSvcDiscoveryRegistry(
			disc.Etcd.RootDirectory,
			disc.Etcd.Address,
			watchNames,
			etcd.WithDialTimeout(DefaultTimeout),
			etcd.WithMaxCallSendMsgSize(DefaultMessageSize*MB),
			etcd.WithUsernameAndPassword(disc.Etcd.Username, disc.Etcd.Password))
	default:
		return nil, errs.New("unsupported discovery type", "type", disc.Enable).Wrap()
	}
}
