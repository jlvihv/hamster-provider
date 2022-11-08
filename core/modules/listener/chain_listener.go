package listener

import (
	"context"
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	chain2 "github.com/hamster-shared/hamster-provider/core/modules/chain"
	"github.com/hamster-shared/hamster-provider/core/modules/config"
	"github.com/hamster-shared/hamster-provider/core/modules/demo"
	"github.com/hamster-shared/hamster-provider/core/modules/event"
	"github.com/hamster-shared/hamster-provider/core/modules/utils"
	"github.com/hamster-shared/hamster-provider/log"
)

type ChainListener struct {
	eventService event.IEventService
	api          *gsrpc.SubstrateAPI
	cm           *config.ConfigManager
	reportClient chain2.ReportClient
	cancel       func()
	ctx          context.Context
}

func NewChainListener(eventService event.IEventService, cm *config.ConfigManager) *ChainListener {
	return &ChainListener{
		eventService: eventService,
		cm:           cm,
	}
}

func (l *ChainListener) SetChainApi(api *gsrpc.SubstrateAPI, reportClient chain2.ReportClient) {
	l.api = api
	l.reportClient = reportClient
}

func (l *ChainListener) GetState() bool {
	return l.cancel != nil
}

func (l *ChainListener) SetState(option bool) error {
	if option {
		return l.start()
	} else {
		return l.stop()
	}
}

func (l *ChainListener) start() error {
	log.GetLogger().Debug("chain listener start")
	if l.cancel != nil {
		l.cancel()
	}

	cfg, err := l.cm.GetConfig()
	if err != nil {
		return err
	}

	demo := chain2.ResourceInfoDemo{
		PeerID:   cfg.Identity.PeerID,
		PublicIP: cfg.PublicIP,
		CPU:      uint8(cfg.Vm.Cpu),
		Memory:   uint8(cfg.Vm.Mem),
	}

	err = l.reportClient.RegisterResourceDemo(demo)

	if err != nil {
		log.GetLogger().Errorf("register resource error: ", err)
		return err
	}
	log.GetLogger().Info("register resource success")

	l.ctx, l.cancel = context.WithCancel(context.Background())
	isPanic := make(chan bool)
	go l.setWatchEventState(l.ctx, isPanic)
	return nil
}

func (l *ChainListener) setWatchEventState(ctx context.Context, isPanic chan bool) {
	for {
		go l.watchEvent(ctx, isPanic)
		<-isPanic
		go l.watchEvent(ctx, isPanic)
	}
}

func (l *ChainListener) stop() error {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	cfg, err := l.cm.GetConfig()
	if err != nil {
		return err
	}
	// thegraph.SetIsServer(false)
	log.GetLogger().Info("调用 ChainListener 里的 stop 方法，删除", cfg.ChainRegInfo.ResourceIndex)
	return l.reportClient.RemoveResourceDemo(cfg.ChainRegInfo.ResourceIndex)
}

