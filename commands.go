package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caromo/rinako/collections"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
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
		return
	}

	// if something matches this pattern:
	if strings.Contains(m.Content, "https://twitter.com") || strings.Contains(m.Content, "https://x.com") {
		linkForAPI, err := convertToVXLink(m.Content)
		if err != nil {
			log.Printf("Error converting to vxtwitter link: %s", err)
			s.ChannelMessageSendReply(m.ChannelID, "Error parsing Twitter link...", m.Message.Reference())
			return
		}
		HandleTweet(s, m.Message, linkForAPI, true)
	}

}

//HandleTweet(message, url) will be recursive and handle one level of a tweet at a time
func HandleTweet(s *discordgo.Session, message *discordgo.Message, url string, reply bool) {
	tweet, err := getTweet(url)
	if err != nil {
		log.Printf("Error getting tweet: %s", err)
		return
	}

	embed, videoOpt, err := buildEmbed(tweet)
	if err != nil {
		log.Printf("Error building embed: %s", err)
		return
	}

	if reply {
		s.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embed},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
			Reference: message.Reference(),
		})
		// s.ChannelMessageSendEmbedReply(message.ChannelID, embed, message.Reference())
	} else {
		s.ChannelMessageSendEmbed(message.ChannelID, embed)
	}
	if videoOpt.Needed {
		s.ChannelMessageSend(message.ChannelID, videoOpt.URL)
	}

	if tweet.QrtURL != nil {
		fmt.Printf("Found qrt: %s\n", *tweet.QrtURL)
		//convert to vxtwitter api link
		newLink, err := convertToVXLink(*tweet.QrtURL)
		if err != nil {
			log.Printf("Error converting to vxtwitter link: %s", err)
			return
		}
		// Sleep for 1 second
		time.Sleep(1 * time.Second)
		// Recursively call HandleTweet, no reply
		HandleTweet(s, message, newLink, false)
	}
}

func convertToVXLink(url string) (newLink string, err error) {
	twtxPattern := "http(s)?:\\/\\/(x|twitter)\\.com\\/(.+)"
	r, err := regexp.Compile(twtxPattern)
	if err != nil {
		log.Printf("Error compiling regex: %s", err)
		return
	}
	matches := r.FindStringSubmatch(url)
	if matches == nil {
		err = errors.New("No matches found")
		return
	} else {
		newLink = "https://api.vxtwitter.com/" + matches[3]
		return
	}
}

func buildEmbed(tweet *Tweet) (*discordgo.MessageEmbed, VideoInfo, error) {
	videoOpt := VideoInfo{}
	color, _ := strconv.ParseUint(rinako.config.Color, 16, 32)
	// turn the tweet into an embed
	const layout = "Mon Jan 2 15:04:05 -0700 2006"
	tweetTime, err := time.Parse(layout, tweet.Date)
	if err != nil {
		log.Printf("Error parsing tweet time: %s", err)
	}
	outputLayout := "2 Jan 2006 at 15:04:05"
	dateStr := tweetTime.Format(outputLayout)
	toBuild := NewEmbed().
		SetTitle(fmt.Sprintf("Tweet by %s (%s)", tweet.UserName, tweet.UserScreenName)).
		SetURL(tweet.TweetURL).
		SetDescription(tweet.Text).
		SetColor(int(color)).
		SetThumbnail(tweet.UserProfileImageURL).
		SetFooter(dateStr, "https://i.imgur.com/LKzfWwl.png")

	if len(tweet.MediaExtended) > 0 {
		switch tweet.MediaExtended[0].Type {
		case "video":
			log.Printf("Video: %s\n", tweet.MediaExtended[0].URL)
			videoOpt.Needed = true
			videoOpt.URL = tweet.MediaExtended[0].URL
		case "gif":
			log.Printf("Gif: %s\n", tweet.MediaExtended[0].URL)
			videoOpt.Needed = true
			videoOpt.URL = tweet.MediaExtended[0].URL
		case "image":
			toBuild.SetImage(tweet.MediaExtended[0].URL)
		default:
			toBuild.SetImage(tweet.MediaExtended[0].URL)
		}
	}

	return toBuild.MessageEmbed, videoOpt, err
}

func getTweet(url string) (tweet *Tweet, err error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error getting embed from API: %s", err)
		return
	}
	defer resp.Body.Close()
	tweet = &Tweet{}
	err = json.NewDecoder(resp.Body).Decode(tweet)
	if err != nil {
		fmt.Printf("Error decoding json: %s", err)
		return
	}
	return
}

//TODO: GetEmbedFromTweet, decouple getting from api and making the embed

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
	case "isapexplayable":
		m.checkApex()
	case "choose":
		m.choose(args)
	// case "deathroll":
	// 	m.deathroll()
	default:
	}

	return
}

// func (m *messageEvent) deathroll() {
// 	user := m.member.User.ID

// }

