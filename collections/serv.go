package collections

import (
	"encoding/json"
	"reflect"
)

// Server represents server info pertinent to Rinako on Discord
type Server struct {
	//Server ID(GuildID under Discord API)
	ID            string `gorm: primary_key`
	Name          string
	RoleChannel   string
	AllowedRoles  []byte
	ElevatedRoles []byte
	RouletteNames []byte
}

func (s *Server) ToInfo() ServerInfo {
	var ar []RoleDesc
	var er []string
	var rn []string
	json.Unmarshal(s.AllowedRoles, &ar)
	json.Unmarshal(s.ElevatedRoles, &er)
	json.Unmarshal(s.ElevatedRoles, &rn)

	return ServerInfo{
		ID:            s.ID,
		Name:          s.Name,
		RoleChannel:   s.RoleChannel,
		AllowedRoles:  ar,
		ElevatedRoles: er,
		RouletteNames: rn,
	}
}

func (s Server) IsEmpty() bool {
	return reflect.ValueOf(s).IsZero()
}

// ServerInfo represents server info pertinent to Rinako on Discord
type ServerInfo struct {
	ID            string
	Name          string
	RoleChannel   string
	AllowedRoles  []RoleDesc
	ElevatedRoles []string
	RouletteNames []string
}

func (s ServerInfo) IsEmpty() bool {
	return reflect.ValueOf(s).IsZero()
}

//todo: make AllowedRoles a json representation
//https://github.com/go-gorm/datatypes
// RoleDesc is a role tag and description
type RoleDesc struct {
	Role string `toml:"role" json:"role"`
	Desc string `toml:"desc" json:"desc"`
}

// func (r *RoleDesc) ToJSON() []byte {
// 	b, err := json.Marshal(r)
// 	if err != nil {
// 		log.Printf("Failed to marshall: %v", err)
// 		return nil
// 	}
// 	return b
// }
