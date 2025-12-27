# VisualBoy Go

![VisualBoy Go Logo](assets/logo.png)

O **VisualBoy Go** é um emulador de Game Boy e Game Boy Advance escrito em Go, inspirado no VisualBoyAdvance-M. Este projeto visa fornecer uma implementação moderna e eficiente da arquitetura do Game Boy com suporte a ROMs, save states, e uma interface gráfica intuitiva.

## 🎮 Características

*   **Suporte a múltiplas ROMs**: Carregamento completo de Cartridge, ROM, RAM e Save RAM.
*   **Sistema de Save States**: Gerenciamento completo de estados com slots de 0 a 9 e serialização.
*   **Interface Gráfica Moderna**: Baseada em SDL2 e OpenGL para renderização de alta performance.
*   **Filtros de Vídeo**: Suporte a escalonamento Nearest-neighbor, Bilinear, Scale2x e Scale3x.
*   **Controles Personalizáveis**: Mapeamento completo dos botões do Game Boy.
*   **Motor de Áudio**: Implementação completa da Unidade de Processamento de Áudio (APU).
*   **Temporização Precisa**: Emulação fiel dos temporizadores internos e sistema de interrupções.
*   **Modo Tela Cheia**: Alternância instantânea entre janela e tela cheia.
*   **Modo Pausa**: Funcionalidade de pausa e resumo imediata.

## 🏗️ Arquitetura do Projeto

O projeto segue uma arquitetura modular projetada para manutenibilidade e desempenho:


visual_go/
├── cmd/                    # Ponto de entrada da aplicação
│   └── main.go
├── internal/               # Lógica privada da aplicação
│   ├── core/               # Núcleos de emulação
│   │   ├── gb/             # Lógica específica do Game Boy
│   │   │   ├── cpu/        # Implementação da ARM7TDMI
│   │   │   ├── memory/     # Mapeamento de memória (suporte a MBC1-5)
│   │   │   ├── video/      # Emulação de LCD e PPU
│   │   │   ├── sound/      # APU (Unidade de Processamento de Áudio)
│   │   │   ├── input/      # Tratamento de entrada
│   │   │   ├── timer/      # Temporizador do sistema
│   │   │   ├── interrupts/ # Gerenciamento de interrupções
│   │   │   └── savestate/  # Serialização de estados
│   │   └── gba/            # Game Boy Advance (Futuro)
│   ├── gui/                # Interface Gráfica de Usuário
│   │   └── window.go       # Janela principal com contexto OpenGL
│   └── utils/              # Utilitários auxiliares
├── assets/                 # Recursos do projeto (Logos, Ícones)
├── examples/               # ROMs de teste e exemplos
├── go.mod                  # Definição do módulo Go
└── CHECKLIST.md            # Checklist de desenvolvimento


### Componentes Principais

#### CPU ([`internal/core/gb/cpu/cpu.go`](internal/core/gb/cpu/cpu.go))
Implementação completa do processador ARM7TDMI, suportando ambos os conjuntos de instruções ARM e Thumb. Possui um robusto pipeline de execução com gerenciamento preciso de flags e modos de operação (User, Supervisor, etc.).

#### Memória ([`internal/core/gb/memory/memory.go`](internal/core/gb/memory/memory.go))
Mapeamento completo da memória do hardware do Game Boy. Suporta controladores de bancos de memória MBC1, MBC2, MBC3 e MBC5. Gerencia Work RAM, High RAM, VRAM, OAM e detecção automática de cartuchos.

#### Vídeo ([`internal/core/gb/video/lcd.go`](internal/core/gb/video/lcd.go))
Emulação precisa do LCD, incluindo a PPU (Unidade de Processamento de Imagem). Gerencia camadas de fundo, sprites, renderização de janelas e sincronização de tempo com a CPU.

#### Som ([`internal/core/gb/sound/apu.go`](internal/core/gb/sound/apu.go))
Emula a Unidade de Processamento de Áudio com suporte a todos os 4 canais de som: Onda Quadrada (CH1 & CH2), Onda de Triângulo (CH3) e Ruído (CH4).

## 🚀 Começando

### Pré-requisitos

*   Go 1.18 ou superior
*   GCC (para bindings CGO)
*   Bibliotecas de desenvolvimento SDL2
*   Drivers OpenGL

### Instalação

bash
# Clone o repositório
git clone https://github.com/yourusername/visual_go.git
cd visual_go

# Construa o projeto
go build -o visualboy ./cmd/main.go

# Execute o emulador
./visualboy <caminho_para_rom>


## 📝 Uso

| Ação | Tecla |
| :--- | :--- |
| **Start** | Enter |
| **Select** | Shift Direito |
| **A** | Z |
| **B** | X |
| **Cima/Baixo/Esquerda/Direita** | Setas |
| **Salvar Estado** | F1-F9 |
| **Carregar Estado** | Shift + F1-F9 |
| **Alternar Tela Cheia** | Alt + Enter |
| **Pausar** | Espaço |

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para enviar um Pull Request. Certifique-se de que seu código segue a estrutura existente e inclui os testes relevantes.

## 📜 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para mais detalhes.