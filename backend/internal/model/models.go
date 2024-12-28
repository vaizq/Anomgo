package model

type Models struct {
	User            UserModel
	Product         ProductModel
	Price           PriceModel
	Order           OrderModel
	Review          ReviewModel
	Invoice         InvoiceModel
	Wallet          WalletModel
	Withdrawal      WithdrawalModel
	Transaction     TransactionMonel
	Dispute         DisputeModel
	CounterDispute  CounterDisputeModel
	DisputeDecision DisputeDecisionModel
	VendorPledge    VendorPledgeModel
	DeliveryMethod  DeliveryMethodModel
	DeclineReason   DeclineReasonModel
	DeliveryInfo    DeliveryInfoModel
	Ticket          TicketModel
	TicketResponse  TicketResponseModel
	Ban             BanModel
}

var M Models
