package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

const source = "https://raw.githubusercontent.com/jie65535/awesome-balatro/refs/heads/main/README.md"

type Metadata struct {
	Mods []List
}

type List struct {
	Category string     `json:"category"`
	Items    []ListItem `json:"items"`
}

type ListItem struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Credit      string `json:"credit"`
	Url         string `json:"url"`
	GHType      string `json:"install_type"`
	LoaderType  string `json:"mod_loader"`
}

func main() {
	sections := make(map[string][]*List)

	lines := strings.Split(downloadLastestReadme(), "\n")
	currentSection := ""
	currentList := &List{}
	loaderType := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if strings.Contains(line, "Balamod") {
				continue
			}
			loaderType = "unknown"
			if strings.Contains(line, "Lovely") {
				loaderType = "lovely"
			}

			if strings.Contains(line, "SteamModded") {
				loaderType = "steammodded"
			}

			currentSection = line[3:]
			sections[currentSection] = make([]*List, 0)
		}

		if strings.HasPrefix(line, "### ") {
			// this is the start of a new subsection which denotes a different
			// category
			currentList = &List{
				Category: line[4:],
				Items:    make([]ListItem, 0),
			}

			sections[currentSection] = append(sections[currentSection], currentList)
		}

		if strings.HasPrefix(line, "- ") {
			li, skip := lineToListItem(line[2:])
			if skip {
				continue
			}
			li.LoaderType = loaderType
			currentList.Items = append(currentList.Items, li)
		}
	}

	// cleanup entries which have no items
	for k, v := range sections {
		if len(v) == 0 {
			delete(sections, k)
		}
	}

	md := Metadata{
		Mods: make([]List, 0),
	}

	for _, v := range sections {
		for _, l := range v {
			md.Mods = append(md.Mods, *l)
		}
	}

	yaml, err := yaml.Marshal(md)
	if err != nil {
		panic(err)
	}

	os.WriteFile("jesters.yaml", yaml, 0644)
}

func lineToListItem(line string) (ListItem, bool) {
	nameAndLink, descriptionAndCred, found := strings.Cut(line, " - ")
	if !found {
		return ListItem{}, true
	}

	name, link, _ := strings.Cut(nameAndLink[1:], "]")
	link = strings.ReplaceAll(link, "(", "")
	link, _, _ = strings.Cut(link, ")")

	if strings.Contains(link, "discord") || strings.Contains(link, "nexusmods") {
		return ListItem{}, true
	}

	GHType := "repo"
	if strings.Contains(link, "gist.github.com") {
		GHType = "gist"
	}

	l := ListItem{
		DisplayName: strings.TrimSpace(name),
		Url:         strings.TrimSpace(link),
		GHType:      GHType,
	}

	description, cred, found := strings.Cut(descriptionAndCred, "by")
	if found {
		l.Description = strings.TrimSpace(description)
		l.Credit = strings.TrimSpace(cred)
	}

	return l, false
}

func downloadLastestReadme() string {
	resp, err := http.DefaultClient.Get(source)
	if err != nil {
		panic(fmt.Errorf("failed to download readme: %w", err))
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read body: %w", err))
	}

	return string(b)
}
