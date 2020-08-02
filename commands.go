package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
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
		m.role(args)
		// botM, _ := m.sendMessage("Role cannot be added here. Please go to <#739258927044100157>")
		// timer := time.NewTimer(5 * time.Second)
		// go func() {
		// 	<-timer.C
		// 	m.cleanup(botM)
		// }()
	default:

	}

	return
}

// TODO: Clean up this code because it's pretty hacky right now.
func (m *messageEvent) role(args []string) {
	var botM *discordgo.Message
	opt, role := removeHead(args)
	fmt.Printf("args: %v\n", args)
	toModify := toCleanRole(strings.Join(role, " "))
	_, exists := find(allowedRoles, toModify)
	if !exists {
		botM, _ = m.sendMessagef("Role %s does not exist or is off-limits.", toModify)
		go func() {
			timer1 := time.NewTimer(10 * time.Second)
			<-timer1.C
			m.cleanup(botM)
		}()
		return
	}
	switch opt {
	case "add":
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
		if !exists {
			botM, _ = m.sendMessagef("Role %s is not interactable.", toModify)
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

	default:
		botM, _ = m.sendMessage("'role' command usage: `role [add/remove] \"role\"`")
	}
	timer1 := time.NewTimer(10 * time.Second)
	go func() {
		<-timer1.C
		m.cleanup(botM)
	}()
}

func (m *messageEvent) cleanup(botMessage *discordgo.Message) (err error) {
	err = m.session.ChannelMessageDelete(m.message.ChannelID, m.message.ID)
	err = m.session.ChannelMessageDelete(m.message.ChannelID, botMessage.ID)
	return
}

func (m *messageEvent) sendMessage(text string) (res *discordgo.Message, err error) {
	res, err = m.session.ChannelMessageSend(m.message.ChannelID, text)
	return
}

func (m *messageEvent) sendMessagef(format string, a ...interface{}) (res *discordgo.Message, err error) {
	res, err = m.session.ChannelMessageSend(m.message.ChannelID, fmt.Sprintf(format, a))
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
		if r.Name == role {
			res = r
		}
	}
	if res == nil {
		err = errors.Errorf("Role not found: %s", role)
		log.Printf("Role not found: %s", err)
	}
	return
}

func (m *messageEvent) test() {
	m.sendMessage("璃奈ちゃんボード「ヤッタゼー！」")
}

func (m *messageEvent) enrich() (err error) {
	m.guild, err = m.session.Guild(m.message.GuildID)
	if err != nil {
		log.Printf("Error fetching guild data: %v", err)
		return
	}
	m.channel, err = m.session.Channel(m.message.ChannelID)
	if err != nil {
		log.Printf("Error fetching channel data: %v", err)
		return
	}
	m.member, err = m.session.GuildMember(m.guild.ID, m.message.Author.ID)
	if err != nil {
		log.Printf("Error fetching member data: %v", err)
		return
	}
	return
}

type messageEvent struct {
	session *discordgo.Session
	message *discordgo.MessageCreate

	guild   *discordgo.Guild
	channel *discordgo.Channel
	member  *discordgo.Member
}
