package util

import (
	"github.com/e-zhydzetski/ws-test/api"
	"testing"
)

func BenchmarkProto(b *testing.B) {
	var clientID *api.ClientID
	for i := 0; i < b.N; i++ {
		bytes, err := MarshalProtoMessage(&api.ClientID{
			Id: "123",
		})
		if err != nil {
			b.Fatal(err)
		}

		msg, err := UnmarshalProtoMessage(bytes)
		if err != nil {
			b.Fatal(err)
		}
		clientID = msg.(*api.ClientID)
	}
	_ = clientID
}
