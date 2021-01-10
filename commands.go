package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caromo/rinako/collections"
	"github.com/pkg/errors"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "璃奈ちゃんボード「むっ！」")
	}

	mEvent := messageEvent{
		session: s,
		message: m,
	}
	err := mEvent.enrich()
	if err != nil {
		log.Printf("Failed to enrich message event: %v", err)
	}

	// if (mEvent.member.User.ID == "168273566448615425")

	if strings.HasPrefix(m.Content, rinako.config.Discriminator) {
		removeDisc := strings.Replace(m.Content, rinako.config.Discriminator, "", 1)
		command, args := removeHead(strings.Split(removeDisc, " "))

		if err = mEvent.processCommand(command, args); err != nil {
			log.Printf("Error executing %s with args %s: %s", command, args, err)
		}
	}
	fmt.Printf("Roulette name: %s\n", rinako.config.RouletteName)
}

func (m *messageEvent) processCommand(command string, args []string) (err error) {
	switch command {
	case "register":
		m.register()
	case "authorize":
		m.authorize(args)
	case "deauthorize":
		m.deauthorize(args)
	case "promote":
		m.promote(args)
	case "demote":
		m.demote(args)
	case "test":
		m.test()
	case "setCh":
		m.setCh()
	case "role":
		rolech := rinako.GetRoleCh(m.message.GuildID)
		if rolech == "" {
			m.sendMessagef("Role channel not set. use %ssetCh in desired channel to set.", rinako.config.Discriminator)
		} else if m.message.ChannelID == rolech {
			m.role(args)
		} else {
			botM, _ := m.sendMessagef("Role cannot be added here. Please go to <#%s>", rolech)
			timer := time.NewTimer(5 * time.Second)
			go func() {
				<-timer.C
				m.cleanup(botM)
			}()
		}
	case "tag":
		m.tag(args)
	case "untag":
		m.untag(args)
	case rinako.config.RouletteName:
		m.roulette()
	default:
	}

	return
}

func checkExists(m *messageEvent, role string) (exists bool) {
	roleDescs := rinako.GetAllowedRolesForServer(m.guild.ID)
	var roleNameList []string
	for _, x := range roleDescs {
		roleNameList = append(roleNameList, x.Role)
	}
	fmt.Printf("CheckExists: roleNameList: %v, role %s\n", roleNameList, role)
	_, exists = findCaseInsensitive(roleNameList, role)
	if !exists {
		botM, _ := m.sendMessagef("Role %s does not exist or is off-limits.", role)
		go func() {
			timer1 := time.NewTimer(10 * time.Second)
			<-timer1.C
			m.cleanup(botM)
		}()
		return
	}
	return
}

func (m *messageEvent) register() {
	serv, err := rinako.GetServer(m.guild.ID)

	if !serv.IsEmpty() {
		m.sendMessage("Server is already registered.")
	} else if err != nil {
		fmt.Printf("Err %v\n", err)
		m.sendMessagef("Failed to find server %s", m.guild.Name)
	} else {
		err := rinako.AddServer(m.guild.ID, m.guild.Name)
		if err != nil {
			_, _ = m.sendMessagef("Failed to add server %s", m.guild.Name)
		} else {
			m.sendMessage("Successfully registered server.")
		}
	}
}

func (m *messageEvent) setCh() {
	if !m.isElevatedOrOwner() {
		return
	} else {
		serv, err := rinako.GetServer(m.guild.ID)

		if err = rinako.SetRoleCh(serv.ID, m.channel.ID); err != nil {
			m.sendMessagef("Failed to set channel as Role Channel: %s", err)
		} else {
			m.sendMessage("Successfully set channel as Role Channel")
		}
	}
}

func (m *messageEvent) authorize(args []string) {
	if len(args) == 0 {
		m.sendMessage("`authorize` command use: aaskdpoaksdopaksodp")
	} else if !m.isElevatedOrOwner() {
		return
	} else {
		fmt.Printf("%s\n", args)
		role, descList, err := m.getRoleFromArgs(args)
		if err != nil {
			return
		}
		desc := strings.ReplaceAll(strings.Join(descList, " "), "\"", "")

		roleDesc := collections.RoleDesc{
			Role: role.Name,
			Desc: desc,
		}

		if err = rinako.AddAllowedRole(m.guild.ID, roleDesc); err != nil {
			m.sendMessagef("Error authorizing role %s: %s", roleDesc.Role, err)
		} else {
			m.sendMessagef("Successfully authorized role: %s", roleDesc.Role)
		}
	}
}

