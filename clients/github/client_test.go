package github

import (
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	ch := make(chan string)
	type args struct {
		announceChan chan string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "normal client init",
			args: struct{ announceChan chan string }{announceChan: ch},
			want: NewClient(ch),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.announceChan); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
