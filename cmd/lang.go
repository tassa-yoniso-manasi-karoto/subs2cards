package cmd

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	iso "github.com/barbashov/iso639-3"
)

const (
	StrippedSDH = iota
	Sub
	Dub
	CC
)

type Lang struct {
	*iso.Language
	// Typically a ISO 3166-1 region but can also be a ISO 15924 script
	Subtag string
}



var refmatch = map[string]int{
	"closedcaptions": CC,
	"dubtitles":      Dub,
	"subtitles":      Sub,
	"dialog":         Sub,
	"stripped_sdh":   StrippedSDH,
}

func ReadStdLangCode(arr []string) (langs []Lang, err error) {
	for _, tmp := range arr {
		var lang Lang
		arr := strings.Split(tmp, "-")
		lang.Language = iso.FromAnyCode(arr[0])
		if lang.Language == nil {
			return nil, fmt.Errorf("An invalid language code was passed: '%s'", arr[0])
		}
		if len(arr) > 1 {
			lang.Subtag = arr[1]
		}
		langs = append(langs, lang)
	}
	return
}

func SetPrefered(langs []Lang, l, atm Lang, name string, out *string) bool {
	if setPreferedLang(langs, l, atm) && isPreferedSubtypeOver(*out, name) {
		*out = name
		//color.Yellowln(setPreferedLang(langs, l, atm), isPreferedSubtypeOver(*out, name), "out becomes", name)
		return true
	}
	return false
}

// Only 1st subtag found is considered, the others are ignored
func GuessLangFromFilename(name string) (lang Lang, err error) {
	// this int was in the original algo, not sure if I need it or not at some point
	var fn_start int
	l := guessLangFromFilename(name, &fn_start)
	if arr := strings.Split(l, "-"); strings.Contains(l, "-") {
		lang.Subtag = arr[1]
		l = arr[0]
	}
	lang.Language = iso.FromAnyCode(l)
	if lang.Language == nil {
		err = fmt.Errorf("No language could be identified.")
	}
	return
}

/*
This is a slightly improved version of mpv's algorithm to parse langs from filename available here:
https://github.com/mpv-player/mpv/blob/a6c363e8dac1e0fd932b1a17838e0da3b13f3b4c/misc/language.c#L300

While none of their code was directly reused, I'd rather give attribution to the authors nonetheless.
The C code they provide is under GNU Lesser General Public License LGPL version 2 or later:
https://github.com/mpv-player/mpv/blob/master/LICENSE.LGPL

The implementation here is provided under GPL3 as the rest of this project.
You can test if your subs are found by their algo with "mpv --msg-level=find_files=trace video.mp4"

*/
func guessLangFromFilename(name string, langStart *int) string {
	stripname := name
	var ok bool
	var i, langLength int
	// this iter decorticates some more in case lang isn't located at the end of the name
	for x := 0; x < 2; x++ {
		stripname = strings.TrimSuffix(path.Base(stripname), path.Ext(stripname))
		stripname = strings.TrimSpace(stripname)

		if len(stripname) < 2 {
			return ""
		}

		langLength = 0
		i = len(stripname) - 1
		suffixesLength := 0

		delimiter := '.'
		if stripname[i] == ')' {
			delimiter = '('
			i--
		}
		if stripname[i] == ']' {
			delimiter = '['
			i--
		}
		ok = true
		for {
			// Identify alphabetic characters for the language tag
			for i >= 0 && unicode.IsLetter(rune(stripname[i])) {
				langLength++
				i--
			}

			// Validate the length of subtags
			if langLength < suffixesLength+1 || langLength > suffixesLength+8 {
				ok = false // invalid length, return empty
			}

			// Check for subtag separator '-'
			if i >= 0 && stripname[i] == '-' {
				langLength++
				i--
				suffixesLength = langLength
			} else {
				//println(stripname[i+1 : i+1+langLength])
				break
			}
		}
		// Validate the primary subtag's length (2-3 letters)
		if langLength < suffixesLength+2 || langLength > suffixesLength+3 || i <= 0 || rune(stripname[i]) != delimiter {
			//color.Yellowln("langLength=", langLength, stripname[i+1 : i+1+langLength])
			//color.Yellowln("rune(stripname[i])=", string(stripname[i]))
			//color.Redln(langLength < suffixesLength+2, langLength > suffixesLength+3, i <= 0, rune(name[i]) != delimiter)
			ok = false
		}
		// if longer than 3 letters, must have hyphen and subtag
		if langLength > suffixesLength+3 && !strings.Contains(stripname[i+1:i+1+langLength], "-") {
			//color.Yellowln(string(stripname[i+1 : i+1+langLength]), "contains no '-'!")
			ok = false
		}
		if ok {
			break
		}
	}
	if !ok {
		return ""
	}

	// Set the starting position of the language tag
	if langStart != nil {
		*langStart = i
	}

	// Return the detected language tag as a substring
	return name[i+1 : i+1+langLength]
}



func setPreferedLang(langs []Lang, l, atm Lang) (b bool) {
	//println(fmt.Sprintf("%#v", l), "l idx", getIdx(langs, l))
	//println(fmt.Sprintf("%#v", atm), "atm idx", getIdx(langs, atm))
	langIdx, langIsDesired := getIdx(langs, l)
	atmIdx, _ := getIdx(langs, atm)
	// idx = idx in the row of lang sorted by preference the user has passed
	if langIsDesired && langIdx <= atmIdx {
		b = true
	}
	return
}

func getIdx(langs []Lang, ref Lang) (int, bool) {
	for i, l := range langs {
		if l.Part3 == ref.Part3 && l.Subtag == ref.Subtag {
			return i, true
		}
	}
	return 0, false
}

func isPreferedSubtypeOver(curr, test string) bool {
	currval := subtypeMatcher(curr)
	testval := subtypeMatcher(test)
	//println(testval, ">", currval, "IS", testval > currval)
	return testval > currval
}

func subtypeMatcher(s string) int {
	if s == "" {
		return 0
	}
	s = strings.ToLower(s)
	for subtype, v := range refmatch {
		if strings.Contains(s, subtype) {
			return v
		}
	}
	// If nothing is specified in the subtitle file name then it's probably a regular Sub file
	return Sub
}


func placeholder34564() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
