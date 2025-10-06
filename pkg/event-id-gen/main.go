package eventidgen

import (
	"fmt"
	"time"
)

func GenerateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
