package savestate

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SaveState representa um estado salvo do emulador
type SaveState struct {
	// Metadados
	Version     int       `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	ROMName     string    `json:"rom_name"`
	ROMHash     string    `json:"rom_hash"`

	// Estado do emulador
	CPU struct {
		Registers [16]uint32 `json:"registers"`
		CPSR      uint32     `json:"cpsr"`
		SPSR      uint32     `json:"spsr"`
		ThumbMode bool       `json:"thumb_mode"`
		Halted    bool       `json:"halted"`
		Cycles    uint64     `json:"cycles"`
	} `json:"cpu"`

	Memory struct {
		BIOS    []byte `json:"bios"`
		EWRAM   []byte `json:"ewram"`
		IWRAM   []byte `json:"iwram"`
		IO      []byte `json:"io"`
		Palette []byte `json:"palette"`
		VRAM    []byte `json:"vram"`
		OAM     []byte `json:"oam"`
		ROM     []byte `json:"rom"`
		Save    []byte `json:"save"`
	} `json:"memory"`

	GPU struct {
		Mode        int    `json:"mode"`
		VCount      uint16 `json:"vcount"`
		Framebuffer []byte `json:"framebuffer"`
	} `json:"gpu"`

	APU struct {
		PSG struct {
			Channel1 []byte `json:"channel1"`
			Channel2 []byte `json:"channel2"`
			Channel3 []byte `json:"channel3"`
			Channel4 []byte `json:"channel4"`
		} `json:"psg"`
		DirectSound struct {
			ChannelA []byte `json:"channel_a"`
			ChannelB []byte `json:"channel_b"`
		} `json:"direct_sound"`
		FIFO []byte `json:"fifo"`
	} `json:"apu"`

	DMA struct {
		Channel [4]struct {
			Source      uint32 `json:"source"`
			Destination uint32 `json:"destination"`
			Count       uint16 `json:"count"`
			Control     uint16 `json:"control"`
		} `json:"channels"`
	} `json:"dma"`

	Timer struct {
		Channel [4]struct {
			Counter  uint16 `json:"counter"`
			Reload   uint16 `json:"reload"`
			Control  uint16 `json:"control"`
			Overflow bool   `json:"overflow"`
		} `json:"channels"`
	} `json:"timer"`
}

// SaveStateManager gerencia os estados salvos do emulador
type SaveStateManager struct {
	savePath string
	slots    int
	current  *SaveState
}

// NewSaveStateManager cria uma nova instância do gerenciador de estados
func NewSaveStateManager(savePath string, slots int) *SaveStateManager {
	return &SaveStateManager{
		savePath: savePath,
		slots:    slots,
	}
}

// SaveToSlot salva o estado atual em um slot específico
func (sm *SaveStateManager) SaveToSlot(slot int, state *SaveState) error {
	if slot < 0 || slot >= sm.slots {
		return fmt.Errorf("slot inválido: %d", slot)
	}

	// Cria o diretório se não existir
	if err := os.MkdirAll(sm.savePath, 0755); err != nil {
		return err
	}

	// Prepara o arquivo
	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Comprime os dados
	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	// Serializa o estado
	enc := gob.NewEncoder(gzw)
	if err := enc.Encode(state); err != nil {
		return err
	}

	sm.current = state
	return nil
}

// LoadFromSlot carrega um estado de um slot específico
func (sm *SaveStateManager) LoadFromSlot(slot int) (*SaveState, error) {
	if slot < 0 || slot >= sm.slots {
		return nil, fmt.Errorf("slot inválido: %d", slot)
	}

	// Abre o arquivo
	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Descomprime os dados
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	// Deserializa o estado
	var state SaveState
	dec := gob.NewDecoder(gzr)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}

	sm.current = &state
	return &state, nil
}

// GetSlotInfo retorna informações sobre um slot específico
func (sm *SaveStateManager) GetSlotInfo(slot int) (*SaveStateInfo, error) {
	if slot < 0 || slot >= sm.slots {
		return nil, fmt.Errorf("slot inválido: %d", slot)
	}

	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &SaveStateInfo{
				Slot:     slot,
				Empty:    true,
				Filename: filename,
			}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Lê apenas os metadados
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	var state SaveState
	dec := gob.NewDecoder(gzr)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}

	return &SaveStateInfo{
		Slot:        slot,
		Empty:       false,
		Filename:    filename,
		Timestamp:   state.Timestamp,
		Description: state.Description,
		ROMName:     state.ROMName,
	}, nil
}

// SaveStateInfo contém informações sobre um slot de estado salvo
type SaveStateInfo struct {
	Slot        int
	Empty       bool
	Filename    string
	Timestamp   time.Time
	Description string
	ROMName     string
}

// GetCurrentState retorna o estado atual
func (sm *SaveStateManager) GetCurrentState() *SaveState {
	return sm.current
}

// SetCurrentState define o estado atual
func (sm *SaveStateManager) SetCurrentState(state *SaveState) {
	sm.current = state
}

// SaveToFile salva o estado atual em um arquivo específico
func (sm *SaveStateManager) SaveToFile(filename string, state *SaveState) error {
	// Cria o diretório se não existir
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Prepara o arquivo
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Comprime os dados
	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	// Serializa o estado
	enc := gob.NewEncoder(gzw)
	if err := enc.Encode(state); err != nil {
		return err
	}

	sm.current = state
	return nil
}

// LoadFromFile carrega um estado de um arquivo específico
func (sm *SaveStateManager) LoadFromFile(filename string) (*SaveState, error) {
	// Abre o arquivo
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Descomprime os dados
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	// Deserializa o estado
	var state SaveState
	dec := gob.NewDecoder(gzr)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}

	sm.current = &state
	return &state, nil
}

// SaveToBuffer salva o estado atual em um buffer de memória
func (sm *SaveStateManager) SaveToBuffer(state *SaveState) ([]byte, error) {
	var buf bytes.Buffer

	// Comprime os dados
	gzw := gzip.NewWriter(&buf)

	// Serializa o estado
	enc := gob.NewEncoder(gzw)
	if err := enc.Encode(state); err != nil {
		return nil, err
	}

	if err := gzw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// LoadFromBuffer carrega um estado de um buffer de memória
func (sm *SaveStateManager) LoadFromBuffer(data []byte) (*SaveState, error) {
	buf := bytes.NewReader(data)

	// Descomprime os dados
	gzr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	// Deserializa o estado
	var state SaveState
	dec := gob.NewDecoder(gzr)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}

	sm.current = &state
	return &state, nil
}

// DeleteSlot remove um estado salvo de um slot específico
func (sm *SaveStateManager) DeleteSlot(slot int) error {
	if slot < 0 || slot >= sm.slots {
		return fmt.Errorf("slot inválido: %d", slot)
	}

	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	return os.Remove(filename)
}

// GetSlotCount retorna o número total de slots disponíveis
func (sm *SaveStateManager) GetSlotCount() int {
	return sm.slots
}

// GetSavePath retorna o diretório onde os estados são salvos
func (sm *SaveStateManager) GetSavePath() string {
	return sm.savePath
}

// SetSavePath define o diretório onde os estados são salvos
func (sm *SaveStateManager) SetSavePath(path string) {
	sm.savePath = path
}

// ValidateState verifica se um estado é válido para o ROM atual
func (sm *SaveStateManager) ValidateState(state *SaveState, romName, romHash string) bool {
	return state.ROMName == romName && state.ROMHash == romHash
}

// CopyState cria uma cópia profunda de um estado
func (sm *SaveStateManager) CopyState(state *SaveState) (*SaveState, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(state); err != nil {
		return nil, err
	}

	var copy SaveState
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&copy); err != nil {
		return nil, err
	}

	return &copy, nil
}

// GetStateSize retorna o tamanho em bytes de um estado salvo
func (sm *SaveStateManager) GetStateSize(state *SaveState) (int64, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)

	enc := gob.NewEncoder(gzw)
	if err := enc.Encode(state); err != nil {
		return 0, err
	}

	if err := gzw.Close(); err != nil {
		return 0, err
	}

	return int64(buf.Len()), nil
}

// GetSlotFileSize retorna o tamanho em bytes do arquivo de um slot
func (sm *SaveStateManager) GetSlotFileSize(slot int) (int64, error) {
	if slot < 0 || slot >= sm.slots {
		return 0, fmt.Errorf("slot inválido: %d", slot)
	}

	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	return info.Size(), nil
}

// GetSlotModTime retorna a data de modificação do arquivo de um slot
func (sm *SaveStateManager) GetSlotModTime(slot int) (time.Time, error) {
	if slot < 0 || slot >= sm.slots {
		return time.Time{}, fmt.Errorf("slot inválido: %d", slot)
	}

	filename := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", slot))
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

// CopySlot copia o estado de um slot para outro
func (sm *SaveStateManager) CopySlot(srcSlot, dstSlot int) error {
	if srcSlot < 0 || srcSlot >= sm.slots || dstSlot < 0 || dstSlot >= sm.slots {
		return fmt.Errorf("slot inválido")
	}

	srcFile := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", srcSlot))
	dstFile := filepath.Join(sm.savePath, fmt.Sprintf("slot_%d.sav", dstSlot))

	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
