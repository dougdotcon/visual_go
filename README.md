# VisualBoy Go

![VisualBoy Go Logo](assets/logo.png)

**VisualBoy Go** is a Game Boy and Game Boy Advance emulator written in Go, inspired by the classic VisualBoyAdvance-M. This project aims to provide a modern, efficient, and faithful implementation of the Game Boy architecture, featuring support for ROMs, save states, and an intuitive graphical interface.

## 🎮 Features

*   **Multi-ROM Support**: Seamless loading of Cartridge, ROM, RAM, and Save RAM data.
*   **Save State System**: Comprehensive state management with 0-9 slots and full serialization.
*   **Modern GUI**: Built with SDL2 and OpenGL for high-performance rendering.
*   **Video Filters**: Supports Nearest-neighbor, Bilinear, Scale2x, and Scale3x scaling algorithms.
*   **Customizable Controls**: Full remapping support for Game Boy input buttons.
*   **Audio Engine**: Complete implementation of the Audio Processing Unit (APU).
*   **Precision Timing**: Accurate emulation of internal timers and interrupt systems.
*   **Fullscreen Mode**: Easy toggling between windowed and fullscreen views.
*   **Pause Mode**: Instant pause and resume functionality.

## 🏗️ Architecture

The project follows a modular architecture designed for maintainability and performance:


visual_go/
├── cmd/                    # Main application entrypoint
│   └── main.go
├── internal/               # Private application logic
│   ├── core/               # Emulation cores
│   │   ├── gb/             # Game Boy specific logic
│   │   │   ├── cpu/        # ARM7TDMI CPU implementation
│   │   │   ├── memory/     # Memory mapping (MBC1-5 support)
│   │   │   ├── video/      # LCD and PPU emulation
│   │   │   ├── sound/      # APU (Audio Processing Unit)
│   │   │   ├── input/      # Input handling
│   │   │   ├── timer/      # System timer
│   │   │   ├── interrupts/ # Interrupt handling
│   │   │   └── savestate/  # Save state serialization
│   │   └── gba/            # Game Boy Advance (Future)
│   ├── gui/                # Graphical User Interface
│   │   └── window.go       # Main window with OpenGL context
│   └── utils/              # Helper utilities
├── assets/                 # Project assets (Logos, Icons)
├── examples/               # Test ROMs and examples
├── go.mod                  # Go module definition
└── CHECKLIST.md            # Development checklist


### Core Components

#### CPU ([`internal/core/gb/cpu/cpu.go`](internal/core/gb/cpu/cpu.go))
Complete implementation of the ARM7TDMI processor, supporting both ARM and Thumb instruction sets. Features a robust execution pipeline with precise flag management and operation modes (User, Supervisor, etc.).

#### Memory ([`internal/core/gb/memory/memory.go`](internal/core/gb/memory/memory.go))
Full memory mapping for the Game Boy hardware. Supports MBC1, MBC2, MBC3, and MBC5 memory bank controllers. Handles Work RAM, High RAM, VRAM, OAM, and automatic cartridge detection.

#### Video ([`internal/core/gb/video/lcd.go`](internal/core/gb/video/lcd.go))
Precise LCD emulation including the PPU (Picture Processing Unit). Handles background layers, sprites, window rendering, and video timing synchronization with the CPU.

#### Sound ([`internal/core/gb/sound/apu.go`](internal/core/gb/sound/apu.go))
Emulates the Audio Processing Unit with support for all 4 sound channels: Square Wave (CH1 & CH2), Wave Wave (CH3), and Noise (CH4).

## 🚀 Getting Started

### Prerequisites

*   Go 1.18 or higher
*   GCC (for CGO bindings)
*   SDL2 development libraries
*   OpenGL drivers

### Installation

bash
# Clone the repository
git clone https://github.com/yourusername/visual_go.git
cd visual_go

# Build the project
go build -o visualboy ./cmd/main.go

# Run the emulator
./visualboy <path_to_rom>


## 📝 Usage

| Action | Key |
| :--- | :--- |
| **Start** | Enter |
| **Select** | Right Shift |
| **A** | Z |
| **B** | X |
| **Up/Down/Left/Right** | Arrow Keys |
| **Save State** | F1-F9 |
| **Load State** | Shift + F1-F9 |
| **Toggle Fullscreen** | Alt + Enter |
| **Pause** | Space |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. Ensure that your code follows the existing structure and includes relevant tests.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.