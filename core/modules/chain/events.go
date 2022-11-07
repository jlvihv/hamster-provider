package chain

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type Provider_RegisterResourceSuccess struct {
	Phase         types.Phase
	AccountID     types.AccountID
	ResourceIndex types.U64
	CPU           types.U8
	Memory        types.U8
	Topics        []types.Hash
}
type Provider_DeploymentDApp struct {
	Phase       types.Phase
	PeerID      string
	CPU         types.U8
	Memory      types.U8
	StartMethod types.U8
	Command     string
	DAppIndex   types.U64
	Topics      []types.Hash
}

type EventProviderRegisterResourceSuccess struct {
	Phase            types.Phase
	Index            types.U64
	PeerId           string
	Cpu              types.U64
	Memory           types.U64
	System           string
	CpuModel         string
	PriceHour        types.U128
	RentDurationHour types.U32
	PublicIP         string
	Specification    types.U32
	Topics           []types.Hash
}

type EventResourceOrderCreateOrderSuccess struct {
	Phase         types.Phase
	AccountId     types.AccountID
	OrderIndex    types.U64
	ResourceIndex types.U64
	Duration      types.U32
	DeployType    types.U32
	PublicKey     string
	Topics        []types.Hash
}

type EventResourceOrderOrderExecSuccess struct {
	Phase          types.Phase
	AccountId      types.AccountID
	OrderIndex     types.U64
	ResourceIndex  types.U64
	AgreementIndex types.U64
	Topics         []types.Hash
}

type EventResourceOrderReNewOrderSuccess struct {
	Phase          types.Phase
	AccountId      types.AccountID
	OrderIndex     types.U64
	ResourceIndex  types.U64
	AgreementIndex types.U64
	Topics         []types.Hash
}

type EventResourceOrderWithdrawLockedOrderPriceSuccess struct {
	Phase      types.Phase
	AccountId  types.AccountID
	OrderIndex types.U64
	Topics     []types.Hash
}

type EventMarketMoney struct {
	Phase  types.Phase
	Money  types.U128
	Topics []types.Hash
}

type EventCancelAgreementSuccess struct {
	Phase          types.Phase
	AccountId      types.AccountID
	AgreementIndex types.U64
	OrderIndex     types.U64
	Topics         []types.Hash
}

type MyEventRecords struct {
	types.EventRecords
	Provider_DeploymentDApp          []Provider_DeploymentDApp          //nolint:stylecheck,golint
	Provider_RegisterResourceSuccess []Provider_RegisterResourceSuccess //nolint:stylecheck,golint
	Provider_ResourceHeartbeat       []Provider_ResourceHeartbeat       //nolint:stylecheck,golint
	Provider_DAppHeartbeat           []Provider_DAppHeartbeat           //nolint:stylecheck,golint
	Provider_DAppRedistribution      []Provider_DAppRedistribution      //nolint:stylecheck,golint
	Provider_EndDAppSuccess          []Provider_EndDAppSuccess          //nolint:stylecheck,golint
	Provider_StopDApp                []Provider_StopDApp                //nolint:stylecheck,golint
	// Provider_RegisterResourceSuccess              []EventProviderRegisterResourceSuccess //nolint:stylecheck,golint
	ResourceOrder_CreateOrderSuccess              []EventResourceOrderCreateOrderSuccess //nolint:stylecheck,golint
	ResourceOrder_OrderExecSuccess                []EventResourceOrderOrderExecSuccess
	ResourceOrder_ReNewOrderSuccess               []EventResourceOrderReNewOrderSuccess
	ResourceOrder_WithdrawLockedOrderPriceSuccess []EventResourceOrderWithdrawLockedOrderPriceSuccess
	ResourceOrder_CancelAgreementSuccess          []EventCancelAgreementSuccess
}

type Provider_ResourceHeartbeat struct {
	Phase  types.Phase
	PeerID string
	DApps  []types.U64
	Topics []types.Hash
}

type Provider_DAppHeartbeat struct {
	Phase     types.Phase
	AccountId types.AccountID
	DAppName  string
	Topics    []types.Hash
}

type Provider_DAppRedistribution struct {
	Phase  types.Phase
	DApps  []string
	Topics []types.Hash
}
type Provider_EndDAppSuccess struct {
	Phase     types.Phase
	AccountId types.AccountID
	DAppName  string
	DAppIndex types.U64
	Topics    []types.Hash
}

type Provider_StopDApp struct {
	Phase     types.Phase
	PeerID    string
	DAppIndex types.U64
	Topics    []types.Hash
}
