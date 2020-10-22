package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
)

//retrieve list of sounds from Github instead of Kaamelott soundboard website because their URL is regularly modified
//const kaamelottSoundsURL = "https://kaamelott-soundboard.2ec0b4.fr/sounds/sounds.a7b9de88.json"
const kaamelottSoundsURL = "https://raw.githubusercontent.com/2ec0b4/kaamelott-soundboard/master/sounds/sounds.json"

//baselink of the sound on Kaamelott soundboard website
const kaamelottSoundURL = "https://kaamelott-soundboard.2ec0b4.fr/#son/"

type sound struct {
	Character string `json:"character"`
	Episode   string `json:"episode"`
	Title     string `json:"title"`
	File      string `json:"file"`
	Words     []string
}

type match struct {
	Index int
	Sound *sound
}

type slackMessage struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

var sounds []sound

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func retrieveSounds() {
	//retrieve current sounds
	resp, err := http.Get(kaamelottSoundsURL)
	if err == nil {
		json.NewDecoder(resp.Body).Decode(&sounds)
		log.Printf("Found %d Kaamelott sounds", len(sounds))

		//build words list for each sounds to search efficiently
		for i := 0; i < len(sounds); i++ {
			sound := sounds[i]
			sound.Words = strings.Split(sound.Title, " ")
		}
		log.Printf("Retrieve all words related to sound")
	} else {
		log.Printf(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		//build sounds cache if required
		//this must be done in the context of a request, even if the cache will then be shared by all following requests
		//use the first request to build this cache
		if len(sounds) == 0 {
			retrieveSounds()
		}
		r.ParseForm()
		//retrieve interactive action
		messageType := r.Form.Get("type")
		if messageType == "interactive_message" {

		}
		//retrieve query stored in the "text" variable
		text := r.Form.Get("text")
		var query []string
		if text != "" {
			query = strings.Split(text, " ")
		}
		//if the is no command or command is help, help will be displayed
		var displayHelp = false
		if len(query) > 0 {
			//extract command with is first word of query
			command := query[0]
			//extract command arguments
			arguments := query[1:]
			log.Printf("Executing command [%s] with arguments %s", command, arguments)
			//executing command
			switch command {
			case "help":
			case "h":
				displayHelp = true
			case "search":
			case "s":
				log.Printf(strconv.Itoa(len(arguments)))
				if len(arguments) == 1 {
					search := arguments[0]
					log.Printf("Searching sound with search [%s]", search)
					var matches []match
					//look for exact match
					for i := 0; i < len(sounds) && len(matches) < 5; i++ {
						sound := sounds[i]
						for j := 0; j < len(sound.Words); j++ {
							if search == sound.Words[j] {
								matches = append(matches, match{Index: i, Sound: &sound})
							}
						}
					}
					//look for approximate match
					for i := 0; i < len(sounds) && len(matches) < 5; i++ {
						sound := sounds[i]
						//check that this sound as not already been selected
						var selected = false
						for j := 0; j < len(matches); j++ {
							if matches[j].Sound == &sound {
								selected = true
								break
							}
						}
						if !selected && strings.Contains(sound.Title, search) {
							matches = append(matches, match{Index: i, Sound: &sound})
						}
					}
					log.Printf("Found %d sounds", len(matches))
					//transform matches into a multiple line string
					var response interface{}
					if len(matches) > 0 {
						var sections []slack.Block
						for i := 0; i < len(matches); i++ {
							buttonText := slack.NewTextBlockObject("plain_text", "Play", false, false)
							button := slack.NewButtonBlockElement("play", strconv.Itoa(i), buttonText)
							accessory := slack.NewAccessory(button)
							text := slack.NewTextBlockObject("mrkdwn", matches[i].Sound.Title, false, false)
							section := slack.NewSectionBlock(text, nil, accessory)
							sections = append(sections, *section)
						}
						response = slack.MsgOptionBlocks(sections...)
					} else {
						response = slackMessage{ResponseType: "ephemeral", Text: "No match"}
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				} else if len(arguments) > 1 {
					fmt.Fprintf(w, "Only one search parameter is supported")
				} else {
					fmt.Fprintf(w, "Add search parameter")
				}
			case "play":
			case "p":
				if len(arguments) > 0 {
					id, err := strconv.Atoi(arguments[0])
					if err == nil {
						log.Printf("Adding sound with id [%v]", id)
						file := sounds[id].File
						message := fmt.Sprintf("%s%s", kaamelottSoundURL, file[0:len(file)-4])
						var response = slackMessage{ResponseType: "in_channel", Text: message}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(response)
					} else {
						fmt.Fprint(w, "Wrong id")
					}
				} else {
					fmt.Fprintf(w, "Add sound id")
				}
			case "random":
			case "r":
				id := rand.Intn(len(sounds))
				file := sounds[id].File
				message := fmt.Sprintf("%s%s", kaamelottSoundURL, file[0:len(file)-4])
				var response = slackMessage{ResponseType: "in_channel", Text: message}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				fmt.Fprintf(w, "Unsupported command %s", command)
			}
		} else {
			displayHelp = true
		}
		if displayHelp == true {
			message := "Commands help\n"
			message += "help or h: this message\n"
			message += "search or s <search>: search for a sound related to <search>\n"
			message += "play or p <id>: add a link to the sound <id>\n"
			message += "random or r: add a link to a random sound\n"
			var response = slackMessage{ResponseType: "ephemeral", Text: message}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	} else {
		fmt.Fprint(w, "Nothing here")
	}
}
