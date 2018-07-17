// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/goodrain/rainbond/cmd/node/option"
	"github.com/goodrain/rainbond/node/core/config"
	"github.com/goodrain/rainbond/node/core/job"
)

//ClusterClient ClusterClient
type ClusterClient interface {
	UpdateStatus(*HostNode) error
	DownNode(*HostNode) error
	GetMasters() ([]*HostNode, error)
	GetNode(nodeID string) (*HostNode, error)
	GetDataCenterConfig() (*config.DataCenterConfig, error)
	WatchJobs() <-chan *job.Event

	//WatchTasks()
	//UpdateTask()
}

//NewClusterClient new cluster client
func NewClusterClient(conf *option.Conf, etcdClient *clientv3.Client) ClusterClient {
	return &etcdClusterClient{
		etcdClient: etcdClient,
		conf:       conf,
	}

}

type etcdClusterClient struct {
	etcdClient *clientv3.Client
	conf       *option.Conf
	onlineLes  clientv3.LeaseID
}

func (e *etcdClusterClient) UpdateStatus(n *HostNode) error {
	n.UpTime = time.Now()
	n.Alived = true
	if err := e.Update(n); err != nil {
		return err
	}
	if err := e.nodeOnlinePut(n); err != nil {
		return err
	}
	return nil
}

func (e *etcdClusterClient) GetMasters() ([]*HostNode, error) {
	return nil, nil
}

func (e *etcdClusterClient) GetDataCenterConfig() (*config.DataCenterConfig, error) {
	return nil, nil
}

func (e *etcdClusterClient) WatchJobs() <-chan *job.Event {
	return nil
}

func (e *etcdClusterClient) GetNode(nodeID string) (*HostNode, error) {
	return nil, nil
}

//nodeOnlinePut onde noline status update
func (e *etcdClusterClient) nodeOnlinePut(h *HostNode) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if e.onlineLes != 0 {
		if _, err := e.etcdClient.KeepAlive(ctx, e.onlineLes); err == nil {
			return nil
		}
		e.onlineLes = 0
	}
	les, err := e.etcdClient.Grant(ctx, 30)
	if err != nil {
		return err
	}
	e.onlineLes = les.ID
	_, err = e.etcdClient.Put(ctx, e.conf.OnlineNodePath+"/"+h.ID, h.PID, clientv3.WithLease(les.ID))
	if err != nil {
		return err
	}
	return nil
}

//Update update node info
func (e *etcdClusterClient) Update(h *HostNode) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := e.etcdClient.Put(ctx, e.conf.NodePath+"/"+h.ID, h.String())
	return err
}

//Down node
func (e *etcdClusterClient) DownNode(h *HostNode) error {
	h.Alived, h.DownTime = false, time.Now()
	e.Update(h)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := e.etcdClient.Delete(ctx, e.conf.OnlineNodePath+"/"+h.ID)
	return err
}
