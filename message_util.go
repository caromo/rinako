package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func (m *messageEvent) cleanup(botMessage *discordgo.Message) (err error) {
	err = m.session.ChannelMessageDelete(m.message.ChannelID, m.message.ID)
	if botMessage != nil {
		err = m.session.ChannelMessageDelete(m.message.ChannelID, botMessage.ID)
	}
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
