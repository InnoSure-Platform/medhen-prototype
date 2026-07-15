package pdf

import "github.com/google/uuid"

func init() {
	newUUID = uuid.NewString
}
