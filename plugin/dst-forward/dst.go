package dstforward

func Init() {
	RegisterCustomLogic()
	go registerServer()
}