func (m *messageEvent) deauthorize(args []string) {
	if len(args) == 0 {
		m.sendMessage("`deauthorize` command use: aaskdpoaksdopaksodp")
	} else if !m.isElevatedOrOwner() {
		return
	} else {
		role, _, err := m.getRoleFromArgs(args)
		if err != nil {
			return
		}

		if err := rinako.RemoveAllowedRole(m.guild.ID, role.Name); err != nil {
			m.sendMessagef("Error deauthorizing role %s: %s", role.Name, err)
		} else {
			m.sendMessagef("Successfully deauthorized role: %s", role.Name)
		}
	}

}

func (m *messageEvent) promote(args []string) {
	if len(args) != 1 {
		m.sendMessage("`promote` command use: aaskdpoaksdopaksodp")
	} else if !m.isElevatedOrOwner() {
		return
	} else {
		role, _, err := m.getRoleFromArgs(args)
		if err != nil {
			return
		}

		if err := rinako.PromoteRole(m.guild.ID, role.ID); err != nil {
			m.sendMessagef("Error promoting role %s: %s", role.Name, err)
		} else {
			m.sendMessagef("Successfully promoted role: %s", role.Name)
		}
	}
}

func (m *messageEvent) demote(args []string) {
	if len(args) != 1 {
		m.sendMessage("`demote` command use: aaskdpoaksdopaksodp")
	} else if !m.isElevatedOrOwner() {
		return
	} else {
		role, _, err := m.getRoleFromArgs(args)
		if err != nil {
			return
		}

		if err := rinako.DemoteRole(m.guild.ID, role.ID); err != nil {
			m.sendMessagef("Error demoting role %s: %s", role.Name, err)
		} else {
			m.sendMessagef("Successfully demoted role: %s", role.Name)
		}
	}
}

func (m *messageEvent) getRoleFromArgs(args []string) (role *discordgo.Role, tail []string, err error) {
	strArgs := extractQuotes(strings.Join(args, " "))
	rString, tail := removeHead(strArgs)

	role, err = m.getRole(toCleanRole(rString))
	if err != nil {
		m.sendMessagef("No such role exists: %s", args[0])
	}
	return role, tail, err
}

// TODO: Clean up this code because it's pretty hacky right now.
func (m *messageEvent) role(args []string) {
	var botM *discordgo.Message
	opt, role := removeHead(args)
	fmt.Printf("user: %s\nargs: %v\n", m.member.User.Username, args)
	toModify := toCleanRole(strings.Join(role, " "))

	switch opt {
	case "add":
		exists := checkExists(m, toModify)
		if !exists {
			break
		}
		toModifyRole, err := m.getRole(toModify)
		if err != nil {
			log.Printf("Error getting role matching %s", toModify)
		}

		err = m.session.GuildMemberRoleAdd(m.message.GuildID, m.member.User.ID, toModifyRole.ID)
		if err != nil {
			log.Printf("Error adding role %s on User %s: %s", toModifyRole.Name, m.member.User.Username, err)

			botM, _ = m.sendMessagef("Failed to add role %s", toModifyRole.Name)
		} else {
			botM, _ = m.sendMessagef("Added role %s", toModifyRole.Name)
		}

	case "remove":
		exists := checkExists(m, toModify)
		if !exists {
			break
		}
		toModifyRole, err := m.getRole(toModify)
		if err != nil {
			log.Printf("Error getting role matching %s", toModify)
		}

		err = m.session.GuildMemberRoleRemove(m.message.GuildID, m.member.User.ID, toModifyRole.ID)
		if err != nil {
			log.Printf("Error adding role %s on User %s", toModifyRole.Name, m.member.User.Username)
			botM, _ = m.sendMessagef("Failed to remove role %s", toModifyRole.Name)
		} else {
			botM, _ = m.sendMessagef("Removed role %s", toModifyRole.Name)
		}
	case "list":
		m.listRoles()
		timer1 := time.NewTimer(60 * time.Second)
		go func() {
			<-timer1.C
			m.cleanup(botM)
		}()
	default:
		botM, _ = m.sendMessagef("'role' command usage: `%srole [add/remove] \"role\"`", rinako.config.Discriminator)
		timer1 := time.NewTimer(60 * time.Second)
		go func() {
			<-timer1.C
			m.cleanup(botM)
		}()
	}

}

