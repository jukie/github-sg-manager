package job

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func githubHookCIDRs() []string {
	resp, err := http.Get("https://api.github.com/meta")
	if err != nil {
		log.Fatalln("Error loading Github CIDRs")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var data struct {
		Hooks []string //we only really care about the Hooks
	}
	json.Unmarshal(body, &data)
	fmt.Printf("Valid 'Hooks' CIDRs response from https://api.github.com/meta:\n	%s\n\n\n", data.Hooks)
	return data.Hooks
}
