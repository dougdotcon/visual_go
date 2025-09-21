package link

import (
	"time"
)

// Constantes para comandos JoyBus
const (
	JOY_CMD_RESET  = 0xff
	JOY_CMD_STATUS = 0x00
	JOY_CMD_READ   = 0x14
	JOY_CMD_WRITE  = 0x15
)

// Constantes para bits de controle JoyBus
const (
	JOYCNT_RESET         = 0x01
	JOYCNT_RECV_COMPLETE = 0x02
	JOYCNT_SEND_COMPLETE = 0x04
	JOYCNT_INT_ENABLE    = 0x40
)

// JoyBusLink representa uma conexão via JoyBus
type JoyBusLink struct {
	*Link
	clockSync      int64
	lastCommand    int64
	lastUpdate     int64
	nextUpdate     int64
	booted         bool
	isDisconnected bool
}

// NewJoyBusLink cria uma nova instância de JoyBusLink
func NewJoyBusLink() *JoyBusLink {
	return &JoyBusLink{
		Link:           NewLink(),
		clockSync:      0,
		lastCommand:    0,
		lastUpdate:     0,
		nextUpdate:     0,
		booted:         false,
		isDisconnected: false,
	}
}

// Connect estabelece a conexão JoyBus
func (j *JoyBusLink) Connect() ConnectionState {
	j.mode = LINK_GAMECUBE_DOLPHIN
	j.state = LINK_OK
	j.enabled = true
	j.booted = false
	j.isDisconnected = false
	return LINK_OK
}

// Disconnect desconecta a conexão JoyBus
func (j *JoyBusLink) Disconnect() {
	j.isDisconnected = true
	j.enabled = false
	j.booted = false
}

// IsDisconnected retorna se a conexão está desconectada
func (j *JoyBusLink) IsDisconnected() bool {
	return j.isDisconnected
}

// ClockSync sincroniza o clock
func (j *JoyBusLink) ClockSync(ticks int64) {
	j.clockSync += ticks
}

// ReceiveCommand recebe um comando JoyBus
func (j *JoyBusLink) ReceiveCommand(data []byte, block bool) (uint8, error) {
	if j.isDisconnected {
		return data[0], nil
	}

	// Se block for true, espera até receber dados
	if block {
		timeout := time.Now().Add(6 * time.Second)
		for time.Now().Before(timeout) {
			if j.clockSync > 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// TODO: Implementar recebimento real de comandos
	// Por enquanto, retorna o primeiro byte dos dados
	return data[0], nil
}

// SendResponse envia uma resposta JoyBus
func (j *JoyBusLink) SendResponse(data []byte) error {
	if j.isDisconnected {
		return nil
	}

	// TODO: Implementar envio real de respostas
	return nil
}

// UpdateJoyBus atualiza o estado da conexão JoyBus
func (j *JoyBusLink) UpdateJoyBus(ticks int64) {
	j.lastUpdate += ticks
	j.lastCommand += ticks

	if !j.enabled || j.isDisconnected {
		return
	}

	// Verifica se é hora de atualizar
	if j.lastUpdate <= j.nextUpdate {
		return
	}

	// Recebe comando
	data := make([]byte, 5)
	cmd, err := j.ReceiveCommand(data, j.lastCommand > 4*16780000) // 4 frames em ticks
	if err != nil {
		j.Disconnect()
		return
	}

	// Processa comando
	var response []byte
	switch cmd {
	case JOY_CMD_RESET:
		response = []byte{0x00, 0x04}   // ID do dispositivo GBA
		j.nextUpdate = 16780000 / 38400 // 1 segundo / 38400 bps

	case JOY_CMD_STATUS:
		response = []byte{0x00, 0x04} // ID do dispositivo GBA
		j.nextUpdate = 16780000 / 38400

	case JOY_CMD_READ:
		// TODO: Implementar leitura de dados
		response = make([]byte, 4)
		j.nextUpdate = 16780000 / 38400
		j.booted = true

	case JOY_CMD_WRITE:
		// TODO: Implementar escrita de dados
		j.nextUpdate = 16780000 / 38400
		j.booted = true

	default:
		j.nextUpdate = 16780000 / 40000
		j.lastUpdate = 0
		return
	}

	// Envia resposta
	j.SendResponse(response)
	j.lastUpdate = 0
}
