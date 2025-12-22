package mcstatus

import (
	"context"
	"fmt"
	"mc-server-webui/database"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

// ServerStatus represents the simplified status of a Minecraft server.
type ServerStatus struct {
	Online      bool   `json:"online"`
	MOTD        string `json:"motd,omitempty"`
	Players     int    `json:"players"`    // Removed omitempty
	MaxPlayers  int    `json:"maxPlayers"` // Removed omitempty
	Version     string `json:"version,omitempty"`
	Protocol    int    `json:"protocol,omitempty"`
	Favicon     string `json:"favicon,omitempty"`
	LastUpdated time.Time `json:"lastUpdated"`
	Error       string `json:"error,omitempty"`
}

// QueryMinecraftServer queries a Minecraft server and returns its status.
func QueryMinecraftServer(server *database.Server) (*ServerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	host := server.Address
	port := uint16(25565) // Default Minecraft port

	// 1. Parse address for manual port override (e.g., example.com:12345)
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) == 2 {
			host = parts[0]
			p, err := strconv.ParseUint(parts[1], 10, 16)
			if err == nil {
				port = uint16(p)
			}
		}
	} else {
		// 2. If no manual port, check for SRV record (_minecraft._tcp.example.com)
		_, srvs, err := net.LookupSRV("minecraft", "tcp", host)
		if err == nil && len(srvs) > 0 {
			// Use the highest priority (lowest value) SRV record
			host = srvs[0].Target
			// Target often comes with a trailing dot, remove it
			host = strings.TrimSuffix(host, ".")
			port = srvs[0].Port
		}
	}

	// Default status if query fails
	serverStatus := &ServerStatus{
		Online:      false,
		LastUpdated: time.Now(),
	}

	response, err := status.Modern(ctx, host, port)
	if err != nil {
		serverStatus.Error = err.Error()
		return serverStatus, fmt.Errorf("failed to query server: %w", err)
	}

	serverStatus.Online = true
	serverStatus.MOTD = response.MOTD.Clean 
	
	if response.Players.Online != nil {
		serverStatus.Players = int(*response.Players.Online)
	}
	if response.Players.Max != nil {
		serverStatus.MaxPlayers = int(*response.Players.Max)
	}
	
	serverStatus.Version = response.Version.Name.Clean
	serverStatus.Protocol = int(response.Version.Protocol)

	if response.Favicon != nil {
		serverStatus.Favicon = *response.Favicon
	}

	return serverStatus, nil
}
