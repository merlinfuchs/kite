package message

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/message"
)

func TestEvalPlaceholders(t *testing.T) {
	sentAt := time.Unix(1700000000, 0)
	data := &message.MessageData{
		Content: "<t:{{now().Unix()}}:R> <t:{{now().Unix() + 60}}:R> {{channel.id}} {{server.id}} {{user.mention}}",
	}

	res, err := evalPlaceholders(context.Background(), data, "1", "2", sentAt)
	if err != nil {
		t.Fatal(err)
	}

	want := fmt.Sprintf("<t:%d:R> <t:%d:R> 2 1 {{user.mention}}", sentAt.Unix(), sentAt.Unix()+60)
	if res.Content != want {
		t.Errorf("got %q, want %q", res.Content, want)
	}
	if data.Content == res.Content {
		t.Error("the template was changed instead of a copy")
	}
}
