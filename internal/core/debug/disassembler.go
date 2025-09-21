package debug

// Disassembler fornece funcionalidades para desmontar instruções ARM e Thumb
type Disassembler struct {
	// Callbacks para acessar a memória
	readWord  func(addr uint32) uint32
	readHalf  func(addr uint32) uint16
	isThumb   func(addr uint32) bool
	getSymbol func(addr uint32) string
}

// NewDisassembler cria uma nova instância do disassembler
func NewDisassembler(
	readWord func(addr uint32) uint32,
	readHalf func(addr uint32) uint16,
	isThumb func(addr uint32) bool,
	getSymbol func(addr uint32) string,
) *Disassembler {
	return &Disassembler{
		readWord:  readWord,
		readHalf:  readHalf,
		isThumb:   isThumb,
		getSymbol: getSymbol,
	}
}

// DisassembleRange desmonta um intervalo de instruções
func (d *Disassembler) DisassembleRange(startAddr, endAddr uint32) []string {
	var result []string
	addr := startAddr

	for addr < endAddr {
		if d.isThumb(addr) {
			instr := d.readHalf(addr)
			result = append(result, d.formatThumbInstruction(addr, instr))
			addr += 2
		} else {
			instr := d.readWord(addr)
			result = append(result, d.formatARMInstruction(addr, instr))
			addr += 4
		}
	}

	return result
}

// DisassembleContext desmonta instruções ao redor de um endereço específico
func (d *Disassembler) DisassembleContext(addr uint32, contextLines int) []string {
	startAddr := addr - uint32(contextLines*4)
	endAddr := addr + uint32((contextLines+1)*4)
	return d.DisassembleRange(startAddr, endAddr)
}
