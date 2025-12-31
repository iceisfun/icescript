package object

// Critical represents a non-recoverable error that should halt the VM.
// Unlike Error, which can be returned as a value, Critical errors become runtime panics.
type Critical struct {
	Message string
}

func (c *Critical) Inspect() string  { return "CRITICAL: " + c.Message }
func (c *Critical) Type() ObjectType { return CRITICAL_OBJ }

func (c *Critical) AsFloat() (float64, bool) { return 0, false }
func (c *Critical) AsInt() (int64, bool)     { return 0, false }
func (c *Critical) AsString() (string, bool) { return "", false }
func (c *Critical) AsBool() (bool, bool)     { return false, false }
