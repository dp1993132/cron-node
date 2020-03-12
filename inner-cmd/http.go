package inner_cmd

import (
	"io/ioutil"
	"net/http"
)

func Get(url string) ([]byte, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return buf, resp.StatusCode, err
}