func checkExists(m *messageEvent, role string) (exists bool) {
	roleDescs := rinako.GetAllowedRolesForServer(m.guild.ID)
	var roleNameList []string
	for _, x := range roleDescs {
		roleNameList = append(roleNameList, x.Role)
	}
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
	result := (m.member.User.ID == m.guild.OwnerID || hasElevatedRole || m.member.User.ID == rinako.config.OverrideID)
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
	if err = rinako.AddRoulName(m.guild.ID, memberID); err != nil {
		m.sendMessagef("Error adding user <@%s>: %s", memberID, err)
	} else {
		m.sendMessagef("Added <@%s>", memberID)
	}
	return
}

//...but they can't remove themselves
func (m *messageEvent) untag(args []string) {
	if len(args) == 0 {
		m.sendMessagef("Use: %suntag @<name>", rinako.config.Discriminator)
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
	if _, exists := find(serv.RouletteNames, memberID); exists && m.member.User.ID != rinako.config.OverrideID {
		m.sendMessage(rinako.config.RoulettePText)
	} else {
		if err = rinako.RemoveRoulName(m.guild.ID, memberID); err != nil {
			m.sendMessagef("Error removing user <@%s>: %s", memberID, err)
		} else {
			m.sendMessagef("Removed <@%s>", memberID)
		}
	}
	return
}

func (m *messageEvent) roulette() {
	serv, _ := rinako.GetServer(m.guild.ID)
	rand.Seed(time.Now().Unix())
	m.sendMessagef("<@%s> %s", serv.RouletteNames[rand.Intn(len(serv.RouletteNames))], rinako.config.RouletteRText)

}

func (m *messageEvent) checkApex() {
	url := "https://apexlegendsstatus.com/current-map"
	resp, err := http.Get(url)
	if err != nil {
		m.sendMessage("The URL link is probably busted")
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		m.sendMessage("Parsing no workie")
	}
	mapDone := false
	timerDone := false
	currMap := ""
	currTimer := ""
	var f func(*html.Node, *bool, *bool, *string, *string)
	f = func(n *html.Node, mapDone *bool, timerDone *bool, currMap *string, currTimer *string) {
		if n.Type == html.TextNode && strings.Contains(n.Data, "Battle Royale") {
			*currMap = n.Data
			*mapDone = true
		}
		if n.Type == html.TextNode && strings.Contains(n.Data, "From") {
			*currTimer = n.Data
			*timerDone = true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, mapDone, timerDone, currMap, currTimer)
			if *mapDone && *timerDone {
				break
			}
		}
	}
	f(doc, &mapDone, &timerDone, &currMap, &currTimer)

	timerExp := regexp.MustCompile(`From ((0[0-9]|1[0-9]|2[0-3]):([0-5][0-9])) to ((0[0-9]|1[0-9]|2[0-3]):([0-5][0-9])) UTC`)

	timeHr := timerExp.FindStringSubmatch(currTimer)[5]
	timeMin := timerExp.FindStringSubmatch(currTimer)[6]

	now := time.Now()

	later := createLater(now, timeHr, timeMin)

	timeDiff := later.Sub(now)

	if currMap == "Battle Royale: Storm Point" {
		m.sendMessagef("No (%s - %s remaining)", currMap, fmtDuration(timeDiff))
	} else {
		m.sendMessagef("Yes (%s - %s remaining)", currMap, fmtDuration(timeDiff))
	}

	return
}

func (m *messageEvent) choose(args []string) {
	if len(args) < 2 {
		m.sendMessagef("Use: %schoose \"choice1\" \"choice2\"...", rinako.config.Discriminator)
	} else {
		choiceStr := strings.Join(args, " ")
		r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
		arr := r.FindAllString(choiceStr, -1)

		choice := rand.Intn(len(arr))
		m.sendMessagef("I choose %s", arr[choice])
	}
	//foreach:
}

func createLater(now time.Time, hr string, min string) time.Time {
	hrInt, _ := strconv.ParseInt(hr, 0, 64)
	minInt, _ := strconv.ParseInt(min, 0, 64)
	res := time.Date(now.Year(), now.Month(), now.Day(), int(hrInt), int(minInt), 00, 00, now.UTC().Location())
	//The max time diff is 2 hours, so this should suffice...
	if (24-now.Hour() <= 2) && (hrInt <= 1) {
		res = res.AddDate(0, 0, 1)
	}

	return res
}

type MediaExtended struct {
	AltText        *string `json:"altText"`
	DurationMillis int     `json:"duration_millis"`
	Size           Size    `json:"size"`
	ThumbnailURL   string  `json:"thumbnail_url"`
	Type           string  `json:"type"`
	URL            string  `json:"url"`
}

type Size struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type Tweet struct {
	CommunityNote       *string         `json:"communityNote"`
	ConversationID      string          `json:"conversationID"`
	Date                string          `json:"date"`
	DateEpoch           int             `json:"date_epoch"`
	Hashtags            []string        `json:"hashtags"`
	Likes               int             `json:"likes"`
	MediaURLs           []string        `json:"mediaURLs"`
	MediaExtended       []MediaExtended `json:"media_extended"`
	PossiblySensitive   bool            `json:"possibly_sensitive"`
	QrtURL              *string         `json:"qrtURL"`
	Replies             int             `json:"replies"`
	Retweets            int             `json:"retweets"`
	Text                string          `json:"text"`
	TweetID             string          `json:"tweetID"`
	TweetURL            string          `json:"tweetURL"`
	UserName            string          `json:"user_name"`
	UserProfileImageURL string          `json:"user_profile_image_url"`
	UserScreenName      string          `json:"user_screen_name"`
}

