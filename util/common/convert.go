package common

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func Int64ToUint32(i int64) uint32 {
	if i < -1 {
		return 0
	} else if i == -1 || i == 0 || i > 4294967295 {
		return 4294967295 // 2^32 - 1
	} else {
		return uint32(i)
	}
}

func MessageToAny(msg proto.Message) (*anypb.Any, error) {
	name, err := protoToMessageName(msg)
	if err != nil {
		return nil, err
	}
	buf, err := protoToMessageBytes(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		TypeUrl: name,
		Value:   buf,
	}, nil
}

func protoToMessageName(msg proto.Message) (string, error) {
	typeUrlPrefix := "type.googleapis.com/"

	if s := proto.MessageName(msg); s != "" {
		return fmt.Sprintf("%s%s", typeUrlPrefix, s), nil
	}
	return "", fmt.Errorf("can't determine message name")
}

func protoToMessageBytes(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func AnyToMessage(any *anypb.Any) (proto.Message, error) {
	return any.UnmarshalNew()
}
