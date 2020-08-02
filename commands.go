package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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

	if strings.HasPrefix(m.Content, discriminator) {
		removeDisc := strings.Replace(m.Content, discriminator, "", 1)
		command, args := removeHead(strings.Split(removeDisc, " "))

		if err = mEvent.processCommand(command, args); err != nil {
			log.Printf("Error executing %s with args %s: %s", command, args, err)
		}
	}

}

func (m *messageEvent) processCommand(command string, args []string) (err error) {
	switch command {
	case "test":
		m.test()
	case "role":
		if m.message.ChannelID == rinako.config.RoleChannel {
			m.role(args)
		} else {
			botM, _ := m.sendMessagef("Role cannot be added here. Please go to <#%s>", roleCh)
			timer := time.NewTimer(5 * time.Second)
			go func() {
				<-timer.C
				m.cleanup(botM)
			}()
		}
	default:

	}

	return
}

func checkExists(m *messageEvent, role string) (exists bool) {
	_, exists = findCaseInsensitive(allowedRoleTitles, role)
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

// TODO: Clean up this code because it's pretty hacky right now.
func (m *messageEvent) role(args []string) {
	var botM *discordgo.Message
	opt, role := removeHead(args)
	fmt.Printf("args: %v\n", args)
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
	default:
		botM, _ = m.sendMessage("'role' command usage: `role [add/remove] \"role\"`")
	}
	timer1 := time.NewTimer(10 * time.Second)
	go func() {
		<-timer1.C
		m.cleanup(botM)
	}()
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
		log.Printf("Role not found: %s", err)
	}
	return
}

func (m *messageEvent) listRoles() {
	embedField := constructRoleEmbeds(rinako.config.AllowedRoles)

	var embed = discordgo.MessageEmbed{
		Title:  "Available Roles",
		Fields: embedField,
	}

	m.session.ChannelMessageSendEmbed(m.channel.ID, &embed)
}

func constructRoleEmbeds(field []RoleDesc) (embeds []*discordgo.MessageEmbedField) {
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
