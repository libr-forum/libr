package alias

import (
	"crypto/sha256"
	"fmt"
)

var adjectives = []string{
	"Swift", "Brave", "Calm", "Eager", "Fuzzy", "Gentle", "Happy", "Jolly", "Kind", "Lively",
	"Merry", "Nice", "Proud", "Quick", "Royal", "Shiny", "Tidy", "Upbeat", "Vivid", "Witty",
	"Zany", "Bold", "Clever", "Daring", "Elegant", "Fierce", "Graceful", "Honest", "Keen", "Loyal",
	"Noble", "Peppy", "Quirky", "Sassy", "Truthful", "Unique", "Vibrant", "Warm", "Xenial", "Young",
}

var nouns = []string{
	"Panda", "Eagle", "Lion", "Tiger", "Wolf", "Otter", "Fox", "Bear", "Whale", "Shark",
	"Zebra", "Rhino", "Monkey", "Horse", "Koala", "Giraffe", "Dragon", "Peacock", "Falcon", "Owl",
	"Cheetah", "Crane", "Beetle", "Dolphin", "Lynx", "Mouse", "Puppy", "Quokka", "Raven", "Swan",
	"Turtle", "Urchin", "Vulture", "Weasel", "Yak", "Ibis", "Jackal", "Lemur", "Moose", "Newt",
}

func GenerateAlias(key string) string {
	if len(key) != 44 {
		return "InvalidKeyLength"
	}

	hash := sha256.Sum256([]byte(key))

	adjIndex := int(hash[0]) % len(adjectives)
	nounIndex := int(hash[1]) % len(nouns)
	num := (int(hash[2])<<8 + int(hash[3])) % 10000

	return fmt.Sprintf("%s%s%04d", adjectives[adjIndex], nouns[nounIndex], num)
}
