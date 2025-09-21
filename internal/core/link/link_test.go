package link

import (
	"testing"
)

func TestGetSIOMode(t *testing.T) {
	tests := []struct {
		name   string
		siocnt uint16
		rcnt   uint16
		want   SIOMode
	}{
		{
			name:   "JOYBUS mode",
			siocnt: 0,
			rcnt:   0xC000,
			want:   JOYBUS,
		},
		{
			name:   "GP mode",
			siocnt: 0,
			rcnt:   0x8000,
			want:   GP,
		},
		{
			name:   "NORMAL32 mode",
			siocnt: SIO_TRANS_32BIT,
			rcnt:   0,
			want:   NORMAL32,
		},
		{
			name:   "MULTIPLAYER mode",
			siocnt: 0x2000,
			rcnt:   0,
			want:   MULTIPLAYER,
		},
		{
			name:   "UART mode",
			siocnt: 0x3000,
			rcnt:   0,
			want:   UART,
		},
		{
			name:   "NORMAL8 mode (default)",
			siocnt: 0,
			rcnt:   0,
			want:   NORMAL8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSIOMode(tt.siocnt, tt.rcnt); got != tt.want {
				t.Errorf("GetSIOMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkInitAndClose(t *testing.T) {
	link := NewLink()

	// Testa inicialização com modo desconectado
	if state := link.InitLink(LINK_DISCONNECTED); state != LINK_ABORT {
		t.Errorf("InitLink(LINK_DISCONNECTED) = %v, want %v", state, LINK_ABORT)
	}

	// Testa inicialização com modo válido
	if state := link.InitLink(LINK_CABLE_SOCKET); state != LINK_OK {
		t.Errorf("InitLink(LINK_CABLE_SOCKET) = %v, want %v", state, LINK_OK)
	}

	// Testa inicialização quando já conectado
	if state := link.InitLink(LINK_CABLE_SOCKET); state != LINK_ERROR {
		t.Errorf("InitLink() quando já conectado = %v, want %v", state, LINK_ERROR)
	}
}

func TestStartLink(t *testing.T) {
	link := NewLink()
	link.InitLink(LINK_CABLE_SOCKET)

	tests := []struct {
		name       string
		siocnt     uint16
		wantStart  bool
		wantUpdate bool
	}{
		{
			name:       "MULTIPLAYER master start",
			siocnt:     SIO_TRANS_START,
			wantStart:  true,
			wantUpdate: true,
		},
		{
			name:       "MULTIPLAYER slave no start",
			siocnt:     0,
			wantStart:  false,
			wantUpdate: false,
		},
		{
			name:       "NORMAL32 start",
			siocnt:     SIO_TRANS_START | SIO_TRANS_32BIT,
			wantStart:  true,
			wantUpdate: true,
		},
		{
			name:       "NORMAL8 start",
			siocnt:     SIO_TRANS_START,
			wantStart:  true,
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link.transferring = false
			link.StartLink(tt.siocnt)

			if link.transferring != tt.wantStart {
				t.Errorf("StartLink() transferring = %v, want %v", link.transferring, tt.wantStart)
			}

			if tt.wantUpdate {
				link.Update(1000)
				if link.lastUpdate != 1000 {
					t.Errorf("Update() lastUpdate = %v, want %v", link.lastUpdate, 1000)
				}
			}
		})
	}
}
