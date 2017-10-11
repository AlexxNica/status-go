package message

// Status defines a int type to indicate different status value of a
// message state.
type Status int

// consts of all message delivery status.
const (
	PendingStatus Status = iota
	QueuedStatus
	CachedStatus
	SentStatus
	ExpiredStatus
	ResendStatus
	FutureStatus
	LowPowStatus
	InvalidAESStatus
	OversizedMessageStatus
	OversizedVersionStatus
	RejectedStatus
	DeliveredStatus
)

// String returns the representation of giving state.
func (s Status) String() string {
	switch s {
	case PendingStatus:
		return "Pending"
	case QueuedStatus:
		return "Queued"
	case SentStatus:
		return "Sent"
	case RejectedStatus:
		return "Rejected"
	case DeliveredStatus:
		return "Delivered"
	case ExpiredStatus:
		return "ExpiredTTL"
	case ResendStatus:
		return "Resend"
	case FutureStatus:
		return "FutureDelivery"
	case LowPowStatus:
		return "LowPOWValue"
	case InvalidAESStatus:
		return "Invalid AES-GCM-Nonce"
	case OversizedMessageStatus:
		return "OversizedMessage"
	case OversizedVersionStatus:
		return "HigherWhisperVersion"
	}

	return "unknown"
}
