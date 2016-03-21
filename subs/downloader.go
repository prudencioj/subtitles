package subs

import (
	"net/http"
	"fmt"
	"path/filepath"
	"strings"
	"os"
	"github.com/prudencioj/subtitles/subs/subdb"
	"errors"
)

var (
// Supported extensions in a map.
	extensions = map[string]bool{
		".3g2":true, ".3gp":true, ".3gp2":true, ".3gpp":true, ".60d":true, ".xvid": true,
		".ajp":true, ".asf":true, ".asx":true, ".avchd":true, ".avi":true, ".bik":true, ".bix":true,
		".box":true, ".cam":true, ".dat":true, ".divx":true, ".dmf":true, ".dv":true, ".dvr-ms":true,
		".evo":true, ".flc":true, ".fli":true, ".flic":true, ".flv":true, ".flx":true, ".gvi":true,
		".gvp":true, ".h264":true, ".m1v":true, ".m2p":true, ".m2ts":true, ".m2v":true, ".m4e":true,
		".m4v":true, ".mjp":true, ".mjpeg":true, ".mjpg":true, ".mkv":true, ".moov":true, ".mov":true,
		".movhd":true, ".movie":true, ".movx":true, ".mp4":true, ".mpe":true, ".mpeg":true, ".mpg":true,
		".mpv":true, ".mpv2":true, ".mxf":true, ".nsv":true, ".nut":true, ".ogg":true, ".ogm":true,
		".omf":true, ".ps":true, ".qt":true, ".ram":true, ".rm":true, ".rmvb":true, ".swf":true, ".ts":true,
		".vfw":true, ".vid":true, ".video":true, ".viv":true, ".vivo":true, ".vob":true, ".vro":true,
		".wm":true, ".wmv":true, ".wmx":true, ".wrap":true, ".wvx":true, ".wx":true, ".x264":true,
	}

// Default Languages
	defaultLanguages = []string{"en"}

// Errors
	ErrNoSubtitleFound = errors.New("subs: no subtitle found")
	ErrSubtitleDownload = errors.New("subs: failed to download")
	ErrSubtitleSave = errors.New("subs: failed to save subtitle")
)

type Downloader struct {
	client *subdb.SubDB
}

type DownloadResult struct {
	Video    string
	Subtitle string
	Language string
	Error    error
}

func NewDownloader() (*Downloader) {
	d := Downloader{
		client: subdb.NewSubDB(http.DefaultClient),
	}
	return &d
}

func (d *Downloader) Download(p string, langs []string) []*DownloadResult {
	videos := searchVideos(p)
	c := make(chan *DownloadResult)
	subs := make([]*DownloadResult, 0)

	// If no language preference was provided, use the default
	sLangs := make([]string, 0)
	if len(langs) == 0 {
		sLangs = defaultLanguages
	} else {
		sLangs = langs
	}

	nDownloads := len(sLangs) * len(videos)
	if nDownloads == 0 {
		return subs
	}

	// Start a go routine for each video and language
	for _, v := range videos {
		for _, l := range sLangs {
			go d.download(v, l, c)
		}
	}

	// Use the channel to receive the results
	for res := range c {
		fmt.Println(res)
		subs = append(subs, res)
		// When we get all the answers, we can terminate and return all the subs.
		if len(subs) == nDownloads {
			close(c)
			return subs
		}
	}
	return subs
}

func (d *Downloader) download(p string, preflang string, c chan *DownloadResult) {
	// Result to be sent once to the channel
	res := &DownloadResult{Video: p, Language:preflang}
	defer func() {
		if r := recover(); r != nil {
			e, _ := r.(error)
			res.Error = e
		}
		c <- res
	}()

	// TODO use different providers
	// Search for available languages
	s := d.client
	langs, err := s.Search(p)
	if err != nil {
		panic(ErrNoSubtitleFound)
	}

	// Check if the languages found match the preferred language
	var lang string
	for _, l := range langs {
		// TODO logic to compare, different providers may have different country codes
		if l == preflang {
			lang = preflang
			break
		}
	}

	if lang == "" {
		panic(ErrNoSubtitleFound)
	}

	// Download subtitle if it found something
	subs, err := s.Download(p, lang)
	if err != nil {
		panic(ErrSubtitleDownload)
	}

	// Save subtitle to disk
	ve := filepath.Ext(p)
	dst := p[:strings.Index(p, ve)]

	res.Subtitle = fmt.Sprintf("%s.%s%s", dst, lang, subs.Extension)
	f, err := os.Create(res.Subtitle)
	if err == nil {
		f.WriteString(subs.Content)
		f.Close()
	} else {
		panic(ErrSubtitleSave)
	}
}

func searchVideos(root string) []string {
	var paths []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if extensions[ext] {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}