type VideoInfo struct {
	Needed        bool
	SwitchToRawVX bool
	URL           string
}

// from https://gist.github.com/Necroforger/8b0b70b1a69fa7828b8ad6387ebb3835

// Constants for message embed character limits
const (
	EmbedLimitTitle       = 256
	EmbedLimitDescription = 2048
	EmbedLimitFieldValue  = 1024
	EmbedLimitFieldName   = 256
	EmbedLimitField       = 25
	EmbedLimitFooter      = 2048
	EmbedLimit            = 4000
)

type Embed struct {
	*discordgo.MessageEmbed
}

//NewEmbed returns a new embed object
func NewEmbed() *Embed {
	return &Embed{&discordgo.MessageEmbed{}}
}

//SetTitle ...
func (e *Embed) SetTitle(name string) *Embed {
	e.Title = name
	return e
}

//SetDescription [desc]
func (e *Embed) SetDescription(description string) *Embed {
	if len(description) > 2048 {
		description = description[:2048]
	}
	e.Description = description
	return e
}

//AddField [name] [value]
func (e *Embed) AddField(name, value string) *Embed {
	if len(value) > 1024 {
		value = value[:1024]
	}

	if len(name) > 1024 {
		name = name[:1024]
	}

	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:  name,
		Value: value,
	})

	return e

}

//SetFooter [Text] [iconURL]
func (e *Embed) SetFooter(args ...string) *Embed {
	iconURL := ""
	text := ""
	proxyURL := ""

	switch {
	case len(args) > 2:
		proxyURL = args[2]
		fallthrough
	case len(args) > 1:
		iconURL = args[1]
		fallthrough
	case len(args) > 0:
		text = args[0]
	case len(args) == 0:
		return e
	}

	e.Footer = &discordgo.MessageEmbedFooter{
		IconURL:      iconURL,
		Text:         text,
		ProxyIconURL: proxyURL,
	}

	return e
}

//SetImage ...
func (e *Embed) SetImage(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Image = &discordgo.MessageEmbedImage{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

//SetVideo ...
func (e *Embed) SetVideo(args ...string) *Embed {
	var URL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	e.Video = &discordgo.MessageEmbedVideo{
		URL: URL,
	}
	return e
}

//SetThumbnail ...
func (e *Embed) SetThumbnail(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

//SetAuthor ...
func (e *Embed) SetAuthor(args ...string) *Embed {
	var (
		name     string
		iconURL  string
		URL      string
		proxyURL string
	)

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		name = args[0]
	}
	if len(args) > 1 {
		iconURL = args[1]
	}
	if len(args) > 2 {
		URL = args[2]
	}
	if len(args) > 3 {
		proxyURL = args[3]
	}

	e.Author = &discordgo.MessageEmbedAuthor{
		Name:         name,
		IconURL:      iconURL,
		URL:          URL,
		ProxyIconURL: proxyURL,
	}

	return e
}

//SetURL ...
func (e *Embed) SetURL(URL string) *Embed {
	e.URL = URL
	return e
}

//SetColor ...
func (e *Embed) SetColor(clr int) *Embed {
	e.Color = clr
	return e
}

// InlineAllFields sets all fields in the embed to be inline
func (e *Embed) InlineAllFields() *Embed {
	for _, v := range e.Fields {
		v.Inline = true
	}
	return e
}

// Truncate truncates any embed value over the character limit.
func (e *Embed) Truncate() *Embed {
	e.TruncateDescription()
	e.TruncateFields()
	e.TruncateFooter()
	e.TruncateTitle()
	return e
}

// TruncateFields truncates fields that are too long
func (e *Embed) TruncateFields() *Embed {
	if len(e.Fields) > 25 {
		e.Fields = e.Fields[:EmbedLimitField]
	}

	for _, v := range e.Fields {

		if len(v.Name) > EmbedLimitFieldName {
			v.Name = v.Name[:EmbedLimitFieldName]
		}

		if len(v.Value) > EmbedLimitFieldValue {
			v.Value = v.Value[:EmbedLimitFieldValue]
		}

	}
	return e
}

// TruncateDescription ...
func (e *Embed) TruncateDescription() *Embed {
	if len(e.Description) > EmbedLimitDescription {
		e.Description = e.Description[:EmbedLimitDescription]
	}
	return e
}

// TruncateTitle ...
func (e *Embed) TruncateTitle() *Embed {
	if len(e.Title) > EmbedLimitTitle {
		e.Title = e.Title[:EmbedLimitTitle]
	}
	return e
}

// TruncateFooter ...
func (e *Embed) TruncateFooter() *Embed {
	if e.Footer != nil && len(e.Footer.Text) > EmbedLimitFooter {
		e.Footer.Text = e.Footer.Text[:EmbedLimitFooter]
	}
	return e
}
