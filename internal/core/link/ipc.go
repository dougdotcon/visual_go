package link

import (
	"os"
	"sync"
	"unsafe"
)

// Constantes para IPC
const (
	LOCAL_LINK_NAME = "VBA link memory"
	LINK_EVENT_NAME = "VBA link event"
)

// LinkMemory representa a estrutura de memória compartilhada
type LinkMemory struct {
	linkCmd  [4]int32
	linkData [4]uint32
}

// IPCLink representa uma conexão via IPC
type IPCLink struct {
	*Link
	linkMemory *LinkMemory
	memFile    *os.File
	mutex      sync.Mutex
	events     [4]uintptr
}

// NewIPCLink cria uma nova instância de IPCLink
func NewIPCLink() *IPCLink {
	return &IPCLink{
		Link: NewLink(),
	}
}

// Connect estabelece a conexão IPC
func (i *IPCLink) Connect() ConnectionState {
	var err error

	// Cria ou abre arquivo de memória compartilhada
	i.memFile, err = os.OpenFile(LOCAL_LINK_NAME, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return LINK_ERROR
	}

	// Mapeia memória compartilhada
	size := unsafe.Sizeof(LinkMemory{})
	data, err := mapSharedMemory(i.memFile, int(size))
	if err != nil {
		i.memFile.Close()
		return LINK_ERROR
	}

	i.linkMemory = (*LinkMemory)(unsafe.Pointer(&data[0]))

	// Cria eventos para sincronização
	for j := 0; j < 4; j++ {
		eventName := LINK_EVENT_NAME + string(rune('1'+j))
		event, err := createEvent(eventName)
		if err != nil {
			i.Close()
			return LINK_ERROR
		}
		i.events[j] = event
	}

	i.enabled = true
	i.state = LINK_OK
	return LINK_OK
}

// Close fecha a conexão IPC
func (i *IPCLink) Close() {
	if i.linkMemory != nil {
		size := unsafe.Sizeof(LinkMemory{})
		unmapSharedMemory((*(*[1<<31 - 1]byte)(unsafe.Pointer(i.linkMemory)))[:size])
		i.linkMemory = nil
	}

	if i.memFile != nil {
		i.memFile.Close()
		i.memFile = nil
	}

	for j := 0; j < 4; j++ {
		if i.events[j] != 0 {
			closeEvent(i.events[j])
			i.events[j] = 0
		}
	}

	i.enabled = false
}

// SendData envia dados via IPC
func (i *IPCLink) SendData(data uint32) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if !i.enabled || i.linkMemory == nil {
		return nil
	}

	// Espera até que o comando anterior seja processado
	waitForEvent(i.events[i.linkID])

	// Envia dados
	i.linkMemory.linkData[i.linkID] = data
	i.linkMemory.linkCmd[i.linkID] = 1

	// Sinaliza evento
	setEvent(i.events[i.linkID])

	return nil
}

// ReceiveData recebe dados via IPC
func (i *IPCLink) ReceiveData() (uint32, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if !i.enabled || i.linkMemory == nil {
		return 0, nil
	}

	// Espera por dados
	waitForEvent(i.events[1-i.linkID])

	// Lê dados
	data := i.linkMemory.linkData[1-i.linkID]
	i.linkMemory.linkCmd[1-i.linkID] = 0

	// Sinaliza evento
	setEvent(i.events[1-i.linkID])

	return data, nil
}

// UpdateIPC atualiza o estado da conexão IPC
func (i *IPCLink) UpdateIPC(ticks int64) {
	i.lastUpdate += ticks

	if !i.enabled || !i.transferring {
		return
	}

	// TODO: Implementar lógica de atualização específica para IPC
	// - Verificar dados recebidos
	// - Atualizar estado da transferência
	// - Gerenciar sincronização
}
