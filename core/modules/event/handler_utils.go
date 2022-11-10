package event

import (
	"fmt"
	"time"

	"github.com/hamster-shared/hamster-provider/log"
)

func successDealOrder(ctx *EventContext, orderNo uint64, name string) error {
	err := forwardSSHToP2p(ctx, name)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// report heartbeat
	agreementIndex := ctx.GetConfig().ChainRegInfo.AgreementIndex
	//_ = ctx.ReportClient.Heartbeat(agreementIndex)

	// send timed heartbeats
	go func() {
		// 启动心跳任务
		log.GetLogger().Info("定时心跳已启动")
		ticker := time.NewTicker(time.Second * 175)
		ctx.TimerService.SubTicker(agreementIndex, ticker)
		for {
			<-ticker.C
			// report heartbeat
			agreementIndex := ctx.GetConfig().ChainRegInfo.AgreementIndex
			err = ctx.ReportClient.HeartbeatDemo(agreementIndex)
			if err != nil {
				log.GetLogger().Error("heartbeat error", err)
			} else {
				log.GetLogger().Info("发送资源心跳成功")
			}
		}
	}()

	dealOverdueOrder(ctx, agreementIndex, name)
	return nil
}

func getVmTargetAddress(ctx *EventContext, name string) string {
	//ip, err := ctx.VmManager.GetIp(name)
	port := ctx.GetConfig().ApiPort
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port+1)
}

func forwardSSHToP2p(ctx *EventContext, name string) error {
	// P2P listen port exposure
	targetOpt := getVmTargetAddress(ctx, name)
	err := ctx.P2pClient.Listen("/x/ssh", targetOpt)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func dealOverdueOrder(ctx *EventContext, agreementIndex uint64, name string) bool {
	// calculate instance expiration time
	overdue := ctx.ReportClient.CalculateInstanceOverdue(ctx.GetConfig().ChainRegInfo.OrderIndex)
	instanceTimer := time.NewTimer(overdue)
	ctx.TimerService.SubTimer(agreementIndex, instanceTimer)

	go func(t *time.Timer) {
		<-t.C
		cfg := ctx.GetConfig()

		targetAddress := getVmTargetAddress(ctx, name)
		_, _ = ctx.P2pClient.Close(targetAddress)

		// expires triggers close
		_ = ctx.VmManager.Stop(name)
		_ = ctx.VmManager.Destroy(name)
		// modify the resource status on the chain to unused
		_ = ctx.ReportClient.ChangeResourceStatus(ctx.GetConfig().ChainRegInfo.ResourceIndex)
		// delete agreement number
		cfg.ChainRegInfo.OrderIndex = 0
		cfg.ChainRegInfo.AgreementIndex = 0
		cfg.ChainRegInfo.RenewOrderIndex = 0
		_ = ctx.Cm.Save(cfg)
	}(instanceTimer)

	return overdue < 0
}