func (m *messageEvent) getRoleByID(id string) (res *discordgo.Role, err error) {
	var guild *discordgo.Guild
	guild, err = m.session.Guild(m.message.GuildID)
	if err != nil {
		log.Printf("Error getting guild %s: %s", m.message.GuildID, err)
		return
	}
	roles := guild.Roles
	for _, r := range roles {
		if r.ID == id {
			res = r
		}
	}
	if res == nil {
		err = errors.Errorf("Role not found")
		log.Printf("Error getting role: %s", err)
	}
	return
}

func (m *messageEvent) getRole(role string) (res *discordgo.Role, err error) {
	var guild *discordgo.Guild
	guild, err = m.session.Guild(m.message.GuildID)
	if err != nil {
		log.Printf("Error getting guild %s: %s", m.message.GuildID, err)
		return
	}
	roles := guild.Roles
	for _, r := range roles {
		if strings.ToLower(r.Name) == strings.ToLower(role) {
			res = r
		}
	}
	if res == nil {
		err = errors.Errorf("Role not found: %s", role)
		log.Printf("Error getting role: %s", err)
	}
	return
}

func (m *messageEvent) listRoles() {
	roles := rinako.GetAllowedRolesForServer(m.guild.ID)
	// embedField := constructRoleEmbeds(rinako.config.AllowedRoles)
	embedField := constructRoleEmbeds(roles)
	color, _ := strconv.ParseUint(rinako.config.Color, 16, 32)
	var embed = discordgo.MessageEmbed{
		Title:  "Available Roles",
		Fields: embedField,
		Color:  int(color),
	}

	botM, _ := m.session.ChannelMessageSendEmbed(m.channel.ID, &embed)
	timer := time.NewTimer(30 * time.Second)
	go func() {
		<-timer.C
		m.cleanup(botM)
	}()
}

func (m *messageEvent) isElevatedOrOwner() bool {
	//if message sender is server owner OR belongs to specified roles, let them through
	hasElevatedRole := false
	server, err := rinako.GetServer(m.guild.ID)
	if err != nil {
		m.sendMessagef("Server is not yet registered: %sregister", rinako.config.Discriminator)
	}
	for _, id := range m.member.Roles {
		_, exists := find(server.ElevatedRoles, id)
		if exists {
			hasElevatedRole = true
		}
	}
	result := (m.member.User.ID == m.guild.OwnerID || hasElevatedRole)
	if !result {
		m.sendMessage("Command inaccessible: missing permissions")
	}
	return result
}

func constructRoleEmbeds(field []collections.RoleDesc) (embeds []*discordgo.MessageEmbedField) {
	var value = ""
	for i, rd := range field {
		value = value + "**" + rd.Role + "**" + "  -  " + rd.Desc
		if i < len(field) {
			value = value + "\n"
		}
	}
	toAdd := discordgo.MessageEmbedField{
		Name:  "Roles",
		Value: value,
	}
	embeds = append(embeds, &toAdd)
	return
}

//People tagged under roulette can tag others
func (m *messageEvent) tag(args []string) {

	if len(args) == 0 {
		m.sendMessagef("Use: %stag @<name>", rinako.config.Discriminator)
	} else if !m.isElevatedOrOwner() {
		return
	}
	memberID := args[0]

	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Fatal(err)
		err = errors.New("Failed to extract user from message")
		m.sendMessagef("Failed to extract user from message")
		return
	}
	memberID = reg.ReplaceAllString(memberID, "")
	rinako.AddRoulName(m.guild.ID, memberID)
	m.sendMessagef("Added <@!%s>", memberID)
	return
}

//...but they can't remove themselves
func (m *messageEvent) untag(args []string) {
	if len(args) == 0 {
		m.sendMessagef("Use: %suntagtag @<name>", rinako.config.Discriminator)
	} else if !m.isElevatedOrOwner() {
		return
	}
	serv, _ := rinako.GetServer(m.guild.ID)
	memberID := args[0]
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Fatal(err)
		err = errors.New("Failed to extract user from message")
		m.sendMessagef("Failed to extract user from message")
		return
	}
	memberID = reg.ReplaceAllString(memberID, "")
	if _, exists := find(serv.RouletteNames, memberID); exists && memberID != rinako.config.OverrideID {
		m.sendMessage(rinako.config.RoulettePText)
	} else {
		rinako.RemoveRoulName(m.guild.ID, memberID)
		m.sendMessagef("Removed <@!%s>")
	}
	return
}

func (m *messageEvent) roulette() {
	serv, _ := rinako.GetServer(m.guild.ID)

	rand.Seed(time.Now().Unix())
	m.sendMessagef("<@!%s> %s", serv.RouletteNames[rand.Intn(len(serv.RouletteNames))], rinako.config.RouletteRText)

}
