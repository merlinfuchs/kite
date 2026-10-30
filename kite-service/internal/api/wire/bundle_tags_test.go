package wire

import (
	"strings"
	"testing"
)

func TestBundleListingTags(t *testing.T) {
	listing := func(tags ...string) BundleListing {
		return BundleListing{Name: "Mod", Icon: "gavel", Category: "moderation", Tags: tags, About: []string{"Bans members."}}
	}
	for name, test := range map[string]struct {
		listing BundleListing
		ok      bool
	}{
		"none":         {listing(), true},
		"some":         {listing("roblox", "economy", "ticket-system", "web3"), true},
		"too many":     {listing("a1", "b2", "c3", "d4", "e5", "f6"), false},
		"twice":        {listing("roblox", "roblox"), false},
		"uppercase":    {listing("Roblox"), false},
		"space":        {listing("ticket system"), false},
		"double dash":  {listing("ticket--system"), false},
		"edge dash":    {listing("-roblox"), false},
		"too short":    {listing("a"), false},
		"too long":     {listing(strings.Repeat("a", MaxBundleTagLength+1)), false},
		"empty":        {listing(""), false},
		"emoji":        {listing("🎮"), false},
		"longest fits": {listing(strings.Repeat("a", MaxBundleTagLength)), true},
	} {
		if err := test.listing.validate(true); (err == nil) != test.ok {
			t.Errorf("%s: validate() = %v, want ok %v", name, err, test.ok)
		}
	}
}
