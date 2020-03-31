package peerstore

import (
	"strings"

	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/helpers"

	"github.com/coreos/go-semver/semver"
)

func MultistreamSemverMatcher(base protocol.ID) (func(string) bool, error){
	protos := strings.Split(string(base), "//")
	checkers := make([]func(string) bool, len(protos))

	for i, proto := range protos {
		splitted := strings.Split(proto, "/")
		_, err := semver.NewVersion(splitted[len(splitted) - 1])
		if err != nil {
			checkers[i] = DumbChecker(proto)
		} else {
			checkers[i], err = helpers.MultistreamSemverMatcher(protocol.ID(proto))
			if err != nil {
				return nil, err
			}
		}
	}

	return func(str string) bool {
		Protos := strings.Split(str, "//")

		if len(Protos) != len(checkers) {
			return false
		}

		for i, proto := range protos {
			if !checkers[i](proto){
				return false
			}
		}

		return true
	}, nil
}

func DumbChecker(base string) func(string) bool{
	return func(str string) bool {
		return base == str
	}
}
