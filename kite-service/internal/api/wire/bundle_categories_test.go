package wire

import "testing"

func TestBundleListingCategories(t *testing.T) {
	listing := func(categories ...string) BundleListing {
		return BundleListing{Name: "Mod", Icon: "gavel", Categories: categories, About: []string{"Bans members."}}
	}
	for name, test := range map[string]struct {
		listing BundleListing
		ok      bool
	}{
		"one":      {listing("moderation"), true},
		"three":    {listing("moderation", "ai", "utility"), true},
		"too many": {listing("moderation", "ai", "utility", "games"), false},
		"none":     {listing(), false},
		"unknown":  {listing("moderation", "nope"), false},
		"twice":    {listing("moderation", "ai", "ai"), false},
		"empty":    {listing("moderation", ""), false},
		// Clients from before send only the category.
		"only category": {BundleListing{Name: "Mod", Icon: "gavel", Category: "ai", About: []string{"Bans members."}}, true},
	} {
		if err := test.listing.validate(true); (err == nil) != test.ok {
			t.Errorf("%s: validate() = %v, want ok %v", name, err, test.ok)
		}
	}

	l := BundleListingFromWire(listing("ai", "games"))
	if l.Category != "ai" {
		t.Errorf("Category = %q, want the first of the categories", l.Category)
	}
}
