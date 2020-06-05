package util

import (
	"fmt"
	"github.com/e-zhydzetski/ws-test/api"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

func MarshalProtoMessage(msg proto.Message) ([]byte, error) {
	body, err := ptypes.MarshalAny(msg)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&api.Msg{
		Body: body,
	})
}

func UnmarshalProtoMessage(bytes []byte) (proto.Message, error) {
	var msg api.Msg
	err := proto.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	if msg.GetBody() == nil {
		return nil, nil
	}
	typeURL := msg.Body.GetTypeUrl()
	switch typeURL {
	case "type.googleapis.com/ClientID":
		var body api.ClientID
		if err := ptypes.UnmarshalAny(msg.Body, &body); err != nil {
			return nil, err
		}
		return &body, nil
	case "type.googleapis.com/ServerPing":
		var body api.ServerPing
		if err := ptypes.UnmarshalAny(msg.Body, &body); err != nil {
			return nil, err
		}
		return &body, nil
	case "type.googleapis.com/ClientPong":
		var body api.ClientPong
		if err := ptypes.UnmarshalAny(msg.Body, &body); err != nil {
			return nil, err
		}
		return &body, nil
	default:
		return nil, fmt.Errorf("unknown message type: %s", typeURL)
	}
}
