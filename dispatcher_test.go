package main

import (
	"testing"
	"time"
)

func newTestKVstore() *KVstore {
	return &KVstore{
		kvmap: make(map[string]valueWithExpiry),
	}
}

func TestSetGet(t *testing.T) {
	store := newTestKVstore()

	setResult := store.runDispatcher([]string{"SET", "name", "Joshua"})
	if setResult != "+OK\r\n" {
		t.Errorf("expected %q, got %q", "+OK\r\n", setResult)
	}

	getResult := store.runDispatcher([]string{"GET", "name"})
	if getResult != "$6\r\nJoshua\r\n" {
		t.Errorf("expected %q, got %q", "$6\r\nJoshua\r\n", getResult)
	}
}

func TestGetNonexistentKey(t *testing.T) {
	store := newTestKVstore()

	result := store.runDispatcher([]string{"GET", "missing"})

	if result != "$-1\r\n" {
		t.Errorf("expected %q, got %q", "$-1\r\n", result)
	}
}

func TestDel(t *testing.T) {
	store := newTestKVstore()

	store.runDispatcher([]string{"SET", "name", "Joshua"})

	result := store.runDispatcher([]string{"DEL", "name"})

	if result != ":1\r\n" {
		t.Errorf("expected %q, got %q", ":1\r\n", result)
	}

	getResult := store.runDispatcher([]string{"GET", "name"})

	if getResult != "$-1\r\n" {
		t.Errorf("expected key to be deleted, got %q", getResult)
	}
}

func TestDelNonexistentKey(t *testing.T) {
	store := newTestKVstore()

	result := store.runDispatcher([]string{"DEL", "missing"})

	if result != ":0\r\n" {
		t.Errorf("expected %q, got %q", ":0\r\n", result)
	}
}

func TestWrongNumberOfArguments(t *testing.T) {
	store := newTestKVstore()

	tests := []struct {
		name     string
		commands []string
		expected string
	}{
		{
			name:     "SET too few arguments",
			commands: []string{"SET", "name"},
			expected: "-ERR wrong number of arguments for 'set' command\r\n",
		},
		{
			name:     "GET too few arguments",
			commands: []string{"GET"},
			expected: "-ERR wrong number of arguments for 'get' command\r\n",
		},
		{
			name:     "DEL too few arguments",
			commands: []string{"DEL"},
			expected: "-ERR wrong number of arguments for 'del' command\r\n",
		},
		{
			name:     "EXPIRE too few arguments",
			commands: []string{"EXPIRE", "name"},
			expected: "-ERR wrong number of arguments for 'expire' command\r\n",
		},
		{
			name:     "EXPIRE too many arguments",
			commands: []string{"EXPIRE", "name", "10", "extra"},
			expected: "-ERR wrong number of arguments for 'expire' command\r\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := store.runDispatcher(test.commands)

			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestPing(t *testing.T) {
	store := newTestKVstore()

	tests := []struct {
		name     string
		commands []string
		expected string
	}{
		{
			name:     "PING",
			commands: []string{"PING"},
			expected: "+PONG\r\n",
		},
		{
			name:     "PING with message",
			commands: []string{"PING", "hello"},
			expected: "+hello\r\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := store.runDispatcher(test.commands)

			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	store := newTestKVstore()

	result := store.runDispatcher([]string{"INVALID"})

	expected := "-ERR unknown command invalid\r\n"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpire(t *testing.T) {
	store := newTestKVstore()

	store.runDispatcher([]string{"SET", "name", "Joshua"})

	result := store.runDispatcher([]string{"EXPIRE", "name", "1"})

	if result != ":1\r\n" {
		t.Errorf("expected %q, got %q", ":1\r\n", result)
	}

	time.Sleep(1100 * time.Millisecond)

	getResult := store.runDispatcher([]string{"GET", "name"})

	if getResult != "$-1\r\n" {
		t.Errorf("expected key to expire, got %q", getResult)
	}
}

func TestExpireNonexistentKey(t *testing.T) {
	store := newTestKVstore()

	result := store.runDispatcher([]string{"EXPIRE", "missing", "10"})

	if result != ":0\r\n" {
		t.Errorf("expected %q, got %q", ":0\r\n", result)
	}
}

func TestExpireInvalidValue(t *testing.T) {
	store := newTestKVstore()

	store.runDispatcher([]string{"SET", "name", "Joshua"})

	result := store.runDispatcher([]string{"EXPIRE", "name", "abc"})

	expected := "-ERR value is not an integer or out of range\r\n"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}