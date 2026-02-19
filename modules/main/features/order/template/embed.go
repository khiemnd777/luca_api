package templateasset

import _ "embed"

// DeliveryNoteHTML is the HTML source for delivery note PDF generation.
//
//go:embed delivery_note.html
var DeliveryNoteHTML string
