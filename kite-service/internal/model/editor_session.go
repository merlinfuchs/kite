package model

import "gopkg.in/guregu/null.v4"

// EditorSession is one opened editor of a command or event listener. Exactly
// one of CommandID and EventListenerID is set.
type EditorSession struct {
	ID              string
	AppID           string
	UserID          string
	CommandID       null.String
	EventListenerID null.String
}

// EditorUsage sums the editor sessions of a command or event listener.
type EditorUsage struct {
	CommandID       null.String
	EventListenerID null.String
	ActiveSeconds   int
	SaveCount       int
	LastSavedAt     null.Time
}
