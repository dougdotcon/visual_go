package link

// Constantes para os registradores de comunicação
const (
	// Registradores de comunicação
	COMM_SIODATA32_L = 0x120 // Lower 16bit no modo Normal
	COMM_SIODATA32_H = 0x122 // Higher 16bit no modo Normal
	COMM_SIOCNT      = 0x128 // Registrador de controle
	COMM_SIODATA8    = 0x12a // 8bit no modo Normal/UART
	COMM_SIOMLT_SEND = 0x12a // SIOMLT_SEND (16bit R/W) no modo MultiPlayer (saída local)
	COMM_SIOMULTI0   = 0x120 // SIOMULTI0 (16bit) no modo MultiPlayer (Parent/Master)
	COMM_SIOMULTI1   = 0x122 // SIOMULTI1 (16bit) no modo MultiPlayer (Child1/Slave1)
	COMM_SIOMULTI2   = 0x124 // SIOMULTI2 (16bit) no modo MultiPlayer (Child2/Slave2)
	COMM_SIOMULTI3   = 0x126 // SIOMULTI3 (16bit) no modo MultiPlayer (Child3/Slave3)
	COMM_RCNT        = 0x134 // Modo SIO (4bit data) no modo GeneralPurpose
	COMM_IR          = 0x136 // Registrador Infravermelho (16bit)
	COMM_JOYCNT      = 0x140 // Registrador de controle JoyBus
	COMM_JOY_RECV_L  = 0x150 // Recebe 8bit Lower depois 8bit Higher
	COMM_JOY_RECV_H  = 0x152
	COMM_JOY_TRANS_L = 0x154 // Envia 8bit Lower depois 8bit Higher
	COMM_JOY_TRANS_H = 0x156
	COMM_JOYSTAT     = 0x158 // Envia/Recebe apenas 8bit lower
)

// Constantes para os bits de status do JoyBus
const (
	JOYSTAT_RECV = 2
	JOYSTAT_SEND = 8
)

// Constantes para os bits de controle do SIOCNT
const (
	SIO_INT_CLOCK               = 0x0001 // Clock interno
	SIO_INT_CLOCK_SEL_2MHZ      = 0x0002 // Seleção de clock 2MHz
	SIO_TRANS_FLAG_RECV_ENABLE  = 0x0004 // Habilita recepção
	SIO_TRANS_FLAG_SEND_DISABLE = 0x0008 // Desabilita envio
	SIO_TRANS_START             = 0x0080 // Inicia transferência
	SIO_TRANS_32BIT             = 0x1000 // Modo 32 bits
	SIO_IRQ_ENABLE              = 0x4000 // Habilita interrupção
)

// Modos de comunicação
type LinkMode int

const (
	LINK_DISCONNECTED LinkMode = iota
	LINK_CABLE_IPC
	LINK_RFU_IPC
	LINK_GAMEBOY_IPC
	LINK_CABLE_SOCKET
	LINK_RFU_SOCKET
	LINK_GAMECUBE_DOLPHIN
	LINK_GAMEBOY_SOCKET
)

// Estados da conexão
type ConnectionState int

const (
	LINK_OK ConnectionState = iota
	LINK_ERROR
	LINK_NEEDS_UPDATE
	LINK_ABORT
)

// Modos SIO
type SIOMode int

const (
	NORMAL32 SIOMode = iota
	NORMAL8
	MULTIPLAYER
	UART
	JOYBUS
	GP
)

// Link representa uma conexão de comunicação serial
type Link struct {
	mode            LinkMode
	state           ConnectionState
	enabled         bool
	transferring    bool
	receivedData    uint32
	transmitData    uint32
	numSlaves       int
	connectedSlaves int
	linkID          int
	speed           uint8
	transferStart   int64
	lastUpdate      int64
}

// NewLink cria uma nova instância de Link
func NewLink() *Link {
	return &Link{
		mode:            LINK_DISCONNECTED,
		state:           LINK_OK,
		enabled:         false,
		transferring:    false,
		receivedData:    0,
		transmitData:    0,
		numSlaves:       0,
		connectedSlaves: 0,
		linkID:          0,
		speed:           3,
		transferStart:   0,
		lastUpdate:      0,
	}
}

// GetSIOMode retorna o modo SIO atual baseado nos registradores SIOCNT e RCNT
func GetSIOMode(siocnt, rcnt uint16) SIOMode {
	if rcnt&0x8000 != 0 {
		if rcnt&0x4000 != 0 {
			return JOYBUS
		}
		return GP
	}

	if siocnt&SIO_TRANS_32BIT != 0 {
		return NORMAL32
	}

	if siocnt&0x2000 != 0 {
		return MULTIPLAYER
	}

	if siocnt&0x3000 == 0x3000 {
		return UART
	}

	return NORMAL8
}

// InitLink inicializa a conexão no modo especificado
func (l *Link) InitLink(mode LinkMode) ConnectionState {
	if mode == LINK_DISCONNECTED {
		return LINK_ABORT
	}

	if l.mode != LINK_DISCONNECTED {
		return LINK_ERROR
	}

	l.mode = mode
	l.state = LINK_OK
	l.enabled = true

	return l.state
}

// StartLink inicia uma transferência de dados
func (l *Link) StartLink(siocnt uint16) {
	if !l.enabled {
		return
	}

	mode := GetSIOMode(siocnt, 0) // TODO: Passar RCNT como parâmetro
	switch mode {
	case MULTIPLAYER:
		if l.linkID == 0 && siocnt&SIO_TRANS_START != 0 && !l.transferring {
			l.transferStart = l.lastUpdate
			l.transferring = true
		}
	case NORMAL8, NORMAL32, UART:
		if siocnt&SIO_TRANS_START != 0 {
			l.transferStart = l.lastUpdate
			l.transferring = true
		}
	}
}

// Update atualiza o estado da conexão
func (l *Link) Update(ticks int64) {
	l.lastUpdate += ticks

	if !l.enabled || !l.transferring {
		return
	}

	// TODO: Implementar lógica de atualização baseada no modo
}
