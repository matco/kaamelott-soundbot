package kaamelott

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

//retrieve list of sounds from Github instead of Kaamelott soundboard website because their URL is regularly modified
//const KaamelottSoundsURL = "https://kaamelott-soundboard.2ec0b4.fr/sounds/sounds.a7b9de88.json"
const KaamelottSoundsURL = "https://raw.githubusercontent.com/2ec0b4/kaamelott-soundboard/master/sounds/sounds.json"

//baselink of the sound on Kaamelott soundboard website
const KaamelottSoundURL = "https://kaamelott-soundboard.2ec0b4.fr/#son/"

type Sound struct {
	Character string `json:"character"`
	Episode   string `json:"episode"`
	Title     string `json:"title"`
	File      string `json:"file"`
	Words     []string
}

type Match struct {
	Index int
	Sound *Sound
}

type SlackMessage struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

var sounds []Sound

func init() {
	http.HandleFunc("/", handler)
}

func retrieve_sounds(ctx context.Context) {
	client := urlfetch.Client(ctx)
	//retrieve current sounds
	resp, err := client.Get(KaamelottSoundsURL)
	if err == nil {
		json.NewDecoder(resp.Body).Decode(&sounds)
		log.Infof(ctx, "Found %d Kaamelott sounds", len(sounds))

		//build words list for each sounds to search efficiently
		for i := 0; i < len(sounds); i++ {
			sound := sounds[i]
			sound.Words = strings.Split(sound.Title, " ")
		}
		log.Infof(ctx, "Retrieve all words related to sound")
	} else {
		log.Errorf(ctx, err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ctx := appengine.NewContext(r)
		//build sounds cache if required
		//this must be done in the context of a request, even if the cache will then be shared by all following requests
		//use the first request to build this cache
		if len(sounds) == 0 {
			retrieve_sounds(ctx)
		}
		r.ParseForm()
		//retrieve query stored in the "text" variable
		text := r.Form.Get("text")
		var query []string
		if text != "" {
			query = strings.Split(text, " ")
		}
		if len(query) > 0 {
			//extract command with is first word of query
			command := query[0]
			//extract command arguments
			arguments := query[1:]
			log.Infof(ctx, "Executing command [%s] with arguments %s", command, arguments)
			//executing command
			switch command {
			case "search":
				if len(arguments) == 1 {
					search := arguments[0]
					log.Debugf(ctx, "Searching sound with search [%s]", search)
					var matches []Match
					//look for exact match
					for i := 0; i < len(sounds) && len(matches) < 5; i++ {
						sound := sounds[i]
						for j := 0; j < len(sound.Words); j++ {
							if search == sound.Words[j] {
								matches = append(matches, Match{Index: i, Sound: &sound})
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
							matches = append(matches, Match{Index: i, Sound: &sound})
						}
					}
					log.Infof(ctx, "Found %d sounds", len(matches))
					//transform matches into a multiple line string
					message := ""
					for i := 0; i < len(matches); i++ {
						message += fmt.Sprintf("%v: %s\n", matches[i].Index, matches[i].Sound.Title)
					}
					var response = SlackMessage{ResponseType: "ephemeral", Text: message}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				} else if len(arguments) > 1 {
					fmt.Fprintf(w, "Only one search parameter is supported")
				} else {
					fmt.Fprintf(w, "Add search parameter")
				}
			case "play":
				if len(arguments) > 0 {
					id, err := strconv.Atoi(arguments[0])
					if err == nil {
						log.Infof(ctx, "Adding sound with id [%v]", id)
						file := sounds[id].File
						message := fmt.Sprintf("%s%s", KaamelottSoundURL, file[0:len(file)-4])
						var response = SlackMessage{ResponseType: "in_channel", Text: message}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(response)
					} else {
						fmt.Fprint(w, "Wrong id")
					}
				} else {
					fmt.Fprintf(w, "Add sound id")
				}
			case "random":
				id := rand.Intn(len(sounds))
				file := sounds[id].File
				message := fmt.Sprintf("%s%s", KaamelottSoundURL, file[0:len(file)-4])
				var response = SlackMessage{ResponseType: "in_channel", Text: message}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				fmt.Fprintf(w, "Unsupported command %s", command)
			}
		} else {
			fmt.Fprint(w, "Empty command")
		}
	} else {
		fmt.Fprint(w, "Nothing here")
	}
}
