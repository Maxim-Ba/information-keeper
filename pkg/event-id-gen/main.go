// eventidgen генерирует уникальный идентификатор события
package eventidgen

import (
	"fmt"
	"time"
)

// GenerateEventID генерирует уникальный идентификатор события
func GenerateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
