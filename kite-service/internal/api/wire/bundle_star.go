package wire

// BundleStarResponse has how many users starred the bundle now, which pages
// rendered before don't know.
type BundleStarResponse struct {
	StarCount int `json:"star_count"`
}

type BundleUnstarResponse = BundleStarResponse

// StarredBundleListResponse are the IDs of the bundles the user starred,
// newest first.
type StarredBundleListResponse = []string
