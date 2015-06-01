// Spotify Player

package soundwave

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"
	"time"

	"github.com/op/go-libspotify/spotify"
)

func NewSession(user *string, pass *string, key *string) (*spotify.Session, *audioWriter) {
	debug := true

	appKey, err := ioutil.ReadFile(*key)
	if err != nil {
		log.Fatal(err)
	}

	var silenceStderr = DiscardFd(syscall.Stderr)
	if debug == true {
		silenceStderr.Restore()
	}

	audio, err := newAudioWriter()
	if err != nil {
		log.Fatal(err)
	}
	silenceStderr.Restore()

	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   appKey,
		ApplicationName:  "SOON_ FM",
		CacheLocation:    "tmp",
		SettingsLocation: "tmp",
		AudioConsumer:    audio,

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	credentials := spotify.Credentials{
		Username: *user,
		Password: *pass,
	}
	if err = session.Login(credentials, true); err != nil {
		log.Fatal(err)
	}

	// Set Bitrate
	session.PreferredBitrate(spotify.Bitrate320k)

	// Log messages
	if debug {
		go func() {
			for msg := range session.LogMessages() {
				log.Print(msg)
			}
		}()
	}

	// Wait for login and expect it to go fine
	select {
	case err = <-session.LoggedInUpdates():
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Session Created")
	log.Println(session.ConnectionState())

	return session, audio
}

func Play(session *spotify.Session, id *string) {
	log.Println(session.ConnectionState())

	uri := fmt.Sprintf("spotify:track:%s", *id)
	log.Println(uri)

	// Parse the track
	link, err := session.ParseLink(uri)
	if err != nil {
		log.Fatal(err)
	}
	track, err := link.Track()
	if err != nil {
		log.Fatal(err)
	}

	// Load the track and play it
	track.Wait()
	player := session.Player()
	if err := player.Load(track); err != nil {
		fmt.Println("%#v", err)
		log.Fatal(err)
	}
	defer player.Unload()

	log.Println("Playing...")
	player.Play()

	c1 := time.Tick(time.Millisecond)
	now := time.Now()
	start := now

	for {
		now = <-c1
		elapsed := now.Sub(start)
		if elapsed >= track.Duration() {
			break
		}
	}

	log.Println("End of Track")

	<-session.EndOfTrackUpdates()

	log.Println(session.ConnectionState())
}
