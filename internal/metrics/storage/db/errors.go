package db

type ErrInvalidRecord struct {
	reason string
}

func NewErrInvalidRecord(reason string) *ErrInvalidRecord {
	return &ErrInvalidRecord{reason: reason}
}

func (e *ErrInvalidRecord) Error() string {
	return "invalid db record: " + e.reason
}
