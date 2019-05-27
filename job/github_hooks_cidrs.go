package job

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func githubHookCIDRs() ([]string, error) {
	resp, err := http.Get("https://api.github.com/meta")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var data struct {
		Hooks []string //we only really care about the Hooks
	}
	json.Unmarshal(body, &data)
	fmt.Printf("Valid 'Hooks' CIDRs response from https://api.github.com/meta:\n	%s\n\n\n", data.Hooks)
	return data.Hooks, nil
}
