package savestate

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// Constantes do save state
const (
	SaveStateVersion = 1
	SaveStateMagic   = "VBGO" // VisualBoy Go
)

// SaveState representa um estado salvo do emulador
type SaveState struct {
	// Header
	Magic     [4]byte
	Version   uint32
	Timestamp int64
	ROMTitle  [16]byte

	// Estado do CPU
	CPU CPUState

	// Estado da memória
	Memory MemoryState

	// Estado do LCD
	LCD LCDState

	// Estado do Timer
	Timer TimerState

	// Estado do Input
	Input InputState

	// Estado do Sound
	Sound SoundState

	// Estado das Interrupções
	Interrupts InterruptState
}

// CPUState representa o estado do CPU
type CPUState struct {
	// Registradores
	A, B, C, D, E, H, L uint8
	SP, PC              uint16

	// Flags
	FlagZ, FlagN, FlagH, FlagC bool

	// Estado
	Halted            bool
	InterruptsEnabled bool
}

// MemoryState representa o estado da memória
type MemoryState struct {
	// Work RAM
	WRAM [0x2000]uint8

	// High RAM
	HRAM [0x7F]uint8

	// Estado do cartucho
	CurrentROMBank uint16
	CurrentRAMBank uint16
	RAMEnabled     uint8 // 0 ou 1
	MBCMode        uint16

	// RAM externa (tamanho fixo para serialização)
	ExternalRAMSize uint16
	ExternalRAM     [0x8000]uint8 // Máximo 32KB
}

// LCDState representa o estado do LCD
type LCDState struct {
	// Registradores
	LCDC, STAT, SCY, SCX, LY, LYC uint8
	BGP, OBP0, OBP1               uint8
	WY, WX                        uint8

	// Estado interno
	Mode   uint8
	Cycles uint32

	// VRAM
	VRAM [0x2000]uint8

	// OAM
	OAM [0xA0]uint8
}

// TimerState representa o estado do timer
type TimerState struct {
	// Registradores
	DIV, TIMA, TMA, TAC uint8

	// Contadores internos
	DIVCounter  uint32
	TIMACounter uint32
}

// InputState representa o estado do input
type InputState struct {
	// Registrador
	JOYP uint8

	// Estado dos botões
	Buttons [8]bool
}

// SoundState representa o estado do som
type SoundState struct {
	// Registradores principais
	NR50, NR51, NR52 uint8

	// Wave RAM
	WaveRAM [16]uint8

	// Estado interno
	FrameSequencer uint32
	Cycles         uint32
}

// InterruptState representa o estado das interrupções
type InterruptState struct {
	// Registradores
	InterruptFlag   uint8
	InterruptEnable uint8

	// Estado
	MasterEnable bool
}

// NewSaveState cria um novo save state vazio
func NewSaveState() *SaveState {
	ss := &SaveState{}
	copy(ss.Magic[:], SaveStateMagic)
	ss.Version = SaveStateVersion
	ss.Timestamp = time.Now().Unix()
	return ss
}

// Serialize serializa o save state para bytes
func (ss *SaveState) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Escreve o save state usando binary encoding
	err := binary.Write(&buf, binary.LittleEndian, ss)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar save state: %w", err)
	}

	return buf.Bytes(), nil
}

// Deserialize deserializa bytes para um save state
func Deserialize(data []byte) (*SaveState, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("dados de save state muito pequenos")
	}

	ss := &SaveState{}
	buf := bytes.NewReader(data)

	// Lê o save state usando binary encoding
	err := binary.Read(buf, binary.LittleEndian, ss)
	if err != nil {
		return nil, fmt.Errorf("erro ao deserializar save state: %w", err)
	}

	// Verifica magic e versão
	if string(ss.Magic[:]) != SaveStateMagic {
		return nil, fmt.Errorf("magic inválido no save state")
	}

	if ss.Version != SaveStateVersion {
		return nil, fmt.Errorf("versão de save state não suportada: %d", ss.Version)
	}

	return ss, nil
}

// GetTimestamp retorna o timestamp do save state
func (ss *SaveState) GetTimestamp() time.Time {
	return time.Unix(ss.Timestamp, 0)
}

// GetROMTitle retorna o título da ROM
func (ss *SaveState) GetROMTitle() string {
	// Remove bytes nulos do final
	title := ss.ROMTitle[:]
	for i := len(title) - 1; i >= 0; i-- {
		if title[i] != 0 {
			return string(title[:i+1])
		}
	}
	return ""
}

