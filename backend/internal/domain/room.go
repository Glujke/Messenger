package domain

import "errors"

var (
	ErrSelfChat      = errors.New("cannot start direct chat with yourself")
	ErrInvalidPeerID = errors.New("invalid peer id")
)

// ValidateDirectPeer checks that a direct chat target is valid.
func ValidateDirectPeer(callerID, peerID int64) error {
	if peerID <= 0 {
		return ErrInvalidPeerID
	}
	if callerID == peerID {
		return ErrSelfChat
	}
	return nil
}
