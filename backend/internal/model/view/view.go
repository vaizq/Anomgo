package view

type Views struct {
	Product ProductView
	Order   OrderView
	Invoice InvoiceView
	Review  ReviewView
	Dispute DisputeView
	Vendor  VendorView
	Ticket  TicketView
}

var V Views
