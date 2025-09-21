package savestate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveStateManager(t *testing.T) {
	// Cria um diretório temporário para os testes
	tempDir, err := os.MkdirTemp("", "savestate_test")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria um gerenciador de estados
	sm := NewSaveStateManager(tempDir, 10)

	// Cria um estado de teste
	state := &SaveState{
		Version:     1,
		Timestamp:   time.Now(),
		Description: "Estado de teste",
		ROMName:     "test.gba",
		ROMHash:     "123456789abcdef",
	}

	// Testa SaveToSlot e LoadFromSlot
	t.Run("SaveLoad", func(t *testing.T) {
		// Salva o estado no slot 0
		if err := sm.SaveToSlot(0, state); err != nil {
			t.Errorf("Erro ao salvar estado: %v", err)
		}

		// Carrega o estado do slot 0
		loaded, err := sm.LoadFromSlot(0)
		if err != nil {
			t.Errorf("Erro ao carregar estado: %v", err)
		}

		// Verifica se os dados foram preservados
		if loaded.Version != state.Version {
			t.Errorf("Versão incorreta: esperado %d, obtido %d", state.Version, loaded.Version)
		}
		if loaded.ROMName != state.ROMName {
			t.Errorf("Nome da ROM incorreto: esperado %s, obtido %s", state.ROMName, loaded.ROMName)
		}
		if loaded.ROMHash != state.ROMHash {
			t.Errorf("Hash da ROM incorreto: esperado %s, obtido %s", state.ROMHash, loaded.ROMHash)
		}
	})

	// Testa GetSlotInfo
	t.Run("SlotInfo", func(t *testing.T) {
		info, err := sm.GetSlotInfo(0)
		if err != nil {
			t.Errorf("Erro ao obter informações do slot: %v", err)
		}

		if info.Empty {
			t.Error("Slot deveria estar ocupado")
		}
		if info.ROMName != state.ROMName {
			t.Errorf("Nome da ROM incorreto: esperado %s, obtido %s", state.ROMName, info.ROMName)
		}
	})

	// Testa DeleteSlot
	t.Run("Delete", func(t *testing.T) {
		if err := sm.DeleteSlot(0); err != nil {
			t.Errorf("Erro ao deletar slot: %v", err)
		}

		info, err := sm.GetSlotInfo(0)
		if err != nil {
			t.Errorf("Erro ao obter informações do slot: %v", err)
		}

		if !info.Empty {
			t.Error("Slot deveria estar vazio")
		}
	})

	// Testa SaveToFile e LoadFromFile
	t.Run("SaveLoadFile", func(t *testing.T) {
		filename := filepath.Join(tempDir, "test.sav")

		// Salva o estado em um arquivo
		if err := sm.SaveToFile(filename, state); err != nil {
			t.Errorf("Erro ao salvar arquivo: %v", err)
		}

		// Carrega o estado do arquivo
		loaded, err := sm.LoadFromFile(filename)
		if err != nil {
			t.Errorf("Erro ao carregar arquivo: %v", err)
		}

		// Verifica se os dados foram preservados
		if loaded.Version != state.Version {
			t.Errorf("Versão incorreta: esperado %d, obtido %d", state.Version, loaded.Version)
		}
		if loaded.ROMName != state.ROMName {
			t.Errorf("Nome da ROM incorreto: esperado %s, obtido %s", state.ROMName, loaded.ROMName)
		}
		if loaded.ROMHash != state.ROMHash {
			t.Errorf("Hash da ROM incorreto: esperado %s, obtido %s", state.ROMHash, loaded.ROMHash)
		}
	})

	// Testa SaveToBuffer e LoadFromBuffer
	t.Run("SaveLoadBuffer", func(t *testing.T) {
		// Salva o estado em um buffer
		data, err := sm.SaveToBuffer(state)
		if err != nil {
			t.Errorf("Erro ao salvar buffer: %v", err)
		}

		// Carrega o estado do buffer
		loaded, err := sm.LoadFromBuffer(data)
		if err != nil {
			t.Errorf("Erro ao carregar buffer: %v", err)
		}

		// Verifica se os dados foram preservados
		if loaded.Version != state.Version {
			t.Errorf("Versão incorreta: esperado %d, obtido %d", state.Version, loaded.Version)
		}
		if loaded.ROMName != state.ROMName {
			t.Errorf("Nome da ROM incorreto: esperado %s, obtido %s", state.ROMName, loaded.ROMName)
		}
		if loaded.ROMHash != state.ROMHash {
			t.Errorf("Hash da ROM incorreto: esperado %s, obtido %s", state.ROMHash, loaded.ROMHash)
		}
	})

	// Testa CopySlot
	t.Run("CopySlot", func(t *testing.T) {
		// Salva o estado no slot 0
		if err := sm.SaveToSlot(0, state); err != nil {
			t.Errorf("Erro ao salvar estado: %v", err)
		}

		// Copia do slot 0 para o slot 1
		if err := sm.CopySlot(0, 1); err != nil {
			t.Errorf("Erro ao copiar slot: %v", err)
		}

		// Carrega o estado do slot 1
		loaded, err := sm.LoadFromSlot(1)
		if err != nil {
			t.Errorf("Erro ao carregar estado: %v", err)
		}

		// Verifica se os dados foram preservados
		if loaded.Version != state.Version {
			t.Errorf("Versão incorreta: esperado %d, obtido %d", state.Version, loaded.Version)
		}
		if loaded.ROMName != state.ROMName {
			t.Errorf("Nome da ROM incorreto: esperado %s, obtido %s", state.ROMName, loaded.ROMName)
		}
		if loaded.ROMHash != state.ROMHash {
			t.Errorf("Hash da ROM incorreto: esperado %s, obtido %s", state.ROMHash, loaded.ROMHash)
		}
	})

	// Testa ValidateState
	t.Run("Validate", func(t *testing.T) {
		if !sm.ValidateState(state, state.ROMName, state.ROMHash) {
			t.Error("Estado deveria ser válido")
		}

		if sm.ValidateState(state, "wrong.gba", state.ROMHash) {
			t.Error("Estado não deveria ser válido com nome errado")
		}

		if sm.ValidateState(state, state.ROMName, "wronghash") {
			t.Error("Estado não deveria ser válido com hash errado")
		}
	})

	// Testa GetStateSize
	t.Run("StateSize", func(t *testing.T) {
		size, err := sm.GetStateSize(state)
		if err != nil {
			t.Errorf("Erro ao obter tamanho do estado: %v", err)
		}

		if size <= 0 {
			t.Error("Tamanho do estado deveria ser maior que zero")
		}
	})

	// Testa GetSlotFileSize
	t.Run("SlotFileSize", func(t *testing.T) {
		// Salva o estado no slot 0
		if err := sm.SaveToSlot(0, state); err != nil {
			t.Errorf("Erro ao salvar estado: %v", err)
		}

		size, err := sm.GetSlotFileSize(0)
		if err != nil {
			t.Errorf("Erro ao obter tamanho do arquivo: %v", err)
		}

		if size <= 0 {
			t.Error("Tamanho do arquivo deveria ser maior que zero")
		}
	})

	// Testa GetSlotModTime
	t.Run("SlotModTime", func(t *testing.T) {
		// Salva o estado no slot 0
		if err := sm.SaveToSlot(0, state); err != nil {
			t.Errorf("Erro ao salvar estado: %v", err)
		}

		modTime, err := sm.GetSlotModTime(0)
		if err != nil {
			t.Errorf("Erro ao obter data de modificação: %v", err)
		}

		if modTime.IsZero() {
			t.Error("Data de modificação não deveria ser zero")
		}
	})

	// Testa erros de slot inválido
	t.Run("InvalidSlot", func(t *testing.T) {
		if err := sm.SaveToSlot(-1, state); err == nil {
			t.Error("SaveToSlot deveria falhar com slot negativo")
		}

		if err := sm.SaveToSlot(10, state); err == nil {
			t.Error("SaveToSlot deveria falhar com slot maior que o limite")
		}

		if _, err := sm.LoadFromSlot(-1); err == nil {
			t.Error("LoadFromSlot deveria falhar com slot negativo")
		}

		if _, err := sm.LoadFromSlot(10); err == nil {
			t.Error("LoadFromSlot deveria falhar com slot maior que o limite")
		}
	})
}
