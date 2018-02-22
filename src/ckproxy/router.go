package ckproxy

import (
	"fmt"
	"time"
)

type RouterInterface interface {

	IsOnline() bool
	AddSuccessCount(count int)
	AddFaileCount(count int)
	AddInFlow(flow int)
	AddOutFlow(flow int)
	NextAddr() string
	String() string

}

type Router struct {
	Url               string
	Type              int
	FaileCount        int
	TotalFaileCount   int
	TotalSuccessCount int
	Status            bool      // true: online; false: offline
	RecoveryTime      time.Time //恢复时间
	InFlow            uint64    //网络流量
	OutFlow           uint64    //网络流量
}


/**
 *	service interface status is online or offline.
 *  only for http router interface
 */
func (r *Router) IsOnline() bool {
	if !r.Status {
		now := time.Now()
		if r.RecoveryTime.Before(now) {
			r.Status = true
		}
	}
	return r.Status
}

func (r *Router) offline(recoverySec int) {
	r.Status = false
	r.RecoveryTime = time.Now().Add(time.Duration(recoverySec) * time.Second)
}

func (r *Router) AddInFlow(flow int) {
	r.InFlow += uint64(flow)
}

func (r *Router) AddOutFlow(flow int) {
	r.OutFlow += uint64(flow)
}

func (r *Router) AddFaileCount(count int) {
	r.FaileCount += count
	r.TotalFaileCount += count
	if r.FaileCount >= Instance().OfflineFailedCount {
		r.offline(Instance().RecoverySec)
	}
}

func (r *Router) AddSuccessCount(count int) {
	r.TotalSuccessCount += count
	r.FaileCount = 0
}

func (r *Router) String() string {
	return fmt.Sprintf("url[%v] success[%v] failed[%v] in[%v] out[%v]", r.Url, r.TotalSuccessCount, r.TotalFaileCount, r.InFlow/1024/1024, r.OutFlow/1024/1024)
}

func (r *Router) NextAddr() string {
	return r.Url
}

type ApiSystmeRepairRouter struct {
	Router
}

func (r *ApiSystmeRepairRouter) NextAddr() string {
	return r.Url
}
