package mcstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tionis/mcow/database"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

// Player represents a player on the server.
type Player struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// ServerStatus represents the simplified status of a Minecraft server.
type ServerStatus struct {
	Online        bool      `json:"online"`
	MOTD          string    `json:"motd,omitempty"`
	Players       int       `json:"players"`    // Removed omitempty
	MaxPlayers    int       `json:"maxPlayers"` // Removed omitempty
	SamplePlayers []Player  `json:"samplePlayers,omitempty"`
	Version       string    `json:"version,omitempty"`
	Protocol      int       `json:"protocol,omitempty"`
	Favicon       string    `json:"favicon,omitempty"`
	LastUpdated   time.Time `json:"lastUpdated"`
	Error         string    `json:"error,omitempty"`
}

// BlueMapResponse structure for parsing
type BlueMapResponse struct {
	Players []struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
	} `json:"players"`
}

// QueryMinecraftServer queries a Minecraft server and returns its status.
func QueryMinecraftServer(server *database.Server) (*ServerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	host := server.Address
	port := uint16(25565) // Default Minecraft port

	// 1. Parse address for manual port override
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
		// 2. SRV lookup
		_, srvs, err := net.LookupSRV("minecraft", "tcp", host)
		if err == nil && len(srvs) > 0 {
			host = srvs[0].Target
			host = strings.TrimSuffix(host, ".")
			port = srvs[0].Port
		}
	}

	// Default status if query fails
	serverStatus := &ServerStatus{
		Online:      false,
		LastUpdated: time.Now(),
	}

	res, err := status.Modern(ctx, host, port)
	if err != nil {
		serverStatus.Error = err.Error()
		return serverStatus, fmt.Errorf("failed to query server: %w", err)
	}

	serverStatus.Online = true
	serverStatus.MOTD = res.MOTD.Clean 
	
	if res.Players.Online != nil {
		serverStatus.Players = int(*res.Players.Online)
	}
	if res.Players.Max != nil {
		serverStatus.MaxPlayers = int(*res.Players.Max)
	}
	
	playerMap := make(map[string]bool)

	if res.Players.Sample != nil {
		for _, p := range res.Players.Sample {
			var name, id string
			if p.Name.Clean != "" {
				name = p.Name.Clean
			}
			if p.ID != "" {
				id = p.ID
			}
			serverStatus.SamplePlayers = append(serverStatus.SamplePlayers, Player{Name: name, ID: id})
			playerMap[name] = true
		}
	}
	
	// Fetch BlueMap players if URL is configured
	if server.BlueMapURL != "" {
		blueMapPlayers := fetchBlueMapPlayers(server.BlueMapURL)
		for _, p := range blueMapPlayers {
			if !playerMap[p.Name] {
				serverStatus.SamplePlayers = append(serverStatus.SamplePlayers, p)
				playerMap[p.Name] = true
			}
		}
		// If MC query returned 0 online but BlueMap has players, update count? 
		// Or trust MC query? Usually MC query is authoritative for count. 
		// But sample is limited to 12. BlueMap might show all.
		// Let's rely on MC Query for total count, but augment sample list.
	}
	
	serverStatus.Version = res.Version.Name.Clean
	serverStatus.Protocol = int(res.Version.Protocol)

	if res.Favicon != nil {
		serverStatus.Favicon = *res.Favicon
	}

	return serverStatus, nil
}

func fetchBlueMapPlayers(baseURL string) []Player {
	// Construct URL. Assume baseURL is root.
	// Try /maps/world/live/players.json first (default world name)
	// If the user provided a full path to map, we might need to be smarter, but let's stick to the old app's assumption.
	url := strings.TrimSuffix(baseURL, "/") + "/maps/world/live/players.json"
	
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil
	}
	
	var bmResp BlueMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&bmResp); err != nil {
		return nil
	}
	
	var players []Player
	for _, p := range bmResp.Players {
		players = append(players, Player{Name: p.Name, ID: p.UUID})
	}
	return players
}
