package email

// Sender sends a formatted email.
type Sender interface {
	Send(to, subject string, data ForwardData) error
}
