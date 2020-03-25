package handler

type HandlerSet interface {
	Init()
}

func InitAllHTTPHandlers() {
	var h *StaticHandler
	h.Init()
}
