package subdb

import (
	"net/http"
	"io/ioutil"
	"crypto/md5"
	"os"
	"io"
	"encoding/hex"
	"strings"
	"errors"
	"net/url"
)

const (
	server = "api.thesubdb.com"
	userAgent = "SubDB/1.0 (Pyrrot/0.1; http://github.com/jrhames/pyrrot-cli)"
)

type SubDB struct {
	HttpClient *http.Client
	Endpoint   string
	UserAgent  string
}

type Subtitle struct {
	Content   string
	Extension string
}

func NewSubDB(c *http.Client) (*SubDB) {
	client := &SubDB{
		HttpClient: c,
		Endpoint: server,
		UserAgent: userAgent,
	}
	return client
}

func (c *SubDB) Search(path string) ([]string, error) {
	hash, err := hash(path)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("action", "search")
	params.Set("hash", hash)
	url := url.URL{
		Scheme:   "http",
		Host:     c.Endpoint,
		RawQuery: params.Encode(),
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(body), ","), nil
}

func (c *SubDB) Download(path string, lang string) (*Subtitle, error) {
	hash, err := hash(path)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("action", "download")
	params.Set("hash", hash)
	params.Set("language", lang)
	url := url.URL{
		Scheme:   "http",
		Host:     c.Endpoint,
		RawQuery: params.Encode(),
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Get subtitle file extension from the header
	cnt := resp.Header.Get("Content-Disposition")
	if cnt == "" {
		// FIXME create new error type
		return nil, errors.New("No subtitle found.")
	}
	ext := cnt[strings.Index(cnt, "."):]

	return &Subtitle{string(body), ext}, nil
}

// http://thesubdb.com/api/
// The hash function is the core of our database system. You'll need to know the hash of the video file, either to download or upload subtitles.
// Our hash is composed by taking the first and the last 64kb of the video file, putting all together and generating a md5 of the resulting data (128kb).
func hash(path string) (string, error) {
	var size int64 = 64 * 1024
	hash := md5.New()
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.CopyN(hash, f, size)
	if err != nil {
		return "", err
	}

	_, err = f.Seek(-size, os.SEEK_END)
	if err != nil {
		return "", err
	}

	_, err = io.CopyN(hash, f, size)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}