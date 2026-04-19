package chat

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestHubPublishDeliversToSubscriber(t *testing.T) {
	hub := NewHub(zap.NewNop())
	ch, unsubscribe := hub.Subscribe("profile-a")
	defer unsubscribe()

	hub.Publish("profile-a", Event{Name: "msg", ID: "1", Data: "hello", ChatID: "c1"})

	select {
	case evt := <-ch:
		if evt.Data != "hello" {
			t.Fatalf("unexpected event payload: %q", evt.Data)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestHubPublishIsolatesProfiles(t *testing.T) {
	hub := NewHub(zap.NewNop())
	chA, unsubA := hub.Subscribe("profile-a")
	defer unsubA()
	chB, unsubB := hub.Subscribe("profile-b")
	defer unsubB()

	hub.Publish("profile-a", Event{Data: "for-a"})

	select {
	case evt := <-chA:
		if evt.Data != "for-a" {
			t.Fatalf("profile-a got wrong payload: %q", evt.Data)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("profile-a did not receive its event")
	}

	select {
	case evt := <-chB:
		t.Fatalf("profile-b should not have received event %q", evt.Data)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubPublishAllFanOut(t *testing.T) {
	hub := NewHub(zap.NewNop())
	chA, unsubA := hub.Subscribe("profile-a")
	defer unsubA()
	chB, unsubB := hub.Subscribe("profile-b")
	defer unsubB()

	hub.PublishAll([]string{"profile-a", "profile-b", "missing"}, Event{Data: "broadcast"})

	for _, ch := range []<-chan Event{chA, chB} {
		select {
		case evt := <-ch:
			if evt.Data != "broadcast" {
				t.Fatalf("unexpected payload: %q", evt.Data)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("subscriber did not receive broadcast event")
		}
	}
}

func TestHubUnsubscribeRemovesSubscriber(t *testing.T) {
	hub := NewHub(zap.NewNop())
	_, unsubscribe := hub.Subscribe("profile-a")

	if got := hub.SubscriberCount("profile-a"); got != 1 {
		t.Fatalf("expected 1 subscriber, got %d", got)
	}
	unsubscribe()
	if got := hub.SubscriberCount("profile-a"); got != 0 {
		t.Fatalf("expected 0 subscribers after unsubscribe, got %d", got)
	}
}

func TestHubDropsEventsWhenSubscriberIsSlow(t *testing.T) {
	hub := NewHub(zap.NewNop())
	hub.bufferSize = 2
	ch, unsubscribe := hub.Subscribe("profile-a")
	defer unsubscribe()

	// Fill the buffer (no reads happening) and then publish more. Publishes
	// must not block and excess events must be dropped.
	for i := 0; i < 10; i++ {
		hub.Publish("profile-a", Event{ID: string(rune(i))})
	}

	received := 0
	timeout := time.After(100 * time.Millisecond)
drain:
	for {
		select {
		case <-ch:
			received++
		case <-timeout:
			break drain
		}
	}
	if received == 0 {
		t.Fatal("expected to receive at least one event")
	}
	if received > 2 {
		t.Fatalf("expected at most buffer-size (2) events, got %d", received)
	}
}

func TestHubEvictsBeyondMaxPerProfile(t *testing.T) {
	hub := NewHub(zap.NewNop())
	hub.maxPerProfile = 2

	_, unsub1 := hub.Subscribe("profile-a")
	defer unsub1()
	_, unsub2 := hub.Subscribe("profile-a")
	defer unsub2()
	_, unsub3 := hub.Subscribe("profile-a")
	defer unsub3()

	if got := hub.SubscriberCount("profile-a"); got != 2 {
		t.Fatalf("expected cap to hold count at 2, got %d", got)
	}
}
