package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pkg/errors"

	coll "github.com/caromo/rinako/collections"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func InitializeDB(dbpath string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", dbpath)
}

func (r *Rinako) AddServer(id, name string) (err error) {
	newServer := coll.Server{
		ID:   id,
		Name: name,
	}
	if err = r.db.FirstOrCreate(&newServer).Error; err != nil {
		log.Printf("error registering server: %s", err)
	}
	return
}

func (r *Rinako) GetServer(id string) (servInfo coll.ServerInfo, err error) {
	serv := coll.Server{}
	if err = r.db.Where(&coll.Server{ID: id}).First(&serv).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("error getting server: %s", err)
		return
	}
	fmt.Printf("serv: %v\n", serv)
	servInfo = serv.ToInfo()
	return servInfo, nil
}

func (r *Rinako) getServerInternal(id string) (serv coll.Server, err error) {
	serv = coll.Server{}
	if err = r.db.Where(&coll.Server{ID: id}).First(&serv).Error; err != nil {
		log.Printf("error getting server: %s", err)
		return
	}
	return serv, nil
}

func (r *Rinako) AddAllowedRole(server string, role coll.RoleDesc) (err error) {
	serv := coll.Server{}
	if err = r.db.Where(&coll.Server{ID: server}).Find(&serv).Error; err != nil {
		log.Printf("error getting server: %s", err)
		return errors.New("Could not find server")
	}

	oldRoles := jsonToRoles(serv.AllowedRoles)
	err = checkRoleExists(oldRoles, role)
	if err != nil {
		return err
	}
	serv.AllowedRoles = rolesToJson(append(oldRoles, role))
	if err = r.db.Save(&serv).Error; err != nil {
		log.Printf("error authorizing role: %s", err)
		return errors.New("Failed to save authorized role")
	}

	serv2 := coll.Server{}
	r.db.Where(&coll.Server{ID: server}).Find(&serv2)
	fmt.Printf("New server obj: %+v\n", serv2)
	return
}

func (r *Rinako) RemoveAllowedRole(server string, role string) (err error) {
	serv := coll.Server{}
	if err = r.db.Where(&coll.Server{ID: server}).Find(&serv).Error; err != nil {
		log.Printf("error getting server: %s", err)
		return err
	}
	oldRoles := jsonToRoles(serv.AllowedRoles)
	index := findRoleByName(role, oldRoles)
	if index == -1 {
		log.Printf("error removing role: not found in allowed list")
		err = errors.New(fmt.Sprintf("Role '%s' is already unauthorized", role))
		return err
	}
	newRoles := removeRole(oldRoles, index)
	newAllowed := rolesToJson(newRoles)
	serv.AllowedRoles = newAllowed
	if err = r.db.Save(&serv).Error; err != nil {
		log.Printf("error setting role: %s", err)
		return err
	}
	return
}

func checkRoleExists(roles []coll.RoleDesc, role coll.RoleDesc) (err error) {
	toCheck := role.Role
	for _, r := range roles {
		if r.Role == toCheck {
			return errors.New("Role is already authorized")
		}
	}
	return
}

func findRole(role coll.RoleDesc, list []coll.RoleDesc) (index int) {
	index = -1
	for i, rd := range list {
		if role == rd {
			index = i
		}
	}
	return index
}

func findRoleByName(role string, list []coll.RoleDesc) (index int) {
	index = -1
	for i, rd := range list {
		if role == rd.Role {
			index = i
		}
	}
	return index
}

func removeRole(s []coll.RoleDesc, i int) []coll.RoleDesc {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (r *Rinako) GetAllowedRolesForServer(server string) (roles []coll.RoleDesc) {
	serv := coll.Server{}
	if err := r.db.Where(&coll.Server{ID: server}).Find(&serv).Error; err != nil {
		log.Printf("error getting server: %s", err)
		return nil
	}
	return jsonToRoles(serv.AllowedRoles)
}

func (r *Rinako) PromoteRole(server string, role string) (err error) {
	serv, err := r.getServerInternal(server)

	var oldElevated []string
	var newElevated []byte
	json.Unmarshal(serv.ElevatedRoles, &oldElevated)
	newElevatedSlice := appendUnique(oldElevated, role)
	if len(newElevatedSlice) == len(oldElevated) {
		return errors.New("Role is already elevated")
	}
	newElevated, _ = json.Marshal(newElevatedSlice)
	serv.ElevatedRoles = newElevated
	if err = r.db.Save(&serv).Error; err != nil {
		log.Printf("error elevating role: %s", err)
		return errors.New("Failed to save elevated role")
	}

	return
}

func (r *Rinako) DemoteRole(server string, role string) (err error) {
	serv, err := r.getServerInternal(server)

	var oldElevated []string
	var newElevated []byte
	json.Unmarshal(serv.ElevatedRoles, &oldElevated)
	index, exists := find(oldElevated, role)
	if !exists {
		return errors.New("Role was not elevated")
	}

	newElevated, _ = json.Marshal(removeFromSlice(oldElevated, index))

	serv.ElevatedRoles = newElevated
	if err = r.db.Save(&serv).Error; err != nil {
		log.Printf("error elevating role: %s", err)
		return errors.New("Failed to save elevated role")
	}

	return
}