// SetROMTitle define o título da ROM
func (ss *SaveState) SetROMTitle(title string) {
	// Limpa o array
	for i := range ss.ROMTitle {
		ss.ROMTitle[i] = 0
	}

	// Copia o título (máximo 16 bytes)
	copy(ss.ROMTitle[:], []byte(title))
}

// Validate valida a integridade do save state
func (ss *SaveState) Validate() error {
	// Verifica magic
	if string(ss.Magic[:]) != SaveStateMagic {
		return fmt.Errorf("magic inválido: %s", string(ss.Magic[:]))
	}

	// Verifica versão
	if ss.Version != SaveStateVersion {
		return fmt.Errorf("versão não suportada: %d", ss.Version)
	}

	// Verifica timestamp
	if ss.Timestamp <= 0 {
		return fmt.Errorf("timestamp inválido: %d", ss.Timestamp)
	}

	// Verifica registradores do CPU
	if ss.CPU.SP > 0xFFFF || ss.CPU.PC > 0xFFFF {
		return fmt.Errorf("registradores de CPU inválidos")
	}

	// Verifica estado do LCD
	if ss.LCD.Mode > 3 {
		return fmt.Errorf("modo LCD inválido: %d", ss.LCD.Mode)
	}

	return nil
}

// GetSize retorna o tamanho do save state em bytes
func (ss *SaveState) GetSize() int {
	data, err := ss.Serialize()
	if err != nil {
		return 0
	}
	return len(data)
}

// Clone cria uma cópia do save state
func (ss *SaveState) Clone() (*SaveState, error) {
	data, err := ss.Serialize()
	if err != nil {
		return nil, fmt.Errorf("erro ao clonar save state: %w", err)
	}

	return Deserialize(data)
}

// String retorna uma representação em string do save state
func (ss *SaveState) String() string {
	return fmt.Sprintf("SaveState{Version: %d, ROM: %s, Time: %s, Size: %d bytes}",
		ss.Version, ss.GetROMTitle(), ss.GetTimestamp().Format("2006-01-02 15:04:05"), ss.GetSize())
}

// SaveStateManager gerencia múltiplos save states
type SaveStateManager struct {
	slots map[int]*SaveState
}

// NewSaveStateManager cria um novo gerenciador de save states
func NewSaveStateManager() *SaveStateManager {
	return &SaveStateManager{
		slots: make(map[int]*SaveState),
	}
}

// SaveToSlot salva um save state em um slot
func (ssm *SaveStateManager) SaveToSlot(slot int, saveState *SaveState) error {
	if slot < 0 || slot > 9 {
		return fmt.Errorf("slot inválido: %d (deve ser 0-9)", slot)
	}

	if err := saveState.Validate(); err != nil {
		return fmt.Errorf("save state inválido: %w", err)
	}

	// Clona o save state para evitar modificações externas
	cloned, err := saveState.Clone()
	if err != nil {
		return fmt.Errorf("erro ao clonar save state: %w", err)
	}

	ssm.slots[slot] = cloned
	return nil
}

// LoadFromSlot carrega um save state de um slot
func (ssm *SaveStateManager) LoadFromSlot(slot int) (*SaveState, error) {
	if slot < 0 || slot > 9 {
		return nil, fmt.Errorf("slot inválido: %d (deve ser 0-9)", slot)
	}

	saveState, exists := ssm.slots[slot]
	if !exists {
		return nil, fmt.Errorf("slot %d está vazio", slot)
	}

	// Clona o save state para evitar modificações externas
	return saveState.Clone()
}

// HasSlot verifica se um slot contém um save state
func (ssm *SaveStateManager) HasSlot(slot int) bool {
	if slot < 0 || slot > 9 {
		return false
	}

	_, exists := ssm.slots[slot]
	return exists
}

// ClearSlot limpa um slot
func (ssm *SaveStateManager) ClearSlot(slot int) {
	if slot >= 0 && slot <= 9 {
		delete(ssm.slots, slot)
	}
}

// GetUsedSlots retorna uma lista dos slots em uso
func (ssm *SaveStateManager) GetUsedSlots() []int {
	var slots []int
	for slot := range ssm.slots {
		slots = append(slots, slot)
	}
	return slots
}

// GetSlotInfo retorna informações sobre um slot
func (ssm *SaveStateManager) GetSlotInfo(slot int) (string, error) {
	saveState, err := ssm.LoadFromSlot(slot)
	if err != nil {
		return "", err
	}

	return saveState.String(), nil
}