// WatchEvent chain event listener
func (l *ChainListener) watchEvent(ctx context.Context, channel chan bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("watchEventError:", err)
			channel <- true
		}
	}()

	meta, err := l.api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	// Subscribe to system events via storage
	key, err := types.CreateStorageKey(meta, "System", "Events", nil)
	if err != nil {
		panic(err)
	}

	sub, err := l.api.RPC.State.SubscribeStorageRaw([]types.StorageKey{key})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case set := <-sub.Chan():
			log.GetLogger().Info("watch :", set.Block.Hex())
			for _, chng := range set.Changes {
				if !codec.Eq(chng.StorageKey, key) || !chng.HasStorageData {
					// skip, we are only interested in events with content
					continue
				}
				// Decode the event records
				evt := chain2.MyEventRecords{}
				storageData := chng.StorageData
				meta, err := l.api.RPC.State.GetMetadataLatest()
				if err != nil {
					log.GetLogger().Errorf("get metadata error: ", err)
					continue
				}
				err = types.EventRecordsRaw(storageData).DecodeEventRecords(meta, &evt)
				if err != nil {
					log.GetLogger().Error(err)
					continue
				}
				for _, e := range evt.ResourceOrder_CreateOrderSuccess {
					// order successfully created
					l.dealCreateOrderSuccess(e)
				}

				for _, e := range evt.ResourceOrder_ReNewOrderSuccess {
					// order renewal successful
					l.dealReNewOrderSuccess(e)
				}

				for _, e := range evt.ResourceOrder_WithdrawLockedOrderPriceSuccess {
					// order cancelled successfully
					log.GetLogger().Info("deal cancelOrder")
					l.dealCancelOrderSuccess(e.OrderIndex)
				}

				for _, e := range evt.ResourceOrder_CancelAgreementSuccess {
					// order cancelled successfully
					log.GetLogger().Info("deal agreement order")
					l.dealCancelOrderSuccess(e.OrderIndex)
				}

				for _, e := range evt.Provider_DeploymentDApp {
					log.GetLogger().Info("deal deployment dapp: ", e)
					cfg, err := l.cm.GetConfig()
					if err != nil {
						log.GetLogger().Errorf("get config error: ", err)
						continue
					}
					if e.PeerID != cfg.Identity.PeerID {
						log.GetLogger().Info("此 dapp 不归我部署，忽略")
						continue
					}
					app := demo.NewDApp(e.PeerID, uint8(e.CPU), uint8(e.Memory), uint8(e.StartMethod), e.Command)
					err = app.Start()
					if err != nil {
						log.GetLogger().Error("start dapp error: ", err)
					}
				}

				for _, e := range evt.Provider_ResourceHeartbeat {
					log.GetLogger().Info("resource heartbeat event: ", e)
				}

				for _, e := range evt.Provider_DAppHeartbeat {
					log.GetLogger().Info("dapp heartbeat event: ", e)
				}

				for _, e := range evt.Provider_DAppRedistribution {
					log.GetLogger().Info("dapp redistribution event: ", e)
				}

				for _, e := range evt.Provider_EndDAppSuccess {
					log.GetLogger().Info("end dapp success event: ", e)
				}

				for _, e := range evt.Provider_StopDApp {
					log.GetLogger().Info("stop dapp event: ", e)
				}
			}
		}
	}
}

func (l *ChainListener) dealCreateOrderSuccess(e chain2.EventResourceOrderCreateOrderSuccess) {
	cfg, err := l.cm.GetConfig()
	if err != nil {
		panic(err)
	}
	log.GetLogger().Infof("\tResourceOrder:CreateOrderSuccess:: (phase=%#v)\n", e.Phase)

	if e.ResourceIndex == types.NewU64(cfg.ChainRegInfo.ResourceIndex) {
		// process the order
		log.GetLogger().Info("deal order: ", e.OrderIndex)
		// record the id of the processed order
		cfg.ChainRegInfo.OrderIndex = uint64(e.OrderIndex)
		cfg.ChainRegInfo.AccountAddress = utils.AccountIdToAddress(e.AccountId)
		cfg.ChainRegInfo.DeployType = uint32(e.DeployType)
		_ = l.cm.Save(cfg)
		evt := &event.VmRequest{
			Tag:        event.OPCreatedVm,
			Cpu:        cfg.Vm.Cpu,
			Mem:        cfg.Vm.Mem,
			Disk:       cfg.Vm.Disk,
			OrderNo:    uint64(e.OrderIndex),
			System:     cfg.Vm.System,
			PublicKey:  e.PublicKey,
			Image:      cfg.Vm.Image,
			DeployType: uint32(e.DeployType),
		}
		l.eventService.Create(evt)

	} else {
		log.GetLogger().Warn("resourceIndex is not equals ")
	}
}

func (l *ChainListener) dealReNewOrderSuccess(e chain2.EventResourceOrderReNewOrderSuccess) {
	cfg, err := l.cm.GetConfig()
	if err != nil {
		panic(err)
	}
	if e.ResourceIndex == types.NewU64(cfg.ChainRegInfo.ResourceIndex) {
		evt := &event.VmRequest{
			Tag:     event.OPRenewVM,
			OrderNo: uint64(e.OrderIndex),
		}
		l.eventService.Renew(evt)
	}
}

func (l *ChainListener) dealCancelOrderSuccess(orderIndex types.U64) {
	cfg, err := l.cm.GetConfig()
	if err != nil {
		panic(err)
	}
	if orderIndex == types.NewU64(cfg.ChainRegInfo.OrderIndex) {
		evt := &event.VmRequest{
			Tag:     event.OPDestroyVm,
			Cpu:     cfg.Vm.Cpu,
			Mem:     cfg.Vm.Mem,
			Disk:    cfg.Vm.Disk,
			OrderNo: uint64(orderIndex),
			System:  cfg.Vm.System,
			Image:   cfg.Vm.Image,
		}
		l.eventService.Destroy(evt)
	}
}
